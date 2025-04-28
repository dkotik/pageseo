package pageseo

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func ParseTextContent(node *html.Node) string {
	b := strings.Builder{}
	for descendant := range node.Descendants() {
		if descendant.Type != html.TextNode {
			continue
		}
		_, _ = b.WriteString(strings.TrimSpace(descendant.Data))
		_, _ = b.WriteRune(' ')
	}
	return strings.TrimSuffix(b.String(), ` `)
}

func ParseTagAttributes(node *html.Node) (map[string]string, error) {
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

type Tag struct {
	Name       string
	Attributes map[string]Validator
	Contents   Validator
}

func (t Tag) Match(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == t.Name {
		return true
	}
	return false
}

// func (t Tag) Validate(data string) error {
// 	for attr, validator := range t.AttributeConstraints {
// 		if err := validator.Validate(data); err != nil {
// 			return fmt.Errorf("attribute %q: %w", attr, err)
// 		}
// 	}
// 	return nil
// }
