package pageseo

import (
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
