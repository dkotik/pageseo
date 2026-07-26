package pageseo

import (
	"testing"

	"golang.org/x/net/html"
)

func logAttributes(t *testing.T, attrs []html.Attribute) {
	t.Helper()
	maximumLength := 0
	length := 0
	for _, attr := range attrs {
		length = len(attr.Key)
		if length > maximumLength {
			maximumLength = length
		}
	}
	for _, attr := range attrs {
		if len(attr.Val) < 24 {
			t.Logf("%*s = %s", maximumLength, attr.Key, attr.Val)
		} else {
			t.Logf("%*s = ...%s", maximumLength, attr.Key, truncateMiddle(attr.Val, 24))
		}
	}
}

func truncateMiddle(s string, maxLen int) string {
	runes := []rune(s)

	// Return original string if it's already short enough
	if len(runes) <= maxLen {
		return s
	}

	// If maxLen is too small to fit the ellipsis, handle gracefully
	if maxLen <= 3 {
		return "..."[:maxLen]
	}

	// Calculate how many characters to keep on each side
	remaining := maxLen - 3
	leftCount := remaining / 2
	rightCount := remaining - leftCount

	// Combine the left side, ellipsis, and right side
	leftSide := string(runes[:leftCount])
	rightSide := string(runes[len(runes)-rightCount:])

	return leftSide + "..." + rightSide
}
