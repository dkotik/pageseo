package pageseo

import (
	"bytes"
	"net/url"
	"os"
	"testing"

	"golang.org/x/net/html"
)

var testData = NewFS(os.DirFS("testdata"))

const rootWarning = "the root element must be a html.DocumenatNode; if this assumption does not hold, library will not function properly, especially in cases where Matching is triggered at root node; neither there must be multiple html.DocumenatNode nodes"

func TestMinimalPage(t *testing.T) {
	f, _, err := testData.Load(t.Context(), "minimal.html")
	if err != nil {
		t.Fatal(err)
	}

	tree, err := html.Parse(bytes.NewReader(f))
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Fatal("tree has no first child")
	} else if tree.Type != html.DocumentNode {
		t.Log(rootWarning)
		t.Fatal("tree is not a document node:", tree.Type)
	} else {
		// empty or single tag documents must still have
		// a document node as their root
		single, err := html.Parse(bytes.NewReader([]byte("")))
		if err != nil {
			t.Fatal(err)
		}
		if single == nil {
			t.Fatal("single has no first child")
		}
		if single.Type != html.DocumentNode {
			t.Log(rootWarning)
			t.Fatal("tree is not a document node:", single.Type)
		}

	}

	pageSEO := NewPageTester(testData)
	pageSEO.TestPage(&url.URL{}, tree)(t)
}
