package pageseo

import (
	"errors"
	"fmt"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

func ValidateLink(node *html.Node) error {
	if node.FirstChild == nil || node.FirstChild.Data == "" {
		return errors.New("link text content is empty")
	}
	textContent := htmltest.ParseTextContent(node)
	if len(textContent) > DefaultMaximumTitleLength {
		return fmt.Errorf("link text content is too long: %d vs %d maximum", len(textContent), DefaultMaximumTitleLength)
	}
	return nil
}
