package pageseo

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const pathPrefixForHTML = "•"

func findElementNode(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && strings.ToLower(node.Data) == name {
		return node
	}
	for node = range node.Descendants() {
		if node.Type == html.ElementNode && strings.ToLower(node.Data) == name {
			return node
		}
	}
	return nil
}

func getElementPath(node *html.Node) string {
	if node == nil {
		return ""
	}
	segments := []string{node.Data}
	for ancestor := range node.Ancestors() {
		if ancestor.Type == html.ElementNode {
			segments = append(segments, ancestor.Data)
		}
	}
	slices.Reverse(segments)
	return pathPrefixForHTML + strings.Join(segments, ">")
}

func getAttributes(t *testing.T, node *html.Node) map[string]string {
	attrs := make(map[string]string, len(node.Attr))
	var ok bool
	for _, attr := range node.Attr {
		key := strings.ToLower(attr.Key)
		if _, ok = attrs[key]; ok {
			t.Errorf("duplicate tag attribute found: %s=%q", attr.Key, attr.Val)
		}
		attrs[key] = attr.Val
	}
	return attrs
}

func ParseCommaSeparatedKeyedValues(s string) (map[string]string, error) {
	values := make(map[string]string)
	var ok bool
	for _, pair := range strings.Split(s, ",") {
		key, value, _ := strings.Cut(pair, "=")
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok = values[key]; ok {
			return nil, fmt.Errorf("duplicate tag attribute found: %s", key)
		}
		values[key] = strings.TrimSpace(value)
	}
	return values, nil
}

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
