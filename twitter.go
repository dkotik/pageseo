package pageseo

import (
	"strings"
	"testing"
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

func TestTwitterMeta(
	t testing.TB,
	metaData map[string]string,
	requirements HeadNodeConstraints,
) {
	card := twitter{}
	for name, content := range metaData {
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

	if card.Type == "" {
		t.Error(MetaTwitterCard + " not found")
	}
	if card.Title == "" {
		t.Error(MetaTwitterTitle + " not found")
	}
	if card.Description == "" {
		t.Error(MetaTwitterDescription + " not found")
	}
	// TODO: enforce head requirements
	// else if err = r.TwitterCardDescription.Validate(card.Description); err != nil {
	// 	t.Error(MetaTwitterDescription+" validation failed:", err)
	// }
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
