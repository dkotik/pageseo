package pageseo

import (
	"errors"
	"io"
	"iter"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/text/language"
)

const warningPrefix = "<WARNING>"

//go:generate go run ./testdata/generate.go

type StringConstraints struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

// NodeTester creates HTML node tests.
//
// NodeTester should not issue any warnings or errors
// if a [Loader] returns a [Skip] sentinel error.
type NodeTester interface {
	// Match returns true if the current node should be tested.
	Match(testing.TB, *html.Node) bool

	// ListResourcesForPreloading returns a list of resource
	// locations that should be preloaded for the matched node.
	ListResourcesForPreloading(
		originPage *url.URL,
		matchedNode *html.Node,
	) []string

	// TestNode enforces search engine optimization standards
	// by testing the matched node and its descendents.
	//
	// The resource loader is populated with resources provided by
	// [NodeTester.ListResourcesForPreloading] in advance.
	TestNode(
		t testing.TB,
		originPage *url.URL,
		matchedNode *html.Node,
		resourceLoader Loader,
	)
}

type PageTester interface {
	TestPage(origin *url.URL, node *html.Node) func(t *testing.T)
}

type pageSEO struct {
	loader      Loader
	nodeTesters []NodeTester
}

func NewPageTester(
	loader Loader,
	nodeTesters ...NodeTester,
) PageTester {
	if loader == nil {
		loader = skipAllLoadingSingleton
	}
	for _, tester := range nodeTesters {
		if tester == nil {
			panic("node tester is nil")
		}
	}
	if len(nodeTesters) == 0 {
		nodeTesters = DefaultNodeTests()
	}
	return pageSEO{
		loader:      loader,
		nodeTesters: nodeTesters,
	}
}

type nodeTests struct {
	Node  *html.Node
	Tests []NodeTester
}

func (p pageSEO) walkNodeTree(
	t *testing.T,
	origin *url.URL,
	node *html.Node,
	preload func([]string),
) iter.Seq[nodeTests] {
	return func(yield func(nodeTests) bool) {
		// the root
		next := make([]NodeTester, 0, len(p.nodeTesters))
		for _, nt := range p.nodeTesters {
			if nt.Match(t, node) {
				next = append(next, nt)
				preload(nt.ListResourcesForPreloading(
					origin, node,
				))
			}
		}
		if len(next) > 0 {
			if !yield(nodeTests{
				Node:  node,
				Tests: slices.Clone(next),
			}) {
				return
			}
			next = next[:0]
		}

		// the children of the root
		for child := range node.ChildNodes() {
			for _, nt := range p.nodeTesters {
				if nt.Match(t, child) {
					next = append(next, nt)
					preload(nt.ListResourcesForPreloading(
						origin, child,
					))
				}
			}
			if len(next) > 0 {
				if !yield(nodeTests{
					Node:  child,
					Tests: slices.Clone(next),
				}) {
					return
				}
				next = next[:0]
			}
			for nodeTests := range p.walkNodeTree(t, origin, child, preload) {
				if !yield(nodeTests) {
					return
				}
			}
		}
	}
}

