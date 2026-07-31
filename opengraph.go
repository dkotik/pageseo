package pageseo

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const (
	MetaOpenGraphPrefix      = "og:"
	MetaOpenGraphType        = MetaOpenGraphPrefix + "type"
	MetaOpenGraphTitle       = MetaOpenGraphPrefix + "title"
	MetaOpenGraphDescription = MetaOpenGraphPrefix + "description"
	MetaOpenGraphURL         = MetaOpenGraphPrefix + "url"
	MetaOpenGraphImage       = MetaOpenGraphPrefix + "image"
)

type openGraph struct {
	Type        string
	Title       string
	Description string
	Site        string
	URL         string
	Image       string
}

func (r PageValidator) testOpenGraphCard(meta headMeta) func(t *testing.T) {
	return func(t *testing.T) {
		card := openGraph{}
		for property, content := range meta.Properties {
			switch property {
			case MetaOpenGraphType:
				card.Type = content
			case MetaOpenGraphTitle:
				card.Title = content
			case MetaOpenGraphDescription:
				card.Description = content
			case MetaOpenGraphURL:
				card.URL = content
			case MetaOpenGraphImage:
				card.Image = content
			default:
				if strings.HasPrefix(property, MetaOpenGraphPrefix) {
					t.Log("unknown Open Graph property:", property, content)
				}
			}
		}

		var err error
		if card.Type == "" {
			t.Error(MetaOpenGraphType + " not found")
		}
		if card.Title == "" {
			t.Error(MetaOpenGraphTitle + " not found")
		} else if err = r.OpenGraphCardTitle.Validate(card.Title); err != nil {
			t.Error(MetaOpenGraphTitle+" validation failed:", err)
		}
		if card.Description == "" {
			t.Error(MetaOpenGraphDescription + " not found")
		} else if err = r.OpenGraphCardDescription.Validate(card.Description); err != nil {
			t.Error(MetaOpenGraphDescription+" validation failed:", err)
		}
		// if card.URL == "" {
		// 	t.Error(MetaOpenGraphURL + " not found")
		// }
		if card.Image == "" {
			t.Error(MetaOpenGraphImage + " not found")
		}
		for name, _ := range meta.Data {
			if strings.HasPrefix(name, MetaOpenGraphPrefix) {
				t.Log("Open Graph meta was set with \"name\" instead of the \"property\" attribute:", name)
			}
		}
	}

}

func (r PageValidator) TestOpenGraphCard(node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		head := findElementNode(node, "head")
		if head == nil {
			t.Fatal("document <head> node is absent")
		}
		r.testOpenGraphCard(getHeadMeta(t, head))(t)
	}
}
