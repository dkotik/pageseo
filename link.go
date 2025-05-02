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
	DefaultMinimumLinkTextLength = 4
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
		if err := r.LinkText.Validate(htmltest.ParseTextContent(node)); err != nil {
			for descendant := range node.Descendants() {
				if descendant.Type != html.ElementNode {
					continue
				}
				switch strings.ToLower(descendant.Data) {
				case "a", "svg":
					return // contains an image
				}
			}
			t.Errorf("invalid anchor text: %v", err)
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

		if href, ok := attributes["href"]; ok {
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
