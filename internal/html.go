package internal

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/text/language"
)

var reValidLanguageCode = regexp.MustCompile(`^\w\w(\-\w\w)?$`)

func ValidateLanguage(t testing.TB, lang string) {
	if _, err := language.Parse(lang); err != nil {
		t.Errorf("<html> BCP47 [lang] attribute %q is not canonical: %v", lang, err)
	}
	if !reValidLanguageCode.MatchString(lang) {
		t.Errorf("<html> BCP47 [lang] attribute %q is not a valid language code", lang)
	}
}

// GetAttributes returns the attributes of the node as a map[string]string.
//
// If an HTML node has multiple attributes with the identical name,
// the browser will only recognize the first occurrence and will
// completely ignore all subsequent duplicates.
func GetAttributes(t testing.TB, node *html.Node) map[string]string {
	attrs := make(map[string]string, len(node.Attr))
	ok := false
	for _, attr := range node.Attr {
		if _, ok = attrs[attr.Key]; ok {
			continue // only the first value remains
		}
		attrs[attr.Key] = attr.Val
	}
	return attrs
}

func GetFirstElementOrSibling(node *html.Node) *html.Node {
	for {
		if node == nil {
			return nil
		}
		switch node.Type {
		case html.ElementNode:
			return node
		default:
			node = node.NextSibling
		}
	}
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

func GetText(node *html.Node) string {
	b := strings.Builder{}
	if node.Type == html.TextNode {
		_, _ = b.WriteString(strings.TrimSpace(node.Data))
		_, _ = b.WriteRune(' ')
	}
	last := node.LastChild
	for descendant := range node.Descendants() {
		if descendant.Type != html.TextNode {
			continue
		}
		_, _ = b.WriteString(strings.TrimSpace(descendant.Data))
		if descendant != last {
			_, _ = b.WriteRune(' ')
		}
	}
	return b.String()
}
