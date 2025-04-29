package pageseo

import (
	"errors"
	"fmt"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

var (
	ErrMissingSrcAttribute = errors.New("img missing src attribute")
	ErrMissingAltAttribute = errors.New("img missing alt attribute")
)

func ValidateImage(node *html.Node) error {
	attributes, err := htmltest.ParseAttributes(node)
	if err != nil {
		return err
	}

	if src, ok := attributes["src"]; !ok || src == "" {
		return ErrMissingSrcAttribute
	}

	alt, ok := attributes["alt"]
	if !ok || alt == "" {
		return ErrMissingAltAttribute
	} else if len(alt) > DefaultMaximumTitleLength {
		return fmt.Errorf("image alternative text is too long: %d vs %d maximum", len(alt), DefaultMaximumTitleLength)
	}
	return nil
}
