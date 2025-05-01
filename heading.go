package pageseo

import (
	"errors"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

const (
	DefaultMinimumHeadingLength = 4
	DefaultMaximumHeadingLength = 55
)

func NewHeadingValidator(s StringConstraints) htmltest.Validator {
	if s.Normalizer == nil {
		s.Normalizer = NormalizeLineToNFC
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumHeadingLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumHeadingLength
	}
	return headingValidator(s)
}

type headingValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func (s headingValidator) Validate(value string) error {
	normalized, err := s.Normalizer.Normalize(value)
	if err != nil {
		return err
	}
	if normalized != value {
		return errors.New("page heading is not normalized")
	}

	switch length := len(normalized); {
	case length == 0:
		return errors.New("page heading is empty")
	case length < s.MinimumLength:
		return errors.New("page heading is too short")
	case length > s.MaximumLength:
		return errors.New("page heading is too long")
	default:
		return nil
	}
}

func (r PageValidator) TestHeadings(node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		if r.Heading == htmltest.SkipValidator || r.Heading == nil {
			t.Skip("heading validation is skipped by user request")
		}

		foundValidH1 := false
		t.Cleanup(func() {
			if !foundValidH1 {
				t.Errorf("H1 tag not found under %q tag", node.Data)
			}
		})

		var err error
		for descendant := range node.Descendants() {
			if descendant.Type != html.ElementNode {
				continue
			}
			switch descendant.Data {
			case "h1":
				text := htmltest.ParseTextContent(descendant)
				if text == "" {
					t.Errorf("H1 tag has no text content")
					continue
				}
				if err = r.Title.Validate(text); err != nil {
					t.Errorf("H1 tag text content is not valid: %v", err)
					continue
				}
				foundValidH1 = true
			case "h2":
				text := htmltest.ParseTextContent(descendant)
				if text == "" {
					t.Errorf("heading tag %q has no text content", descendant.Data)
					continue
				}
				if err = r.Title.Validate(text); err != nil {
					t.Errorf("heading tag %q text content is not valid: %v", descendant.Data, err)
					continue
				}
			case "h3", "h4", "h5", "h6":
				// TODO: examine only in strict mode
			}
		}
	}
}
