package htmltest

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func ParseTextContent(node *html.Node) string {
	b := strings.Builder{}
	if node.Type == html.TextNode {
		_, _ = b.WriteString(strings.TrimSpace(node.Data))
		_, _ = b.WriteRune(' ')
	}
	for descendant := range node.Descendants() {
		if descendant.Type != html.TextNode {
			continue
		}
		_, _ = b.WriteString(strings.TrimSpace(descendant.Data))
		_, _ = b.WriteRune(' ')
	}
	return strings.TrimSuffix(b.String(), ` `)
}

// TODO: deprecate into a pageseo.TestDuplicateAttributes
func ParseAttributes(node *html.Node) (map[string]string, error) {
	attrs := make(map[string]string)
	var ok bool
	for _, attr := range node.Attr {
		if _, ok = attrs[attr.Key]; ok {
			return nil, fmt.Errorf("duplicate tag attribute found: %s", attr.Key)
		}
		attrs[attr.Key] = attr.Val
	}
	return attrs, nil
}
