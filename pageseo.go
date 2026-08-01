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
	"sync/atomic"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/text/language"
)

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
	Match(*html.Node) bool

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
		originPage *url.URL,
		matchedNode *html.Node,
		resourceLoader Loader,
	) func(*testing.T)
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
		panic("loader is nil")
	}
	for _, tester := range nodeTesters {
		if tester == nil {
			panic("node tester is nil")
		}
	}
	if len(nodeTesters) == 0 {
		nodeTesters = GetDefaultNodeTests()
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

func (p pageSEO) eachChild(
	t *testing.T,
	origin *url.URL,
	node *html.Node,
	preload func([]string),
) iter.Seq[nodeTests] {
	return func(yield func(nodeTests) bool) {
		next := make([]NodeTester, 0, len(p.nodeTesters))
		for child := range node.ChildNodes() {
			// might want to test text nodes for something
			// if child.Type != html.ElementNode {
			// 	continue
			// }
			for _, nt := range p.nodeTesters {
				if nt.Match(child) {
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
			for nodeTests := range p.eachChild(t, origin, child, preload) {
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
				if strings.ToLower(attr.Key) == "lang" {
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

		for nodeTests := range p.eachChild(
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
			hotSwap = skipAllLoaderSingleton
		} else if len(reploadURLs) == 0 {
			hotSwap = p.loader
		} else {
			hotSwap = NewHotSwap(t.Context(), p.loader, reploadURLs)
		}
		for _, pair := range testsToRun {
			t.Run(
				getElementPath(pair.Node),
				func(t *testing.T) {
					for _, test := range pair.Tests {
						test.TestNode(origin, pair.Node, hotSwap)(t)
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

	Title       Validator
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
	Title                    Validator
	Description              Validator
	OpenGraphCardTitle       Validator
	OpenGraphCardDescription Validator
	TwitterCardTitle         Validator
	TwitterCardDescription   Validator
	Heading                  Validator
	Language                 Validator

	URL          Validator
	LinkText     Validator
	ImageAltText Validator
	ImageSrc     Validator

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

	if r.Title == nil {
		r.Title = NewTitleValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.Description == nil {
		r.Description = NewDescriptionValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.Heading == nil {
		r.Heading = NewHeadingValidator(StringConstraints{Normalizer: r.Normalizer})
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
	if r.ImageAltText == nil {
		r.ImageAltText = NewImageAltTextValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.ImageSrc == nil {
		r.ImageSrc = NewURLValidator(StringConstraints{
			MaximumLength: DefaultMaximumImageSourceLength,
		})
	}

	return PageValidator{
		Loader:                   loader,
		Title:                    r.TitleDeduplicator.Wrap(r.Title),
		Description:              r.DescriptionDeduplicator.Wrap(r.Description),
		OpenGraphCardTitle:       r.OpenGraphCardTitleDeduplicator.Wrap(r.Title),
		OpenGraphCardDescription: r.OpenGraphCardDescriptionDeduplicator.Wrap(r.Description),
		TwitterCardTitle:         r.TwitterCardTitleDeduplicator.Wrap(r.Title),
		TwitterCardDescription:   r.TwitterCardDescriptionDeduplicator.Wrap(r.Description),
		Heading:                  r.Heading,
		Language:                 r.Language,

		URL:          r.URL,
		LinkText:     r.LinkText,
		ImageAltText: r.ImageAltText,
		ImageSrc:     r.ImageSrc,

		cachedURLs: &cachedParsedURLs{},
	}
}

func NewStrict(loader Loader, r Requirements) PageValidator {
	if r.Normalizer == nil {
		r.Normalizer = NormalizeTextToNFC
	}
	if r.Title == nil {
		r.Title = NewTitleValidator(StringConstraints{Normalizer: NormalizeLineToNFC})
	}
	if r.Heading == nil {
		r.Heading = NewHeadingValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.LinkText == nil {
		r.LinkText = NewLinkTextValidator(StringConstraints{Normalizer: NormalizeLineToNFC})
	}
	if r.ImageAltText == nil {
		r.ImageAltText = NewImageAltTextValidator(StringConstraints{Normalizer: NormalizeLineToNFC})
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
		var child *html.Node

		for {
			child, ok = nextChild()
			if !ok {
				t.Error("HTML tag is missing a <HEAD> tag at the top")
				break
			}
			if child.Type != html.ElementNode {
				continue
			}
			if child.Data != "head" {
				t.Errorf("first child element tag is not a <HEAD> tag: %s", child.Data)
				break
			}
			t.Run(getElementPath(child), r.TestHead(child))
			break // found a head tag
		}

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
			t.Run(getElementPath(child), r.TestHeadings(child))

			hotSwap := r.preloadResources(t, originURL, child)
			for node := range child.Descendants() {
				switch node.Data {
				case "a":
					t.Run(getElementPath(node), r.testLink(originURL, node, hotSwap))
				case "img":
					t.Run(getElementPath(node), r.testImage(originURL, node, hotSwap))
				}
			}

			break // found a body tag
		}

		child, ok = nextChild()
		if ok {
			t.Errorf("HTML tag contains more than two children: %s", child.Data)
		}

		// for node := range node.Descendants() {
		// 	if node.Type != html.ElementNode {
		// 		continue
		// 	}
		// 	switch node.Data {
		// 	case "a":
		// 		// if r.LinkText == SkipValidator {
		// 		// 	continue
		// 		// }
		// 		t.Run(Path(node), r.TestLink(originURL, node))
		// 	case "img":
		// 		// if (r.ImageAltText == nil || r.ImageAltText == SkipValidator) && (r.ImageSrc == nil || r.ImageSrc == SkipValidator) {
		// 		// 	continue
		// 		// }
		// 		t.Run(Path(node), r.TestImage(origin, node))
		// 		// if err = ValidateImage(node); err != nil {
		// 		// 	t.Errorf("invalid link tag %q: %v", Path(node), err)
		// 		// }
		// 		// case "script":
		// 		// 	t.Run("script tag has valid attributes", r.TestScript(node))
		// 		// case "style":
		// 		// 	t.Run("style tag has valid attributes", r.TestStyle(node))
		// 	}
		// }
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
	Count *atomic.Uint32
}

func (mc matchCounter) Match(node *html.Node) (matched bool) {
	matched = mc.NodeTester.Match(node)
	if matched {
		mc.Count.Add(1)
	}
	return matched
}

// MustMatch fails the test with a message if the
// NodeTester does not match any nodes during page validation.
func MustMatch(t *testing.T, nt NodeTester, message string) NodeTester {
	mc := matchCounter{
		NodeTester: nt,
		Count:      new(atomic.Uint32),
	}
	if message == "" {
		message = "required node tester did not match any nodes"
	}
	t.Cleanup(func() {
		if mc.Count.Load() == 0 {
			t.Error(message)
		}
	})
	return mc
}

// MustMatchExactly fails the test with a message if the
// NodeTester does not match the exact number of nodes during
// page validation.
func MustMatchExactly(t *testing.T, nt NodeTester, message string, timesMatched uint32) NodeTester {
	mc := matchCounter{
		NodeTester: nt,
		Count:      new(atomic.Uint32),
	}
	if message == "" {
		message = "required node tester did not match exactly"
	}
	t.Cleanup(func() {
		actuallyMatchedTimes := mc.Count.Load()
		if actuallyMatchedTimes != timesMatched {
			t.Logf(
				"node tester matched %d times instead of %d", actuallyMatchedTimes, timesMatched,
			)
			t.Error(message)
		}
	})
	return mc
}

// MustMatchAtLeast fails the test with a message if the
// NodeTester does not match at least the specified number
// of nodes during page validation.
func MustMatchAtLeast(t *testing.T, nt NodeTester, message string, timesMatched uint32) NodeTester {
	mc := matchCounter{
		NodeTester: nt,
		Count:      new(atomic.Uint32),
	}
	if message == "" {
		message = "required node tester did not match enough times"
	}
	t.Cleanup(func() {
		actuallyMatchedTimes := mc.Count.Load()
		if actuallyMatchedTimes < timesMatched {
			t.Logf(
				"node tester matched %d times instead of %d", actuallyMatchedTimes, timesMatched,
			)
			t.Error(message)
		}
	})
	return mc
}

type elementTester struct {
	Data   string
	Tester func(*html.Node) func(*testing.T)
}

// NewNodeElementTester is a helper function that creates a
// simplified [NodeTester] for an HTML element type.
// It will never load resources.
//
// Wrap it with [MustMatchExactly] to build fluent page
// validators.
func NewNodeElementTester(name string, tester func(*html.Node) func(*testing.T)) NodeTester {
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

func (e elementTester) Match(
	possible *html.Node,
) bool {
	if possible.Type != html.ElementNode {
		return false
	}
	return possible.Data != strings.ToLower(e.Data)
}

func (e elementTester) ListResourcesForPreloading(
	originPage *url.URL,
	matchedNode *html.Node,
) []string {
	return nil
}

func (e elementTester) TestNode(
	originPage *url.URL,
	matchedNode *html.Node,
	resourceLoader Loader,
) func(*testing.T) {
	return e.Tester(matchedNode)
}
