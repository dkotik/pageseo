package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func NewImageNodeTester(s StringConstraints) NodeTester {
	if s.Normalizer == nil {
		s.Normalizer = NormalizeLineToNFC
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumImageAltTextLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumImageAltTextLength
	}
	return image{
		Normalizer:    s.Normalizer,
		MinimumLength: s.MinimumLength,
		MaximumLength: s.MaximumLength,
	}
}

type image struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func (i image) Match(t testing.TB, node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "img"
}

func (i image) ListResourcesForPreloading(origin *url.URL, node *html.Node) (URLs []string) {
	for _, attr := range node.Attr {
		if attr.Key != "src" || strings.HasPrefix(attr.Val, "data:") {
			continue
		}
		URLs = append(URLs, joinRelativePath(origin, attr.Val))
	}
	for _, src := range GetPictureSourceList(node) {
		if strings.HasPrefix(src, "data:") {
			continue
		}
		URLs = append(URLs, joinRelativePath(origin, src))
	}
	return URLs
}

func (i image) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	source, alt, title := "", "", ""

	for _, attr := range node.Attr {
		switch attr.Key {
		case "src":
			if source != "" {
				t.Error("duplicate <img[src]> attribute:", source)
			}
			source = attr.Val
		case "alt":
			if alt != "" {
				t.Error("duplicate <img[alt]> attribute:", alt)
			}
			alt = attr.Val
		case "title":
			if title != "" {
				t.Log("duplicate <img[title]> attribute:", title)
			}
			title = attr.Val
		}
	}

	normalized, err := i.Normalizer.Normalize(alt)
	if err != nil {
		t.Logf(warningPrefix+" unable to normalize <img[alt]> text: %v", err)
	} else if normalized != alt {
		t.Log(warningPrefix + " <img[alt]> is not normalized")
	}
	length := len(alt)
	if length < i.MinimumLength {
		t.Error("<img[alt]> is too short")
	} else if length > i.MaximumLength {
		t.Error("<img[alt]> is too long")
	}

	switch title {
	case "":
	default:
		switch length = len(title); {
		case length < i.MinimumLength:
			t.Log("<img[title]> is too short")
		case length > i.MaximumLength:
			t.Log("<img[title]> is too long")
		}
	}

	validateImage(t, origin, source, loader)
	for _, source = range GetPictureSourceList(node) {
		validateImage(t, origin, source, loader)
	}
}

func GetPictureSourceList(node *html.Node) (result []string) {
	if node.Parent.Type != html.ElementNode || node.Parent.Data != "picture" {
		return
	}
	for node := range node.ChildNodes() {
		if node.Type != html.ElementNode || node.Data != "source" {
			continue
		}
		for _, attr := range node.Attr {
			if attr.Key != "srcset" || attr.Val == "" {
				continue
			}
			for _, src := range strings.Split(attr.Val, ",") {
				src = strings.TrimSpace(src)
				if src == "" {
					continue
				}
				src, _, _ = strings.Cut(src, ";") // just the first part
				result = append(result, src)
			}
		}
	}
	return result
}

func validateImage(
	t testing.TB,
	origin *url.URL,
	URL string,
	loader Loader,
) {
	if URL == "" {
		t.Error("empty <image> source")
		return
	}
	if strings.HasPrefix(URL, "data:") {
		// TODO: decode the base64 data and fallthrough
		return // skip embedded image
	}

	image, contentType, err := loader.Load(
		t.Context(),
		joinRelativePath(origin, URL),
	)
	if err != nil {
		if errors.Is(err, Skip) {
			return
		}
		t.Errorf("unable to load image %q: %v", URL, err)
	}

	switch contentType {
	case "":
		t.Error("empty Content-Type for the image file")
	case "image/jpeg":
	case "image/png":
	case "image/webp":
	case "image/gif":
	case "image/svg+xml":
	case "image/avif":
	default:
		t.Log("strange image Content-Type:", contentType)
	}

	if len(image) == 0 {
		t.Error("empty image data:", contentType)
	}
}