func (p pageSEO) TestPage(origin *url.URL, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		if origin == nil {
			t.Fatal("origin is nil")
		}
		if node == nil || node.FirstChild == nil {
			t.Fatal("HTML document is empty")
		}
		if node.FirstChild.Type != html.DoctypeNode {
			t.Error("HTML document root is not a <!DOCTYPE html>")
		} else if node.FirstChild.Data != "html" {
			t.Error("document type is not a <!DOCTYPE html>")
		}

		htmlTag := getFirstElementOrSibling(node.FirstChild)
		if htmlTag == nil {
			t.Error("HTML document has no root <html> node")
		} else if htmlTag.Type != html.ElementNode || htmlTag.Data != "html" {
			t.Error("HTML document root is not an <html> node")
		} else {
			foundLanguageAttribute := 0
			for _, attr := range htmlTag.Attr {
				if attr.Key == "lang" {
					foundLanguageAttribute++
					if _, err := language.Parse(attr.Val); err != nil {
						t.Errorf("<html> BCP47 [lang] attribute %q is not canonical: %v", attr.Val, err)
					}
				}
			}
			switch foundLanguageAttribute {
			case 1: // ok
			case 0:
				t.Error("<html> tag has no [lang] attribute")
			default:
				t.Errorf("<html> tag has %d extra [lang] attributes", foundLanguageAttribute-1)
			}

			head := getFirstElementOrSibling(htmlTag.FirstChild)
			if head == nil || head.Data != "head" {
				t.Fatal("<html> tag has no <head> node")
			} else {
				body := getFirstElementOrSibling(head.NextSibling)
				if body == nil || body.Data != "body" {
					t.Fatal("<html> tag has no <body> node")
				} else {
					// standard library parser ignores trailing nodes,
					// but let's try it anyway, for completeness:
					// are there any extra trailing body nodes?
					body = body.NextSibling
					for {
						if body == nil {
							break
						}
						switch body.Type {
						case html.CommentNode: // ok
						case html.ElementNode:
							t.Error("HTML document has an extra trailing <" + body.Data + "> node")
						case html.TextNode:
							t.Error("HTML document has an extra trailing text node")
						default:
							t.Error("HTML document has an unexpected extra trailing node")
						}
						body = body.NextSibling
					}
				}
			}
		}

		testsToRun := make([]nodeTests, 0, 8)
		reploadURLs := make([]string, 0, 8)

		for nodeTests := range p.walkNodeTree(
			t, origin, node,
			func(URLs []string) {
				if len(URLs) == 0 {
					return
				}
				reploadURLs = slices.Concat(reploadURLs, URLs)
			},
		) {
			testsToRun = append(testsToRun, nodeTests)
		}

		var hotSwap Loader
		if testing.Short() {
			t.Log("[SKIP] short tests do not load any page resources")
			hotSwap = skipAllLoadingSingleton
		} else if len(reploadURLs) == 0 {
			hotSwap = p.loader
		} else {
			hotSwap = NewHotSwap(t.Context(), p.loader, reploadURLs)
		}
		for _, nodeTest := range testsToRun {
			t.Run(
				getTestName(nodeTest.Node),
				func(t *testing.T) {
					writeElementPath(t.Output(), nodeTest.Node)
					logAttributes(t, nodeTest.Node.Attr)
					for _, test := range nodeTest.Tests {
						test.TestNode(t, origin, nodeTest.Node, hotSwap)
					}
				},
			)
		}

		// standard library parser ignores trailing nodes,
		// but let's try it anyway, for completeness:
		// are there any extra trailing root nodes?
		if htmlTag != nil {
			htmlTag = htmlTag.NextSibling
			for {
				if htmlTag == nil {
					break
				}
				switch htmlTag.Type {
				case html.CommentNode: // ok
				case html.ElementNode:
					t.Error("HTML document has an extra trailing <" + htmlTag.Data + "> node")
				case html.TextNode:
					t.Error("HTML document has an extra trailing text node")
				default:
					t.Error("HTML document has an unexpected extra trailing node")
				}
				htmlTag = htmlTag.NextSibling
			}
		}
	}
}

type Requirements struct {
	// Normalizer is passed to all default validator constructors.
	// If you are using custom validators, you should pass your
	// own normalizer to each constructor manually.
	//
	// Default value is [PassthroughNormalizer] that does not do anything.
	Normalizer Normalizer

	DeduplicationNamespace               string
	TitleDeduplicator                    ValidationMiddleware
	DescriptionDeduplicator              ValidationMiddleware
	OpenGraphCardTitleDeduplicator       ValidationMiddleware
	OpenGraphCardDescriptionDeduplicator ValidationMiddleware
	TwitterCardTitleDeduplicator         ValidationMiddleware
	TwitterCardDescriptionDeduplicator   ValidationMiddleware

	Description Validator
	Heading     Validator
	Language    Validator

	URL          Validator
	LinkText     Validator
	ImageAltText Validator
	ImageSrc     Validator
}

type PageValidator struct {
	Loader                   Loader
	Description              Validator
	OpenGraphCardDescription Validator
	TwitterCardDescription   Validator
	Heading                  Validator
	Language                 Validator

	URL      Validator
	LinkText Validator
	ImageSrc Validator

	cachedURLs *cachedParsedURLs
}

