package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const (
	// DefaultMinimumLinkTextLength sets the minimum length of the anchor text.
	// A pagination link is often a single character.
	DefaultMinimuLinkLength         = 1
	DefaultMaximumLinkLength        = 120
	DefaultMaximumImageSourceLength = 2048 // older browser constraint
	DefaultMinimumLinkTextLength    = 1
	DefaultMaximumLinkTextLength    = DefaultMaximumTitleLength * 6
)

type linkTextValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func NewLinkTextValidator(s StringConstraints) Validator {
	if s.Normalizer == nil {
		s.Normalizer = PassthroughNormalizer
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumLinkTextLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumLinkTextLength
	}
	return linkTextValidator(s)
}

func (s linkTextValidator) Validate(value string) error {
	normalized, err := s.Normalizer.Normalize(value)
	if err != nil {
		return err
	}
	if normalized != value {
		return errors.New("anchor text is not UTF normalized")
	}

	switch length := len(normalized); {
	case length == 0:
		return errors.New("anchor text is empty")
	case length < s.MinimumLength:
		return errors.New("anchor text is too short")
	case length > s.MaximumLength:
		return errors.New("anchor text is too long")
	default:
		return nil
	}
}

func (r PageValidator) testLink(
	origin *url.URL,
	node *html.Node,
	loader Loader,
) func(t *testing.T) {
	return func(t *testing.T) {
		logAttributes(t, node.Attr)
		isEmpty := true
		for descendant := range node.Descendants() {
			switch descendant.Type {
			case html.TextNode:
				// if err := r.LinkText.Validate(ParseTextContent(node)); err != nil {
				// 	t.Errorf("invalid anchor text: %v", err)
				// }
				isEmpty = false
			case html.ElementNode:
				switch descendant.Data {
				case "a", "svg": // contains an image
				}
				isEmpty = false
			}
		}
		if isEmpty {
			t.Error("anchor is empty of meaningful content")
		}

		attributes := getAttributes(t, node)
		rel := make(map[string]struct{})
		if relString, ok := attributes["rel"]; ok {
			for _, directive := range strings.Fields(relString) {
				_, ok = rel[directive]
				if ok {
					t.Errorf("duplicatel a[rel] directive: %s", directive)
				} else {
					rel[directive] = struct{}{}
				}
			}
		}

		if target, ok := attributes["target"]; ok {
			if strings.ToLower(strings.TrimSpace(target)) == "_blank" {
				if _, ok = rel["noopener"]; !ok {
					t.Errorf("anchor text with target=\"_blank\" should have a rel=\"noopener\" setting to prevent tab nabbing; if you need to support older versions of Firefox, use rel=\"noopener noreferrer\"")
				}
			}
		}

		if href, ok := attributes["href"]; ok && len(href) > 0 {

			href, _, ok := strings.Cut(href, "#")
			if ok {
				if strings.TrimSpace(href) == "" {
					return // just a #hash reference
				}
			}
			err := r.URL.Validate(href)
			if err != nil {
				t.Errorf("deformed URL: %v", err)
			}
			url, err := r.cachedURLs.Get(href)
			if err != nil {
				t.Fatalf("failed to parse location: %v", err)
			}
			// TODO: is external should be pulled out of resource
			// path merging should be taken care of when
			// resource is loading
			if IsExternalLocation(origin, url) {
				if _, ok = rel["external"]; !ok {
					t.Error("add \"external\" directive to [rel] attribute")
				}
				if _, ok = rel["nofollow"]; !ok {
					t.Error("add \"nofollow\" directive to [rel] attribute")
				}
			} else {
				url = joinInternalPath(origin, url)
			}
			// _, _, err = r.Loader.Load(t.Context(), href)
			// if err != nil {
			// 	t.Fatalf("unable to load URL <%s>: %v", err, href)
			// }
		} else {
			t.Log("anchor text without href")
		}
	}
}
