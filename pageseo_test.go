package pageseo

import (
	"embed"
	"testing"
)

//go:embed testdata/*
var testData embed.FS

func TestMinimalPage(t *testing.T) {
	f, err := testData.Open("testdata/minimal.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	NewStrict(Requirements{}).TestReader(t.Name(), f)(t)
}
