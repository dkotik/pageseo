package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const (
	DefaultMinimumImageAltTextLength = 0
	DefaultMaximumImageAltTextLength = DefaultMaximumTitleLength * 12
)

func NewImageAltTextValidator(s StringConstraints) Validator {
	if s.Normalizer == nil {
		s.Normalizer = PassthroughNormalizer
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

func GetPictureSourceList(node *html.Node) (result []string) {
	if node.Parent.Type != html.ElementNode || node.Parent.Data != "picture" {
		return
	}
	for node := range node.ChildNodes() {
		if node.Type != html.ElementNode || node.Data != "source" {
			continue
		}
		for _, attr := range node.Attr {
			if strings.ToLower(attr.Key) != "srcset" || attr.Val == "" {
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

func (r PageValidator) TestImage(origin *url.URL, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		logAttributes(t, node.Attr)
		attributes := getAttributes(t, node)
		if src, ok := attributes["src"]; ok {
			if !strings.HasPrefix(src, "data:") {
				// TODO:? r.ImageSrc
				if err := r.URL.Validate(src); err != nil {
					t.Log("Src:", src)
					t.Errorf("invalid image source: %v", err)
				}
				_, err := url.Parse(src)
				if err != nil {
					t.Error("invalid URL:", err)
				}
			}
		} else {
			srcSet := GetPictureSourceList(node.Parent)
			if len(srcSet) == 0 {
				t.Fatal("no srcSet tag attribute")
			}
			for _, src := range srcSet {
				// TODO:? r.ImageSrc
				if err := r.URL.Validate(src); err != nil {
					t.Log("Src:", src)
					t.Errorf("invalid image source: %v", err)
				}
				_, err := url.Parse(src)
				if err != nil {
					t.Error("invalid URL:", err)
				}
			}
		}

		if r.ImageAltText == SkipValidator {
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
