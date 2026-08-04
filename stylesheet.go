package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func NewStyleSheetNodeTester() NodeTester {
	return styleSheet{}
}

type styleSheet struct{}

func (s styleSheet) Match(t testing.TB, node *html.Node) bool {
	if node.Type != html.ElementNode || node.Data != "style" {
		for _, attr := range node.Attr {
			return attr.Key == "rel" && attr.Val == "stylesheet"
		}
	}
	return false
}

func (s styleSheet) ListResourcesForPreloading(origin *url.URL, node *html.Node) (URLs []string) {
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			if strings.TrimSpace(attr.Val) != "" {
				URLs = append(URLs, joinRelativePath(origin, attr.Val))
			}
			break // only take the first attribute
		}
	}
	return URLs
}

func (s styleSheet) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	href := ""
	for _, attr := range node.Attr {
		switch attr.Key {
		case "href":
			if href != "" {
				t.Error("duplicate <link[href]> attribute:", href)
			} else {
				href = attr.Val
			}
		}
	}

	if href == "" {
		t.Error("missing <link[href]> attribute")
		return
	}
	styleSheet, contentType, err := loader.Load(
		t.Context(),
		joinRelativePath(origin, href),
	)
	if err != nil {
		if errors.Is(err, Skip) {
			return
		}
		t.Errorf("unable to load style sheet %q: %v", href, err)
	}

	switch contentType {
	case "":
		t.Error("empty Content-Type for the style sheet file")
	case "text/css":
	default:
		t.Log("strange style sheet Content-Type:", contentType)
	}

	if len(styleSheet) == 0 {
		t.Error("empty style sheet file")
	}
}
