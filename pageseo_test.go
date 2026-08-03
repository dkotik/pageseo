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
	internal.RunGoldenTest(t, "minimal", []testing.InternalTest{
		internal.NewTest(t.Name(), pageSEO.TestPage(&url.URL{}, tree)),
	})

	// result := captureOut(func() {
	// 	// fmt.Println("sdfsdf")
	// 	pageSEO := NewPageTester(testData)
	// 	// pageSEO.TestPage(&url.URL{}, tree)(t)
	// 	t.Run("name string", func(t *testing.T) {
	// 		pageSEO.TestPage(&url.URL{}, tree)(t)
	// 		t.Fail()
	// 	})
	// 	t.Log("args ...any")
	// })

	// goldie.New(t).Assert(t, "minimal", []byte(result))
}
