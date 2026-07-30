package pageseo

import (
	"bytes"
	"testing"

	"golang.org/x/net/html"
)

func TestOpenGraphCard(t *testing.T) {
	f, _, err := testData.Load(t.Context(), "opengraph.html")
	if err != nil {
		t.Fatal(err)
	}

	tree, err := html.Parse(bytes.NewReader(f))
	if err != nil {
		t.Fatal("unable to parse HTML tree:", err)
	}
	if tree == nil {
		t.Fatal("html.Parse returned nil")
	}

	NewStrict(testData, Requirements{}).TestOpenGraphCard(tree)(t)
}
