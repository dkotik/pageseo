package pageseo

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func logAttributes(t *testing.T, attrs []html.Attribute) {
	t.Helper()
	maximumLength := 0
	length := 0
	filtered := make([]html.Attribute, 0, len(attrs))
	for _, attr := range attrs {
		if strings.HasPrefix(attr.Key, "data-") {
			continue
		}
		length = len(attr.Key)
		if length > maximumLength {
			maximumLength = length
		}
		filtered = append(filtered, attr)
	}
	for _, attr := range filtered {
		if len(attr.Val) < 48 {
			t.Logf("%*s = %s", maximumLength, attr.Key, attr.Val)
		} else {
			t.Logf("%*s = %s", maximumLength, attr.Key, truncateMiddle(attr.Val, 48))
		}
	}
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
