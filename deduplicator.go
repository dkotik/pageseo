package pageseo

import "testing"

type PageLocation struct {
	URL         string
	ElementPath string
}

type DuplicateContentError struct {
	Content   string
	Original  PageLocation
	Duplicate PageLocation
}

func (d DuplicateContentError) Error() string {
	return "duplicate content: " + d.Content
}

func (d DuplicateContentError) TestLog(t testing.TB) {
	if d.Original.URL == d.Duplicate.URL {
		t.Log("same page:", d.Original.URL)
		t.Log("    ", d.Original.ElementPath)
		t.Log("    ", d.Duplicate.ElementPath)
	} else {
		t.Log("first seen:", d.Original.URL)
		t.Log("    ", d.Original.ElementPath)
		t.Log(" duplicate:", d.Duplicate.URL)
		t.Log("    ", d.Duplicate.ElementPath)
	}
	t.Errorf("duplicate content: %s", d.Content)
}

// TODO: implement content deduplicator NodeTester
