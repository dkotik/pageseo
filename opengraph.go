package pageseo

import (
	"slices"
	"strings"
	"testing"
)

const (
	MetaOpenGraphPrefix      = "og:"
	MetaOpenGraphType        = MetaOpenGraphPrefix + "type"
	MetaOpenGraphTitle       = MetaOpenGraphPrefix + "title"
	MetaOpenGraphDescription = MetaOpenGraphPrefix + "description"
	MetaOpenGraphURL         = MetaOpenGraphPrefix + "url"
	MetaOpenGraphImage       = MetaOpenGraphPrefix + "image"
)

type openGraphMeta struct {
	Type        string
	Title       string
	Description string
	Site        string
	URL         string
	Image       string
}

func TestOpenGraphMeta(
	t testing.TB,
	metaProperties map[string]string,
	requirements HeadNodeConstraints,
) {
	var og openGraphMeta
	unknownProperties := []string{}
	for property, content := range metaProperties {
		switch property {
		case MetaOpenGraphType:
			og.Type = content
		case MetaOpenGraphTitle:
			og.Title = content
		case MetaOpenGraphDescription:
			og.Description = content
		case MetaOpenGraphURL:
			og.URL = content
		case MetaOpenGraphImage:
			og.Image = content
		default:
			if strings.HasPrefix(property, MetaOpenGraphPrefix) {
				unknownProperties = append(unknownProperties, property)
			}
		}
	}

	if len(unknownProperties) > 0 {
		slices.Sort(unknownProperties)
		t.Log("unknown Open Graph properties:")
		for _, property := range unknownProperties {
			t.Log(" - ", property)
		}
	}

	if og.Type == "" {
		t.Error(MetaOpenGraphType + " not found")
	}
	if og.Title == "" {
		t.Error(MetaOpenGraphTitle + " not found")
	}
	// TODO: apply head node constraints
	// if og.Description == "" {
	// 	t.Error(MetaOpenGraphDescription + " not found")
	// } else if err = requirements.Validate(og.Description); err != nil {
	// 	t.Error(MetaOpenGraphDescription+" validation failed:", err)
	// }
	// if og.URL == "" {
	// 	t.Error(MetaOpenGraphURL + " not found")
	// }
	if og.Image == "" {
		t.Error(MetaOpenGraphImage + " not found")
	}
}
