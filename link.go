package pageseo

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

const (
	// DefaultMinimumLinkTextLength sets the minimum length of the anchor text.
	// A pagination link is often a single character.
	DefaultMinimumLinkTextLength = 1
	DefaultMaximumLinkTextLength = DefaultMaximumTitleLength * 6
)

func NewLinkTextValidator(s StringConstraints) htmltest.Validator {
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

type linkTextValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
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

func (r PageValidator) TestLink(origin string, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		logAttributes(t, node.Attr)
		isEmpty := true
		for descendant := range node.Descendants() {
			switch descendant.Type {
			case html.TextNode:
				// if err := r.LinkText.Validate(htmltest.ParseTextContent(node)); err != nil {
				// 	t.Errorf("invalid anchor text: %v", err)
				// }
				isEmpty = false
			case html.ElementNode:
				switch strings.ToLower(descendant.Data) {
				case "a", "svg": // contains an image
				}
				isEmpty = false
			}
		}
		if isEmpty {
			t.Error("anchor is empty of meaningful content")
		}

		attributes, err := htmltest.ParseAttributes(node)
		if err != nil {
			t.Errorf("failed to parse attributes: %v", err)
			return
		}

		if target, ok := attributes["target"]; ok {
			if strings.ToLower(strings.TrimSpace(target)) == "_blank" {
				rel, ok := attributes["rel"]
				if !ok {
					t.Errorf("anchor text with target=\"_blank\" should have a rel attribute")
				} else if slices.Index(strings.Fields(rel), "noopener") == -1 {
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
			href, err := htmltest.JoinURL(origin, href)
			if err != nil {
				t.Errorf("failed to join path: %v", err)
				return
			}
			if err := r.URL.Validate(href); err != nil {
				t.Fatalf("dead URL: %v", err)
			}
		} else {
			t.Log("anchor text without href")
		}
	}
}
