package sitemap

import (
	"encoding/xml"
	"testing"

	"github.com/dkotik/pageseo"
)

type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod"`
	ChangeFreq string  `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

type URLSet struct {
	URLs []URL `xml:"url"`
}

type SiteMap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

type SiteMapIndex struct {
	SiteMaps []SiteMap `xml:"sitemap"`
}

func Test(loader pageseo.Loader, URL string) func(testing.TB) {
	return func(t testing.TB) {
		ctx := t.Context()
		data, ct, err := loader.Load(ctx, URL)
		if err != nil {
			t.Fatal(err)
		}
		if ct != "application/xml" && ct != "text/xml" {
			t.Fatalf("expected application/xml, got %s", ct)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty data")
		}

		var index SiteMapIndex
		if err := xml.Unmarshal(data, &index); err != nil {
			t.Fatal(err)
		}
		if len(index.SiteMaps) == 0 {
			var sitemap URLSet
			if err := xml.Unmarshal(data, &sitemap); err != nil {
				t.Fatal(err)
			}
			if len(sitemap.URLs) == 0 {
				t.Fatal("expected non-empty sitemap")
			}
		}
	}
}
