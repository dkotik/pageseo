package pageseo

import (
	"io"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/internal"
	"golang.org/x/net/html"
)

//go:generate go run ./testdata/generate.go

type StringConstraints struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func (s StringConstraints) apply(t testing.TB, subject, text string) {
	t.Helper()
	normalized, err := s.Normalizer.Normalize(text)
	if err != nil {
		t.Logf("%s %s text cannot be normalized: %v", internal.WP, subject, err)
	} else if subject != normalized {
		t.Logf("%s %s text is not normalized", internal.WP, subject)
		subject = normalized
	}

	if s.MinimumLength > 0 && len(subject) < s.MinimumLength {
		t.Errorf("%s text is too short: got %d, want at least %d", subject, len(subject), s.MinimumLength)
	}
	if s.MaximumLength > 0 && len(subject) > s.MaximumLength {
		t.Errorf("%s text is too long: got %d, want at most %d", subject, len(subject), s.MaximumLength)
	}
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
	TestPage(string, io.Reader) func(t *testing.T)
	TestFile(string) func(t *testing.T)
}

type pageSEO struct {
	loader      Loader
	nodeTesters []NodeTester
}

func New(
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

func (p pageSEO) TestPage(URL string, r io.Reader) func(t *testing.T) {
	return func(t *testing.T) {
		origin, err := url.Parse(URL)
		if err != nil {
			t.Fatalf("unable to parse URL for file %q: %v", URL, err)
		}
		tree, err := html.Parse(r)
		if err != nil {
			t.Fatalf("unable to parse HTML file %q: %v", URL, err)
		}
		p.TestTree(origin, tree)(t)
	}
}

func (p pageSEO) TestFile(path string) func(t *testing.T) {
	return func(t *testing.T) {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("unable to open file %q: %v", path, err)
		}
		tree, err := html.Parse(f)
		if err != nil {
			t.Fatalf("unable to parse HTML file %q: %v", path, err)
		}
		if cerr := f.Close(); cerr != nil {
			t.Errorf("unable to close HTML file %q: %v", path, cerr)
		}
		p.TestTree(&url.URL{Scheme: "file", Path: path}, tree)(t)
	}
}

type nodeTests struct {
	Node  *html.Node
	Tests []NodeTester
}

func validateDocumentTypeElement(t *testing.T, node *html.Node) {
	if node == nil || node.FirstChild == nil {
		t.Fatal("HTML document is empty")
	}
	if node.FirstChild.Type != html.DoctypeNode {
		t.Error("HTML document root is not a <!DOCTYPE html>")
	} else if node.FirstChild.Data != "html" {
		t.Error("document type is not a <!DOCTYPE html>:", node.FirstChild.Data)
	}
}

func validateHTMLElement(t *testing.T, node *html.Node) {
	if node == nil {
		t.Fatal("HTML document has no root <html> node")
	}
	if node.Type != html.ElementNode || node.Data != "html" {
		t.Fatal("HTML document root is not an <html> node")
	}
	foundLanguageAttribute := 0
	for _, attr := range node.Attr {
		if attr.Key == "lang" {
			foundLanguageAttribute++
			internal.ValidateLanguage(t, attr.Val)
		}
	}
	switch foundLanguageAttribute {
	case 1: // ok
	case 0:
		t.Error("<html> tag has no [lang] attribute")
	default:
		t.Errorf("<html> tag has %d extra [lang] attributes", foundLanguageAttribute-1)
	}
}

func (p pageSEO) TestTree(origin *url.URL, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		if origin == nil {
			t.Fatal("origin is nil")
		}
		validateDocumentTypeElement(t, node)
		htmlTag := internal.GetFirstElementOrSibling(node.FirstChild)
		validateHTMLElement(t, htmlTag)

		head := internal.GetFirstElementOrSibling(htmlTag.FirstChild)
		if head == nil || head.Data != "head" {
			t.Fatal("<html> tag has no <head> node")
		} else {
			body := internal.GetFirstElementOrSibling(head.NextSibling)
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

		foundNav, foundHeader, foundFooter := false, false, false
		t.Cleanup(func() {
			if foundNav == false {
				t.Log("add a <nav> element to the page")
			}
			if foundHeader == false {
				t.Log("add a <header> element to the page")
			}
			if foundFooter == false {
				t.Log("add a <footer> element to the page")
			}
		})
		testsToRun := make([]nodeTests, 0, 8)
		reploadURLs := make([]string, 0, 8)
		packTests := func(node *html.Node, nts []NodeTester) {
			next := make([]NodeTester, 0, len(nts)/4)
			for _, nt := range nts {
				if nt.Match(t, node) {
					next = append(next, nt)
					reploadURLs = append(reploadURLs, nt.ListResourcesForPreloading(
						origin, node,
					)...)
				} else if node.Type == html.ElementNode {
					switch node.Data {
					case "nav":
						foundNav = true
					case "header":
						foundHeader = true
					case "footer":
						foundFooter = true
					}
				}
			}
			if len(next) == 0 {
				return
			}
			testsToRun = append(testsToRun, nodeTests{
				Node:  node,
				Tests: next,
			})
		}

		packTests(node, p.nodeTesters)
		for child := range node.Descendants() {
			packTests(child, p.nodeTesters)
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
				internal.GetTestName(nodeTest.Node),
				func(t *testing.T) {
					internal.WriteElementPath(t.Output(), nodeTest.Node)
					internal.LogAttributes(t, nodeTest.Node.Attr)
					for _, test := range nodeTest.Tests {
						test.TestNode(t, origin, nodeTest.Node, hotSwap)
					}
				},
			)
			// fmt.Print("=================")
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
