package pageseo

import (
	"errors"
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
		s.Normalizer = NormalizeLine
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

func (r Requirements) TestLink(node *html.Node) func(t *testing.T) {
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
	}
}
