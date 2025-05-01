package pageseo

import (
	"errors"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

const (
	MetaOpenGraphType        = "og:type"
	MetaOpenGraphTitle       = "og:title"
	MetaOpenGraphDescription = "og:description"
	MetaOpenGraphURL         = "og:url"
	MetaOpenGraphImage       = "og:image"
)

type openGraph struct {
	Type        string
	Title       string
	Description string
	Site        string
	URL         string
	Image       string
}

func loadOpenGraphCard(node *html.Node) (result openGraph, err error) {
	if node.Type != html.ElementNode || strings.ToLower(node.Data) != "head" {
		for descendants := range node.Descendants() {
			result, err = loadOpenGraphCard(descendants)
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
		name, ok := attributes["property"]
		if !ok || name == "" {
			continue
		}
		content, ok := attributes["content"]
		if !ok || content == "" {
			continue
		}
		switch strings.ToLower(name) {
		case MetaOpenGraphType:
			result.Type = content
		case MetaOpenGraphTitle:
			result.Title = content
		case MetaOpenGraphDescription:
			result.Description = content
		case MetaOpenGraphURL:
			result.URL = content
		case MetaOpenGraphImage:
			result.Image = content
		}
	}
	return
}

func (r PageValidator) TestOpenGraphCard(node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		card, err := loadOpenGraphCard(node)
		if err != nil {
			t.Fatal("unable to load openGraph card")
		}
		if card.Type == "" {
			t.Error(MetaOpenGraphType + " not found")
		}
		if card.Title == "" {
			t.Error(MetaOpenGraphTitle + " not found")
		} else if err = r.Title.Validate(card.Title); err != nil {
			t.Error(MetaOpenGraphTitle+" validation failed:", err)
		}
		if card.Description == "" {
			t.Error(MetaOpenGraphDescription + " not found")
		} else if err = r.Description.Validate(card.Description); err != nil {
			t.Error(MetaOpenGraphDescription+" validation failed:", err)
		}
		// if card.URL == "" {
		// 	t.Error(MetaOpenGraphURL + " not found")
		// }
		if card.Image == "" {
			t.Error(MetaOpenGraphImage + " not found")
		}
	}
}
