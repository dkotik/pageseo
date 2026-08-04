package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/internal"
	"golang.org/x/net/html"
)

func NewLinkNodeTester() NodeTester {
	return link{}
}

//	 <link rel="alternate"
//		href="https://google.com"
//		hreflang="en-gb" />
type link struct{}

func (s link) Match(t testing.TB, node *html.Node) bool {
	if node.Type != html.ElementNode || node.Data != "link" {
		for _, attr := range node.Attr {
			return attr.Key == "rel" && attr.Val == "alternate"
		}
	}
	return false
}

func (s link) ListResourcesForPreloading(origin *url.URL, node *html.Node) (URLs []string) {
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

func (s link) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	href := ""
	for _, attr := range node.Attr {
		switch attr.Key {
		case "href":
			if href != "" {
				t.Error("duplicate <link[href]> attribute:", href)
			} else {
				href = attr.Val
			}
		case "hreflang":
			internal.ValidateLanguage(t, attr.Val)
		}
	}

	if href == "" {
		t.Error("missing <link[href]> attribute")
		return
	}
	link, contentType, err := loader.Load(
		t.Context(),
		joinRelativePath(origin, href),
	)
	if err != nil {
		if errors.Is(err, Skip) {
			return
		}
		t.Errorf("unable to load link %q: %v", href, err)
	}

	switch contentType {
	case "":
		t.Error("empty Content-Type for the link file")
	case "text/html":
	default:
		t.Log("strange link Content-Type:", contentType)
	}

	if len(link) == 0 {
		t.Error("empty link file")
	}
}
