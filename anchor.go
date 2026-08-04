package pageseo

import (
	"errors"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/internal"
	"golang.org/x/net/html"
)

type anchor struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
	Cache         *cachedParsedURLs
}

func NewAnchorNodeTester(s StringConstraints) NodeTester {
	if s.Normalizer == nil {
		s.Normalizer = NormalizeLineToNFC
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumAnchorTextLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumAnchorTextLength
	}
	return anchor{
		Normalizer:    s.Normalizer,
		MinimumLength: s.MinimumLength,
		MaximumLength: s.MaximumLength,
		Cache:         &cachedParsedURLs{},
	}
}

func (a anchor) Match(t testing.TB, node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "a"
}

func (a anchor) ListResourcesForPreloading(origin *url.URL, node *html.Node) (URLs []string) {
	for _, attr := range node.Attr {
		if attr.Key == "href" && strings.TrimSpace(attr.Val) != "" {
			url, _, _ := strings.Cut(attr.Val, "#")
			if url != "" {
				URLs = append(URLs, joinRelativePath(origin, url))
			}
		}
	}
	return URLs
}

func (a anchor) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	attributes := internal.GetAttributes(t, node)
	rel := []string{}
	relString, ok := attributes["rel"]
	if ok {
		for _, field := range strings.Fields(relString) {
			if slices.Index(rel, field) == -1 {
				rel = append(rel, field)
			} else {
				t.Log(internal.WP, "duplicate <a[rel]> field:", field)
			}
		}
	}

	target, ok := attributes["target"]
	if ok {
		if strings.ToLower(target) == "_blank" {
			if slices.Index(rel, "noopener") == -1 {
				t.Error("add rel=\"noopener\" attribute to prevent tab nabbing")
				t.Log("older versions of Firefox require rel=\"noopener noreferrer\"")
			}
		}
	}

	title, ok := attributes["title"]
	if !ok {
		t.Log(internal.WP, "<a[title]> attribute is empty")
	} else {
		normalized, err := a.Normalizer.Normalize(title)
		if err != nil {
			t.Logf(internal.WP+" unable to normalize <a[title]>: %v", err)
		} else if normalized != title {
			t.Log(internal.WP + " <a[title]> is not normalized")
		}

		length := len(title)
		if length < a.MinimumLength {
			t.Log("<a[title]> is too short")
		} else if length > a.MaximumLength {
			t.Log("<a[title]> is too long")
		}
	}

	href, ok := attributes["href"]
	if !ok {
		if _, ok = attributes["onclick"]; !ok {
			if _, ok = attributes["id"]; !ok { // <a id="#hash" />
				t.Error("add <a[href]> link attribute")
			}
		}
	} else if href, _, _ = strings.Cut(href, "#"); href != "" {
		url, err := a.Cache.Get(href)
		if err != nil {
			t.Logf("%s failed to parse location: %v", internal.WP, err)
		} else {
			if IsExternalLocation(origin, url) {
				if !IsSubdomainOfOrigin(origin, url) {
					if slices.Index(rel, "external") == -1 {
						t.Log(internal.WP, "add \"external\" directive to [rel] attribute")
					}
					if slices.Index(rel, "nofollow") == -1 {
						t.Log(internal.WP, "add \"nofollow\" directive to [rel] attribute")
					}
				}
			} else {
				url = joinInternalPath(origin, url)
			}
			href = url.String()
		}

		target, contentType, err := loader.Load(t.Context(), href)
		if err != nil {
			if errors.Is(err, Skip) {
				return
			}
			t.Errorf("unable to load anchor %q: %v", href, err)
		}

		switch contentType {
		case "":
			t.Error("empty Content-Type for the link <a[href]> target")
		case "text/html":
		case "text/markdown":
		case "application/pdf":
		case "text/plain":
		case "image/jpeg":
		case "image/png":
		case "image/webp":
		case "image/gif":
		case "image/svg+xml":
		case "image/avif":
		case "audio/mpeg":
		case "video/mp4":
		default:
			t.Log("strange link <a[href]> target Content-Type:", contentType)
		}
		if len(target) == 0 {
			t.Error("empty <a[href]> target file")
		}
	}

	isEmpty := true
	for descendant := range node.Descendants() {
		switch descendant.Type {
		case html.TextNode:
			isEmpty = false
		case html.ElementNode:
			switch descendant.Data {
			case "a", "svg", "button": // contains an image
			}
			isEmpty = false
		}
	}

	if isEmpty {
		t.Error("anchor is empty of meaningful content")
	} else {
		text := internal.GetText(node)
		normalized, err := a.Normalizer.Normalize(text)
		if err != nil {
			t.Logf(internal.WP+" unable to normalize anchor text: %v", err)
		} else if normalized != text {
			t.Log(internal.WP + " anchor text is not normalized")
		}
		length := len(text)
		if length < a.MinimumLength && href != "" {
			t.Error("anchor text is too short")
		} else if length > a.MaximumLength {
			t.Error("anchor text is too long")
		}
	}
}
