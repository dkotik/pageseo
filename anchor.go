package pageseo

import (
	"errors"
	"net/url"
	"slices"
	"strings"
	"testing"

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
	href, title, rel, target := "", "", []string{}, ""

	for _, attr := range node.Attr {
		switch attr.Key {
		case "href":
			if href != "" {
				t.Error("duplicate <a[href]> attribute:", href)
			} else if attr.Val == "" {
				t.Log(warningPrefix, "<a[href]> attribute is empty")
			}
			href = attr.Val
		case "rel":
			if len(rel) > 0 {
				t.Error("duplicate <a[rel]> attribute:", rel)
			} else if attr.Val == "" {
				t.Log(warningPrefix, "<a[rel]> attribute is empty")
			}
			for _, field := range strings.Fields(attr.Val) {
				if slices.Index(rel, field) == -1 {
					rel = append(rel, field)
				} else {
					t.Log(warningPrefix, "duplicate <a[rel]> field:", field)
				}
			}
		case "target":
			if target != "" {
				t.Error("duplicate <a[target]> attribute:", target)
			} else if target == "" {
				t.Log(warningPrefix, "<a[target]> attribute is empty")
			}
			target = attr.Val
		case "title":
			if title != "" {
				t.Log("duplicate <a[title]> attribute:", title)
			} else if title == "" {
				t.Log(warningPrefix, "<a[title]> attribute is empty")
			}
			title = attr.Val
		}
	}

	if href == "" {
		t.Error("add <a[href]> link attribute")
	} else if href, _, _ = strings.Cut(href, "#"); href != "" {
		url, err := a.Cache.Get(href)
		if err != nil {
			t.Logf("%s failed to parse location: %v", warningPrefix, err)
		} else {
			if IsExternalLocation(origin, url) {
				if slices.Index(rel, "external") == -1 {
					t.Log("add \"external\" directive to [rel] attribute")
				}
				if slices.Index(rel, "nofollow") == -1 {
					t.Log(warningPrefix, "add \"nofollow\" directive to [rel] attribute")
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

	if title != "" {
		normalized, err := a.Normalizer.Normalize(title)
		if err != nil {
			t.Logf(warningPrefix+" unable to normalize <a[title]>: %v", err)
		} else if normalized != title {
			t.Log(warningPrefix + " <a[title]> is not normalized")
		}

		length := len(title)
		if length < a.MinimumLength {
			t.Log("<a[title]> is too short")
		} else if length > a.MaximumLength {
			t.Log("<a[title]> is too long")
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
		text := GetText(node)
		normalized, err := a.Normalizer.Normalize(text)
		if err != nil {
			t.Logf(warningPrefix+" unable to normalize anchor text: %v", err)
		} else if normalized != text {
			t.Log(warningPrefix + " anchor text is not normalized")
		}
		length := len(text)
		if length < a.MinimumLength {
			t.Error("anchor text is too short")
		} else if length > a.MaximumLength {
			t.Error("anchor text is too long")
		}
	}

	if strings.ToLower(target) == "_blank" {
		if slices.Index(rel, "noopener") == -1 {
			t.Error("add rel=\"noopener\" attribute to prevent tab nabbing")
			t.Log("older versions of Firefox require rel=\"noopener noreferrer\"")
		}
	}
}