func New(loader Loader, r Requirements) PageValidator {
	if loader == nil {
		panic("nil loader")
	}
	if r.Normalizer == nil {
		r.Normalizer = PassthroughNormalizer
	}

	if r.TitleDeduplicator == nil {
		r.TitleDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.DescriptionDeduplicator == nil {
		r.DescriptionDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.OpenGraphCardTitleDeduplicator == nil {
		r.OpenGraphCardTitleDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.OpenGraphCardDescriptionDeduplicator == nil {
		r.OpenGraphCardDescriptionDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.TwitterCardTitleDeduplicator == nil {
		r.TwitterCardTitleDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.TwitterCardDescriptionDeduplicator == nil {
		r.TwitterCardDescriptionDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}

	if r.Description == nil {
		r.Description = NewDescriptionValidator(StringConstraints{Normalizer: r.Normalizer})
	}

	if r.Language == nil {
		r.Language = ValidatorFunc(func(s string) error {
			if !regexp.MustCompile(`^\w\w(\-\w\w)?$`).MatchString(s) {
				return errors.New("invalid language code")
			}
			return nil
		})
	}
	if r.URL == nil {
		r.URL = NewURLValidator(StringConstraints{})
	}
	if r.LinkText == nil {
		r.LinkText = NewLinkTextValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.ImageSrc == nil {
		r.ImageSrc = NewURLValidator(StringConstraints{
			MaximumLength: DefaultMaximumImageSourceLength,
		})
	}

	return PageValidator{
		Loader:                   loader,
		Description:              r.DescriptionDeduplicator.Wrap(r.Description),
		OpenGraphCardDescription: r.OpenGraphCardDescriptionDeduplicator.Wrap(r.Description),
		TwitterCardDescription:   r.TwitterCardDescriptionDeduplicator.Wrap(r.Description),
		Language:                 r.Language,

		URL:      r.URL,
		LinkText: r.LinkText,
		ImageSrc: r.ImageSrc,

		cachedURLs: &cachedParsedURLs{},
	}
}

func NewStrict(loader Loader, r Requirements) PageValidator {
	if r.Normalizer == nil {
		r.Normalizer = NormalizeTextToNFC
	}

	if r.LinkText == nil {
		r.LinkText = NewLinkTextValidator(StringConstraints{Normalizer: NormalizeLineToNFC})
	}
	return New(loader, r)
}

func (r PageValidator) Test(origin string, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		var ok bool
		originURL, err := r.cachedURLs.Get(origin)
		if err != nil {
			t.Fatal("invalid URL:", err)
		}
		if node.FirstChild == nil {
			t.Fatal("page contains no HTML nodes")
		}
		err = ValidateDoctypeTag(node.FirstChild)
		if err != nil {
			t.Errorf("page has an invalid <DOCTYPE> tag: %v", err)
		}
		TestDocumentRootHasExactlyDoctypeAndHTMLNodes(node)(t)
		attributes := getAttributes(t, node.FirstChild.NextSibling)
		language, ok := attributes["lang"]
		if !ok {
			t.Error("HTML tag is missing a lang attribute")
		}
		if err = r.Language.Validate(language); err != nil {
			t.Errorf("HTML tag has an invalid lang attribute %q: %v", language, err)
		}

		nextChild, closer := iter.Pull[*html.Node](node.FirstChild.NextSibling.ChildNodes())
		defer closer()

		for {
			child, ok := nextChild()
			if !ok {
				t.Fatal("HTML tag is missing a <BODY> tag")
			}
			if child.Type != html.ElementNode {
				continue
			}
			if child.Data != "body" {
				t.Fatalf("second child element tag is not a <BODY> tag: %s", child.Data)
			}

			hotSwap := r.preloadResources(t, originURL, child)
			for node := range child.Descendants() {
				switch node.Data {
				case "a":
					t.Run(getElementPath(node), r.testLink(originURL, node, hotSwap))
				}
			}

			break // found a body tag
		}
	}
}

func (v PageValidator) TestReader(
	origin string,
	r io.Reader,
) func(t *testing.T) {
	return func(t *testing.T) {
		tree, err := html.Parse(r)
		if err != nil {
			t.Fatalf("unable to parse the HTML page: %v", err)
		}
		if tree == nil {
			t.Fatal("no HTML tree found in the reader")
		}
		v.Test(origin, tree)(t)
	}
}

func (v PageValidator) TestFile(p string) func(t *testing.T) {
	return func(t *testing.T) {
		f, err := os.Open(p)
		if err != nil {
			t.Fatalf("unable to open file %q: %v", p, err)
		}
		t.Cleanup(func() {
			if cerr := f.Close(); cerr != nil {
				t.Errorf("unable to close HTML file %q: %v", p, cerr)
			}
		})
		v.TestReader("file://"+p, f)(t)
	}
}

type matchCounter struct {
	NodeTester
	ValidateCount func(testing.TB, uint32)
}

func (mc matchCounter) Match(t testing.TB, node *html.Node) (matched bool) {
	if node.Type == html.DocumentNode {
		t.Cleanup(func() {
			// at the end of the test walk the tree
			// and validate the count
			var matchedCount uint32
			if mc.NodeTester.Match(t, node) {
				matchedCount++
			}
			for descendant := range node.Descendants() {
				if mc.NodeTester.Match(t, descendant) {
					matchedCount++
				}
			}
			mc.ValidateCount(t, matchedCount)
		})
	}
	return mc.NodeTester.Match(t, node)
}

// MustMatch fails the test with a message if the
// NodeTester does not match any nodes during page validation.
func MustMatch(nt NodeTester, message string) NodeTester {
	if nt == nil {
		panic("nil node tester")
	}
	if message == "" {
		panic("empty message")
	}
	// message = strconv.Quote(message)
	mc := matchCounter{
		NodeTester: nt,
		ValidateCount: func(t testing.TB, matchedCount uint32) {
			if matchedCount == 0 {
				t.Error(message)
			}
		},
	}
	return mc
}

// MustMatchExactly fails the test with a message if the
// NodeTester does not match the exact number of nodes during
// page validation.
func MustMatchExactly(t *testing.T, nt NodeTester, message string, timesMatched uint32) NodeTester {
	if nt == nil {
		panic("nil node tester")
	}
	if message == "" {
		panic("empty message")
	}
	// message = strconv.Quote(message)
	mc := matchCounter{
		NodeTester: nt,
		ValidateCount: func(t testing.TB, matchedCount uint32) {
			if matchedCount != timesMatched {
				t.Logf(
					"node tester matched %d times instead of expected %d", matchedCount, timesMatched,
				)
				t.Error(message)
			}
		},
	}
	return mc
}

// MustMatchAtLeast fails the test with a message if the
// NodeTester does not match at least the specified number
// of nodes during page validation.
func MustMatchAtLeast(t *testing.T, nt NodeTester, message string, timesMatched uint32) NodeTester {
	if nt == nil {
		panic("nil node tester")
	}
	if message == "" {
		panic("empty message")
	}
	// message = strconv.Quote(message)
	mc := matchCounter{
		NodeTester: nt,
		ValidateCount: func(t testing.TB, matchedCount uint32) {
			if matchedCount < timesMatched {
				t.Logf(
					"node tester matched %d times instead of expected %d", matchedCount, timesMatched,
				)
				t.Error(message)
			}
		},
	}
	return mc
}

// MustMatchAtMost fails the test with a message if the
// NodeTester does not match at most the specified number
// of nodes during page validation.
func MustMatchAtMost(t *testing.T, nt NodeTester, message string, timesMatched uint32) NodeTester {
	if nt == nil {
		panic("nil node tester")
	}
	if message == "" {
		panic("empty message")
	}
	mc := matchCounter{
		NodeTester: nt,
		ValidateCount: func(t testing.TB, matchedCount uint32) {
			if matchedCount > timesMatched {
				t.Logf(
					"node tester matched %d times instead of expected %d", matchedCount, timesMatched,
				)
				t.Error(message)
			}
		},
	}
	return mc
}

type elementTester struct {
	Data   string
	Tester func(testing.TB, *html.Node)
}

// NewNodeElementTester is a helper function that creates a
// simplified [NodeTester] for an HTML element type.
// It will never load resources.
//
// Wrap it with [MustMatchExactly] to build fluent page
// validators.
func NewNodeElementTester(name string, tester func(testing.TB, *html.Node)) NodeTester {
	if name == "" {
		panic("empty element name")
	}
	if tester == nil {
		panic("nil element tester")
	}
	return elementTester{
		Data:   strings.ToLower(name),
		Tester: tester,
	}
}

func (e elementTester) Match(t testing.TB, possible *html.Node) bool {
	if possible.Type != html.ElementNode {
		return false
	}
	return possible.Data != e.Data
}

func (e elementTester) ListResourcesForPreloading(
	originPage *url.URL,
	matchedNode *html.Node,
) []string {
	return nil
}

func (e elementTester) TestNode(
	t testing.TB,
	originPage *url.URL,
	matchedNode *html.Node,
	resourceLoader Loader,
) {
	e.Tester(t, matchedNode)
}
