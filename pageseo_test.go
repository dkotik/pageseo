package pageseo

import (
	"bytes"
	"net/url"
	"os"
	"testing"

	"github.com/dkotik/pageseo/internal"
	"golang.org/x/net/html"
)

var testData = NewFS(os.DirFS("testdata"))

func TestGoldenExecCapture(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping golden test in short mode")
	}
	internal.RunGoldenTest(t, "minimal", "TestMinimalPage")
}

func TestMinimalPage(t *testing.T) {
	const rootWarning = "the root element must be a html.DocumenatNode; if this assumption does not hold, library will not function properly, especially in cases where Matching is triggered at root node; neither there must be multiple html.DocumenatNode nodes"
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
		// empty or empty tag documents must still have
		// a document node as their root
		empty, err := html.Parse(bytes.NewReader([]byte("")))
		if err != nil {
			t.Fatal(err)
		}
		if empty == nil {
			t.Fatal("single has no first child")
		}
		if empty.Type != html.DocumentNode {
			t.Log(rootWarning)
			t.Fatal("tree is not a document node:", empty.Type)
		}

		single, err := html.Parse(bytes.NewReader([]byte("<hR aTTr=\"VaL\" />")))
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

		hr := single.FirstChild.FirstChild.NextSibling.FirstChild
		if hr.Data != "hr" {
			t.Log("ensuring that element tags .Data is converted to lower case: hR => hr; this assumption holds in the rest of pageSEO library for matching and everything else")
			t.Fatal("<hr> has unexpected data:", hr.Data)
		} else {
			if hr.Attr[0].Key != "attr" {
				t.Log("ensuring that element attribute .Key is converted to lower case: aTTr => attr; this assumption holds in the rest of pageSEO library for matching and everything else")
				t.Fatal("<hr> has unexpected attribute key:", hr.Attr[0].Key)
			}
			if hr.Attr[0].Val != "VaL" {
				t.Fatal("<hr> has unexpected attribute value:", hr.Attr[0].Val)
			}
		}
	}

	pageSEO := NewPageTester(testData)
	pageSEO.TestPage(&url.URL{}, tree)(t)
	// internal.RunGoldenTest(t, "minimal", []testing.InternalTest{
	// 	internal.NewTest(t.Name(), pageSEO.TestPage(&url.URL{}, tree)),
	// })
}
