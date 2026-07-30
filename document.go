package pageseo

import (
	"testing"

	"golang.org/x/net/html"
)

func getAttribute(node *html.Node, name string) (value string, ok bool) {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val, true
		}
	}
	return "", false
}

func TestDocumentRootHasExactlyDoctypeAndHTMLNodes(root *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		if root.FirstChild.NextSibling == nil {
			t.Fatal("page is missing an HTML tag")
		}
		// if root.FirstChild.NextSibling.NextSibling == nil {
		// 	t.Fatal("page is missing an HTML tag")
		// }
		if root.FirstChild.NextSibling.NextSibling != nil {
			t.Log("found an unexpected third root tag:", root.FirstChild.NextSibling.NextSibling.Data)
			t.Fatal("page has an un expected number of root tags: should include only <DOCTYPE> and <HTML> tags")
		}
	}
}
