package pageseo

import (
	"net/url"
	"testing"

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
		case "h1", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9", "h10":
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
				t.Error("document has no H1 headings")
			default:
				t.Logf(warningPrefix+" document has %d extra H1 headings", countOfTopHeadings-1)
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
	textContent := GetText(node)
	normalized, err := h.Normalizer.Normalize(textContent)
	if err != nil {
		t.Error(warningPrefix, "normalization failed:", err)
	} else if normalized != textContent {
		t.Error(warningPrefix, "text content is not normalized")
	}

	switch length := len(normalized); {
	case length == 0:
		t.Error("heading is empty")
	case length < h.MinimumLength:
		if node.Data == "h1" {
			t.Error("top heading text content is too short:", length, "vs", h.MinimumLength, "characters")
		} else {
			t.Log(warningPrefix, "text content is too short:", length, "vs", h.MinimumLength, "characters")
		}
	case length > h.MaximumLength:
		if node.Data == "h1" {
			t.Error("top heading text content is too long:", length, "vs", h.MaximumLength, "characters")
		} else {
			t.Log(warningPrefix, "text content is too long:", length, "vs", h.MaximumLength, "characters")
		}
	}
}
