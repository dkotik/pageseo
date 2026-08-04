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
		if attr.Key != "src" {
			continue
		}
		if strings.HasPrefix(attr.Val, "data:") {
			URLs = append(URLs, joinRelativePath(origin, attr.Val))
		}
		break // only take the first attribute
	}
	for _, src := range GetPictureSourceList(node) {
		if strings.HasPrefix(src, "data:") {
			continue
		}
		URLs = append(URLs, joinRelativePath(origin, src))
	}
	return URLs
}

// isImageAvailable checks for a common scenario where images
// with a missing alt attribute are marked as unavailable
// via aria-label="image unavailable".
func isImageAvailable(attributes map[string]string) bool {
	ariaLabel, ok := attributes["aria-label"]
	if !ok {
		return true
	}
	fields := strings.Fields(ariaLabel)
	if slices.Index(fields, "image") == -1 {
		return true
	}
	return slices.Index(fields, "unavailable") == -1
}

func (i image) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	attributes := internal.GetAttributes(t, node)
	ok := false

	alt, ok := attributes["alt"]
	if !ok {
		if isImageAvailable(attributes) {
			t.Error("missing <img[alt]> attribute")
		}
	} else {
		normalized, err := i.Normalizer.Normalize(alt)
		if err != nil {
			t.Logf(internal.WP+" unable to normalize <img[alt]> text: %v", err)
		} else if normalized != alt {
			t.Log(internal.WP + " <img[alt]> is not normalized")
		}
		length := len(alt)
		if length < i.MinimumLength {
			t.Errorf("<img[alt]> is %d characters, expected %d or more", length, i.MinimumLength)
		} else if length > i.MaximumLength {
			t.Errorf("<img[alt]> is %d characters, expected %d or less", length, i.MaximumLength)
		} else if length > 80 {
			t.Logf("<img[alt]> is %d characters, screen readers prefer 80 or less", length)
		}
	}

	title, ok := attributes["title"]
	if ok {
		switch length := len(title); {
		case length < i.MinimumLength:
			t.Log("<img[title]> is too short")
		case length > i.MaximumLength:
			t.Log("<img[title]> is too long")
		}
	}

	src, ok := attributes["src"]
	if !ok || src == "" {
		t.Log("If you are loading images lazily with JavaScript, stop,")
		t.Log("and use modern loading=\"lazy\" attribute instead.")
		t.Error("empty <image[src]> source")
	} else {
		validateImage(t, origin, src, loader)
	}
	for _, src = range GetPictureSourceList(node) {
		if src == "" {
			t.Error("empty <picture[srcset]> source")
			continue
		}
		validateImage(t, origin, src, loader)
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
