package pageseo

import (
	"bytes"
	"net/url"
	"os"
	"testing"

	"golang.org/x/net/html"
)

var testData = NewFS(os.DirFS("testdata"))

func TestMinimalPage(t *testing.T) {
	f, _, err := testData.Load(t.Context(), "minimal.html")
	if err != nil {
		t.Fatal(err)
	}

	tree, err := html.Parse(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}

	pageSEO := NewPageTester(testData)
	pageSEO.TestPage(&url.URL{}, tree)(t)
}
