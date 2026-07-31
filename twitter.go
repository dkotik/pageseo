package pageseo

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const (
	MetaTwitterPrefix      = "twitter:"
	MetaTwitterCard        = MetaTwitterPrefix + "card"
	MetaTwitterTitle       = MetaTwitterPrefix + "title"
	MetaTwitterDescription = MetaTwitterPrefix + "description"
	MetaTwitterSite        = MetaTwitterPrefix + "site"
	MetaTwitterURL         = MetaTwitterPrefix + "url"
	MetaTwitterImage       = MetaTwitterPrefix + "image"
)

type twitter struct {
	Type        string
	Title       string
	Description string
	Site        string
	URL         string
	Image       string
}

func (r PageValidator) testTwitterCard(meta headMeta) func(t *testing.T) {
	return func(t *testing.T) {
		card := twitter{}
		for name, content := range meta.Data {
			switch name {
			case MetaTwitterCard:
				card.Type = content
			case MetaTwitterTitle:
				card.Title = content
			case MetaTwitterDescription:
				card.Description = content
			case MetaTwitterSite:
				card.Site = content
			case MetaTwitterURL:
				card.URL = content
			case MetaTwitterImage:
				card.Image = content
			default:
				if strings.HasPrefix(name, MetaTwitterPrefix) {
					t.Log("unknown Twitter property:", name, content)
				}
			}
		}

		var err error
		if card.Type == "" {
			t.Error(MetaTwitterCard + " not found")
		}
		if card.Title == "" {
			t.Error(MetaTwitterTitle + " not found")
		} else if err = r.TwitterCardTitle.Validate(card.Title); err != nil {
			t.Error(MetaTwitterTitle+" validation failed:", err)
		}
		if card.Description == "" {
			t.Error(MetaTwitterDescription + " not found")
		} else if err = r.TwitterCardDescription.Validate(card.Description); err != nil {
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
		for property, _ := range meta.Data {
			if strings.HasPrefix(property, MetaOpenGraphPrefix) {
				t.Log("Twitter meta was set with \"property\" instead of the \"name\" attribute:", property)
			}
		}
	}
}

func (r PageValidator) TestTwitterCard(node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		head := findElementNode(node, "head")
		if head == nil {
			t.Fatal("document <head> node is absent")
		}
		r.testTwitterCard(getHeadMeta(t, head))(t)
	}
}
