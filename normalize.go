package pageseo

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

var (
	reCollapseSpaces   = regexp.MustCompile(`\s+`)
	reCollapseNewlines = regexp.MustCompile(`(\n\r?)+`)
)

type Normalizer interface {
	Normalize(string) (string, error)
}

type NormalizerFunc func(string) (string, error)

func (fn NormalizerFunc) Normalize(s string) (string, error) {
	return fn(s)
}

var URLNormalizer NormalizerFunc = func(s string) (string, error) {
	prefix, s, _ := strings.Cut(s, "//")
	var c rune
	var i int
	for i, c = range prefix {
		switch c {
		case ':':
			if i != len(prefix)-1 {
				return "", fmt.Errorf("colon not at end of URL schema: %s", prefix)
			}
		case
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		default:
			return "", fmt.Errorf("invalid character in URL schema: %c (\\x%x)", c, c)
		}
	}

	var prev rune
	for _, c = range s {
		switch c {
		case '/', ':', '@', '#':
			if prev == c {
				return "", fmt.Errorf("double '%c' in URL: %c (\\x%x)", c, c, c)
			}
		case
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '%',
			'-', '.', '_', '~', '$', '&', '+', ',', '=':
		default:
			return "", fmt.Errorf("invalid character in URL: %c (\\x%x)", c, c)
		}
		prev = c
	}
	return s, nil
}

var PassthroughNormalizer NormalizerFunc = func(s string) (string, error) {
	return s, nil
}

var NormalizeLineToNFC NormalizerFunc = func(line string) (string, error) {
	line = norm.NFC.String(strings.TrimSpace(line))
	return reCollapseSpaces.ReplaceAllString(line, " "), nil
}

var NormalizeTextToNFC NormalizerFunc = func(text string) (line string, err error) {
	b := strings.Builder{}
	for _, line = range reCollapseNewlines.Split(text, -1) {
		line, err = NormalizeLineToNFC(line)
		if err != nil {
			return "", err
		}
		b.WriteString(line)
		b.WriteString("\n\n")
	}
	return strings.TrimSuffix(b.String(), "\n\n"), nil
}
