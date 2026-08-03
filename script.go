package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/internal"
	"golang.org/x/net/html"
)

func NewScriptNodeTester() NodeTester {
	return script{}
}

type script struct{}

func (s script) Match(t testing.TB, node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "script"
}

func (s script) ListResourcesForPreloading(origin *url.URL, node *html.Node) (URLs []string) {
	for _, attr := range node.Attr {
		if attr.Key == "src" && strings.TrimSpace(attr.Val) != "" {
			URLs = append(URLs, joinRelativePath(origin, attr.Val))
		}
	}
	return URLs
}

func (s script) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	source := ""

	for _, attr := range node.Attr {
		switch attr.Key {
		case "src":
			if source != "" {
				t.Error("duplicate <script[src]> attribute:", source)
			}
			source = attr.Val
		case "language":
			t.Log(internal.WP, "<script[language]> attribute is deprecated:", attr.Val)
		}
	}

	if source != "" {
		script, contentType, err := loader.Load(
			t.Context(),
			joinRelativePath(origin, source),
		)
		if err != nil {
			if errors.Is(err, Skip) {
				return
			}
			t.Errorf("unable to load script %q: %v", source, err)
		}

		switch contentType {
		case "":
			t.Error("empty Content-Type for the script file")
		case "text/javascript", "application/javascript":
		default:
			t.Log("strange script Content-Type:", contentType)
		}

		if len(script) == 0 {
			t.Error("empty script file")
		}
	}
}
