package pageseo

import (
	"slices"
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
	MetaTwitterImageAlt    = MetaTwitterPrefix + "image:alt"
	MetaTwitterImageType   = MetaTwitterPrefix + "image:type"
	MetaTwitterImageHeight = MetaTwitterPrefix + "image:height"
	MetaTwitterImageWidth  = MetaTwitterPrefix + "image:width"
)

type twitter struct {
	Card        string
	Title       string
	Description string
	Site        string
	URL         string
	Image       string
	ImageAlt    string
	ImageType   string
	ImageHeight string
	ImageWidth  string
}

func TestTwitterMeta(
	t testing.TB,
	metaData map[string]string,
	requirements HeadNodeConstraints,
) {
	card := twitter{}
	unknownProperties := []string{}
	for name, content := range metaData {
		switch name {
		case MetaTwitterCard:
			card.Card = content
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
		case MetaTwitterImageAlt:
			card.ImageAlt = content
		case MetaTwitterImageType:
			card.ImageType = content
		case MetaTwitterImageHeight:
			card.ImageHeight = content
		case MetaTwitterImageWidth:
			card.ImageWidth = content
		default:
			if strings.HasPrefix(name, MetaTwitterPrefix) {
				unknownProperties = append(unknownProperties, name)
			}
		}
	}

	if len(unknownProperties) > 0 {
		slices.Sort(unknownProperties)
		t.Log("unknown Twitter properties:")
		for _, property := range unknownProperties {
			t.Log(" - ", property)
		}
	}

	if card.Card == "" {
		t.Error(MetaTwitterCard + " not found")
	} else {
		switch card.Card {
		case "summary", "summary_large_image", "app", "player":
		default:
			t.Error(MetaTwitterCard + " is not valid")
		}
	}
	if card.Title == "" {
		t.Error(MetaTwitterTitle + " not found")
	} else {
		requirements.Title.apply(t, MetaTwitterTitle, card.Title)
	}
	if card.Description == "" {
		t.Error(MetaTwitterDescription + " not found")
	} else {
		requirements.Description.apply(t, MetaTwitterDescription, card.Description)
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
