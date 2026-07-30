package pageseo

import (
	"net/url"
	"testing"
)

func TestIsExternalLocation(t *testing.T) {
	tcs := []struct {
		Origin     string
		Location   string
		IsExternal bool
	}{
		{
			Origin:     "http://www.google.com/",
			Location:   "something",
			IsExternal: false,
		},
		{
			Origin:     "http://www.google.com/",
			Location:   "/something",
			IsExternal: false,
		},
		{
			Origin:     "http://www.google.com/",
			Location:   "//something",
			IsExternal: true,
		},
		{
			Origin:     "http://www.google.com/",
			Location:   "./something",
			IsExternal: false,
		},
		{
			Origin:     "http://www.google.com/",
			Location:   "./something//flocks////",
			IsExternal: false,
		},
	}

	for _, tc := range tcs {
		origin, err := url.Parse(tc.Origin)
		if err != nil {
			t.Fatalf("unable to parse origin <%s>: %v", tc.Origin, err)
		}
		location, err := url.Parse(tc.Location)
		if err != nil {
			t.Fatalf("unable to parse location <%s>: %v", tc.Location, err)
		}
		if IsExternalLocation(origin, location) != tc.IsExternal {
			t.Log("  origin:", tc.Origin)
			t.Log("location:", tc.Location)
			if IsExternalLocation(origin, location) {
				t.Fatal("location is external to origin")
			} else {
				t.Fatal("location is not external to origin")
			}
		}
	}
}

func TestStandardBehaviorOfParseURL(t *testing.T) {
	tcs := []struct {
		URL    string
		Scheme string
		Host   string
		Path   string
	}{
		{
			URL:    "//www.google.com/skdjlf//dslfjskjadf",
			Scheme: "",
			Host:   "www.google.com",
			Path:   "/skdjlf//dslfjskjadf",
		},
		{
			URL:    "http://www.google.com/something",
			Scheme: "http",
			Host:   "www.google.com",
			Path:   "/something",
		},
		{
			URL:    "/something//dfgdfg",
			Scheme: "",
			Host:   "",
			Path:   "/something//dfgdfg",
		},
		{
			URL:    "//something",
			Scheme: "",
			Host:   "something",
			Path:   "",
		},
		{
			URL:    "something//dfgdfg",
			Scheme: "",
			Host:   "",
			Path:   "something//dfgdfg",
		},
	}

	for _, tc := range tcs {
		parsed, err := url.Parse(tc.URL)
		if err != nil {
			t.Fatal(err)
		}
		if parsed.Scheme != tc.Scheme {
			t.Log("  result:", parsed.Scheme)
			t.Log("expected:", tc.Scheme)
			t.Fatal("result does not match expected value")
		}
		if parsed.Host != tc.Host {
			t.Log("  result:", parsed.Host)
			t.Log("expected:", tc.Host)
			t.Fatal("result does not match expected value")
		}
		if parsed.Path != tc.Path {
			t.Log("  result:", parsed.Path)
			t.Log("expected:", tc.Path)
			t.Fatal("result does not match expected value")
		}
	}
}
