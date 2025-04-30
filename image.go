package pageseo

import (
	"errors"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

const (
	DefaultMinimumImageAltTextLength = 0
	DefaultMaximumImageAltTextLength = DefaultMaximumTitleLength * 12
)

func NewImageAltTextValidator(s StringConstraints) htmltest.Validator {
	if s.Normalizer == nil {
		s.Normalizer = NormalizeLineToNFC
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumImageAltTextLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumImageAltTextLength
	}
	return ImageAltTextValidator(s)
}

type ImageAltTextValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func (s ImageAltTextValidator) Validate(value string) error {
	normalized, err := s.Normalizer.Normalize(value)
	if err != nil {
		return err
	}
	if normalized != value {
		return errors.New("anchor text is not UTF normalized")
	}

	switch length := len(normalized); {
	case length < s.MinimumLength:
		return errors.New("anchor text is too short")
	case length > s.MaximumLength:
		return errors.New("anchor text is too long")
	default:
		return nil
	}
}

func GetPictureSourceList(node *html.Node) (result []string, err error) {
	if node.Parent.Type != html.ElementNode || node.Parent.Data != "picture" {
		return
	}
	for node := range node.ChildNodes() {
		if node.Type != html.ElementNode || node.Data != "source" {
			continue
		}
		attributes, err := htmltest.ParseAttributes(node)
		if err != nil {
			break
		}
		srcSet, ok := attributes["srcSet"]
		if !ok || srcSet == "" {
			return result, errors.New("picture source definition source set is missing")
		}
		for _, src := range strings.Split(srcSet, ",") {
			if src == "" {
				continue
			}
			src, _, _ = strings.Cut(src, ";") // just the first part
			result = append(result, src)
		}
	}
	return result, err
}

func (r Requirements) TestImage(node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		attributes, err := htmltest.ParseAttributes(node)
		if err != nil {
			t.Fatal("unable to extract image attributes:", err)
		}
		if src, ok := attributes["src"]; !ok || src == "" {
			srcSet, err := GetPictureSourceList(node.Parent)
			if err != nil {
				t.Fatal("unable to extract picture source list:", err)
			}
			if len(srcSet) == 0 {
				t.Fatal("missing src attribute")
			}
		}

		if r.ImageAltText == nil || r.ImageAltText == htmltest.SkipValidator {
			return
		}
		alt, ok := attributes["alt"]
		if !ok {
			t.Fatal("missing alt attribute")
		} else if err := r.ImageAltText.Validate(alt); err != nil {
			t.Log("Alt:", alt)
			t.Errorf("invalid alternative text: %v", err)
		}
	}
}
