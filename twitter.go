package pageseo

import (
	"errors"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

const (
	MetaTwitterCard        = "twitter:card"
	MetaTwitterTitle       = "twitter:title"
	MetaTwitterDescription = "twitter:description"
	MetaTwitterSite        = "twitter:site"
	MetaTwitterURL         = "twitter:url"
	MetaTwitterImage       = "twitter:image"
)

type twitter struct {
	Type        string
	Title       string
	Description string
	Site        string
	URL         string
	Image       string
}

func loadTwitterCard(node *html.Node) (result twitter, err error) {
	if node.Type != html.ElementNode || strings.ToLower(node.Data) != "head" {
		for descendants := range node.Descendants() {
			result, err = loadTwitterCard(descendants)
			if err == nil {
				return
			}
		}
		return result, errors.New("HTML head node not found")
	}
	for meta := range node.ChildNodes() {
		if meta.Type != html.ElementNode || strings.ToLower(meta.Data) != "meta" {
			continue
		}
		attributes, err := htmltest.ParseAttributes(meta)
		if err != nil {
			return result, err
		}
		name, ok := attributes["name"]
		if !ok || name == "" {
			continue
		}
		content, ok := attributes["content"]
		if !ok || content == "" {
			continue
		}
		switch strings.ToLower(name) {
		case MetaTwitterCard:
			result.Type = content
		case MetaTwitterTitle:
			result.Title = content
		case MetaTwitterDescription:
			result.Description = content
		case MetaTwitterSite:
			result.Site = content
		case MetaTwitterURL:
			result.URL = content
		case MetaTwitterImage:
			result.Image = content
		}
	}
	return
}

func (r Requirements) TestTwitterCard(node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		card, err := loadTwitterCard(node)
		if err != nil {
			t.Fatal("unable to load Twitter card")
		}
		if card.Type == "" {
			t.Error(MetaTwitterCard + " not found")
		}
		if card.Title == "" {
			t.Error(MetaTwitterTitle + " not found")
		} else if err = r.Title.Validate(card.Title); err != nil {
			t.Error(MetaTwitterTitle+" validation failed:", err)
		}
		if card.Description == "" {
			t.Error(MetaTwitterDescription + " not found")
		} else if err = r.Description.Validate(card.Description); err != nil {
			t.Error(MetaTwitterDescription+" validation failed:", err)
		}
		// if card.URL == "" {
		// 	t.Error(MetaTwitterCard+" not found")
		// }
		if card.Site == "" {
			t.Error(MetaTwitterSite + " not found")
		}
		if card.Image == "" {
			t.Error(MetaTwitterImage + " not found")
		}
	}
}
