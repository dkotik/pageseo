package internal

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// WP is the warning prefix used in log messages.
const WP = "|WARNING|"

func GetTestName(node *html.Node) string {
	switch node.Type {
	case html.TextNode:
		return ":text:"
	case html.CommentNode:
		return ":comment:"
	default:
		return "<" + node.Data + ">"
	}
}

func WriteElementPath(w io.Writer, node *html.Node) {
	segments := []string{node.Data}
	for ancestor := range node.Ancestors() {
		if ancestor.Type == html.ElementNode {
			segments = append(segments, ancestor.Data)
		}
	}
	count := len(segments)
	if count < 3 {
		return
	}
	if segments[count-1] == "html" {
		count--
	}
	_, _ = w.Write([]byte(`└■ `))
	for i := count - 1; i >= 0; i-- {
		_, _ = w.Write([]byte(segments[i]))
		if i > 0 {
			_, _ = w.Write([]byte(`›`))
		}
	}
	for _, attr := range node.Attr {
		if attr.Key == "id" {
			_, _ = w.Write([]byte(`#`))
			_, _ = w.Write([]byte(attr.Val))
			break
		}
	}
	_, _ = w.Write([]byte{'\n'})
}

func LogAttributes(t testing.TB, attrs []html.Attribute) {
	t.Helper()
	maximumLength := 0
	length := 0
	filtered := make([]html.Attribute, 0, len(attrs))
	duplicateKeys := []string{}
	for _, attr := range attrs {
		if strings.HasPrefix(attr.Key, "data-") {
			continue
		}
		length = len(attr.Key)
		if length > maximumLength {
			maximumLength = length
		}
		if slices.Index(filtered, attr) > -1 {
			duplicateKeys = append(duplicateKeys, attr.Key)
		}
		filtered = append(filtered, attr)
	}
	if len(filtered) == 0 {
		return
	}
	w := t.Output()
	for _, attr := range filtered {
		if len(attr.Val) < 48 {
			_, _ = fmt.Fprintf(w, " │ %*s: %s", maximumLength, attr.Key, attr.Val)
		} else {
			_, _ = fmt.Fprintf(w, " │ %*s: %s", maximumLength, attr.Key, truncateMiddle(attr.Val, 48))
		}

		if attr.Val == "" {
			_, _ = w.Write([]byte(WP + " EMPTY \n"))
		} else if slices.Index(duplicateKeys, attr.Key) != -1 {
			_, _ = w.Write([]byte(" " + WP + " DUPLICATE \n"))
		} else {
			_, _ = w.Write([]byte("\n"))
		}
	}
	_, _ = w.Write([]byte(" └───────────────\n"))
}

func truncateMiddle(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen] + "…"
	}

	remaining := maxLen - 1
	leftCount := remaining / 2
	rightCount := remaining - leftCount
	leftSide := string(runes[:leftCount])
	rightSide := string(runes[len(runes)-rightCount:])

	return leftSide + "…" + rightSide
}
