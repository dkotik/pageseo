package pageseo

import (
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/internal"
)

const (
	MetaOpenGraphPrefix      = "og:"
	MetaOpenGraphType        = MetaOpenGraphPrefix + "type"
	MetaOpenGraphTitle       = MetaOpenGraphPrefix + "title"
	MetaOpenGraphDescription = MetaOpenGraphPrefix + "description"
	MetaOpenGraphURL         = MetaOpenGraphPrefix + "url"
	MetaOpenGraphImage       = MetaOpenGraphPrefix + "image"
	MetaOpenGraphImageAlt    = MetaOpenGraphPrefix + "image:alt"
	MetaOpenGraphImageType   = MetaOpenGraphPrefix + "image:type"
	MetaOpenGraphImageHeight = MetaOpenGraphPrefix + "image:height"
	MetaOpenGraphImageWidth  = MetaOpenGraphPrefix + "image:width"
	MetaOpenGraphSiteName    = MetaOpenGraphPrefix + "site_name"
)

type openGraphMeta struct {
	Type        string
	Title       string
	Description string
	Site        string
	URL         string
	Image       string
	ImageAlt    string
	ImageType   string
	ImageHeight string
	ImageWidth  string
	SiteName    string
}

func TestOpenGraphMeta(
	t testing.TB,
	metaProperties map[string]string,
	requirements HeadNodeConstraints,
) {
	t.Helper()
	var og openGraphMeta
	unknownProperties := []string{}
	for property, content := range metaProperties {
		switch property {
		case MetaOpenGraphTitle:
			og.Title = content
		case MetaOpenGraphType:
			og.Type = content
		case MetaOpenGraphImage:
			og.Image = content
		case MetaOpenGraphURL:
			og.URL = content
		case MetaOpenGraphDescription:
			og.Description = content
		case MetaOpenGraphImageAlt:
			og.ImageAlt = content
		case MetaOpenGraphImageType:
			og.ImageType = content
		case MetaOpenGraphImageHeight:
			og.ImageHeight = content
		case MetaOpenGraphImageWidth:
			og.ImageWidth = content
		case MetaOpenGraphSiteName:
			og.SiteName = content
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

	// The only mandatory basic tags are
	//   og:title, og:type, og:image, and og:url
	if og.Type == "" {
		t.Error(MetaOpenGraphType + " not found")
	} else {
		switch og.Type {
		case "website", "article", "book", "profile":
		case "music.album", "music.song", "music.playlist", "music.radio_station":
		case "video.movie", "video.episode", "video.tv_show", "video.other":
		default:
			t.Error(MetaOpenGraphType + " is not a common type: " + og.Type)
		}
	}
	if og.Title == "" {
		t.Error(MetaOpenGraphTitle + " not found")
	} else {
		requirements.Title.apply(t, MetaOpenGraphTitle, og.Title)
	}
	if og.Image == "" {
		t.Error(MetaOpenGraphImage + " not found")
	}
	if og.URL == "" {
		t.Error(MetaOpenGraphURL + " not found")
	}

	// not mandatory
	if og.Description == "" {
		t.Log(internal.WP, MetaOpenGraphDescription, " not found")
	} else {
		requirements.Description.apply(t, MetaOpenGraphDescription, og.Description)
	}
	if og.ImageAlt == "" {
		t.Log(internal.WP, MetaOpenGraphImageAlt, " not found")
	} else {
		// TODO: should be imagealt validator here
		requirements.Title.apply(t, MetaOpenGraphImageAlt, og.ImageAlt)
	}
	if og.ImageType == "" {
		t.Log(internal.WP, MetaOpenGraphImageType, " not found")
	}
	if og.ImageHeight == "" {
		t.Log(internal.WP, MetaOpenGraphImageHeight, " not found")
	}
	if og.ImageWidth == "" {
		t.Log(internal.WP, MetaOpenGraphImageWidth, " not found")
	}
	if og.SiteName == "" {
		t.Log(internal.WP, MetaOpenGraphSiteName, " not found")
	} else {
		_, err := url.Parse(og.SiteName)
		if err != nil {
			t.Error(MetaOpenGraphSiteName+" is not a valid URL:", err)
		}
	}
}
