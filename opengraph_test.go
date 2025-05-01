package pageseo

import (
	"testing"

	"golang.org/x/net/html"
)

func TestOpenGraphCard(t *testing.T) {
	f, err := testData.Open("testdata/opengraph.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tree, err := html.Parse(f)
	if err != nil {
		t.Fatal("unable to parse HTML tree:", err)
	}
	if tree == nil {
		t.Fatal("html.Parse returned nil")
	}

	NewStrict(Requirements{}).TestOpenGraphCard(tree)(t)
}
