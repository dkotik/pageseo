package pageseo

import (
	"net/url"
	"testing"

	"github.com/dkotik/pageseo/internal"
	"golang.org/x/net/html"
)

func NewHeadingNodeTester(s StringConstraints) NodeTester {
	if s.Normalizer == nil {
		s.Normalizer = NormalizeLineToNFC
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumHeadingLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumHeadingLength
	}
	return heading(s)
}

type heading struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func (h heading) Match(t testing.TB, node *html.Node) bool {
	switch node.Type {
	case html.ElementNode:
		switch node.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			return true
		default:
			return false
		}
	case html.DocumentNode:
		t.Cleanup(func() {
			var countOfTopHeadings uint32
			for child := range node.Descendants() {
				if child.Type == html.ElementNode && child.Data == "h1" {
					countOfTopHeadings++
				}
			}
			switch countOfTopHeadings {
			case 0:
				t.Error("document has no <h1> headings")
			case 1: // as required
			default:
				t.Logf(internal.WP+" document has %d extra <h1> headings", countOfTopHeadings-1)
			}
		})
		return false
	default:
		return false
	}
}

func (h heading) ListResourcesForPreloading(*url.URL, *html.Node) []string {
	return nil
}

func (h heading) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	textContent := internal.GetAndTrimText(node)
	normalized, err := h.Normalizer.Normalize(textContent)
	if err != nil {
		t.Logf(internal.WP+" unable to normalize heading text: %v", err)
	} else if normalized != textContent {
		t.Log(internal.WP + " heading text is not normalized")
	}

	switch length := len(normalized); {
	case length == 0:
		t.Error("heading is empty")
	case length < h.MinimumLength:
		if node.Data == "h1" {
			t.Error("top heading text content is too short:", length, "vs", h.MinimumLength, "characters")
		} else {
			t.Log(internal.WP, "text content is too short:", length, "vs", h.MinimumLength, "characters")
		}
	case length > h.MaximumLength:
		if node.Data == "h1" {
			t.Error("top heading text content is too long:", length, "vs", h.MaximumLength, "characters")
		} else {
			t.Log(internal.WP, "text content is too long:", length, "vs", h.MaximumLength, "characters")
		}
	}
}

func NewTableNodeTester() NodeTester {
	return table{}
}

type table struct{}

func (s table) Match(t testing.TB, node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "table"
}

func (s table) ListResourcesForPreloading(origin *url.URL, node *html.Node) []string {
	return nil
}

func (s table) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	caption := ""
	for child := range node.ChildNodes() {
		if child.Type == html.ElementNode && child.Data == "caption" {
			if caption != "" {
				t.Error("multiple <caption> elements in the table")
			}
			caption = internal.GetAndTrimText(child)
		}
	}

	if caption == "" {
		t.Error("add a <caption> element to the table")
	}
}
