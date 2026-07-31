package htmltest

import (
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
