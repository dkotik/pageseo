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

type Normalizer func(string) (string, error)

func NormalizeLine(line string) (string, error) {
	line = norm.NFC.String(strings.TrimSpace(line))
	return reCollapseSpaces.ReplaceAllString(line, " "), nil
}

func NormalizeText(text string) (line string, err error) {
	b := strings.Builder{}
	for _, line = range reCollapseNewlines.Split(text, -1) {
		line, err = NormalizeLine(line)
		if err != nil {
			return "", err
		}
		b.WriteString(line)
		b.WriteString("\n\n")
	}
	return strings.TrimSuffix(b.String(), "\n\n"), nil
}
