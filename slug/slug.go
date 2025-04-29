/*
Package slug provides functions for generating optimized URL slugs from strings.

Reference:

- https://neilpatel.com/blog/seo-urls/
- https://cseo.com/blog/seo-stop-words/
- https://github.com/mozillazg/go-unidecode
*/
package slug

import "strings"

const (
	Hyphen = '-'
)

// New creates a new slug from the given text. Add phrases to exclude words from.
// The exclusion prevents URL slugs from repeating words mentioned by previous URL path segments.
func New(text string, exclude ...string) (string, error) {
	if len(exclude) == 0 {
		return defaultNormalizer.Normalize(text)
	}
	skipWords := []string{}
	for _, phrase := range exclude {
		for _, word := range strings.Split(
			strings.ToLower(defaultNormalizer.Transliterator.Transliterate(phrase, defaultNormalizer.LanguageShortTag)), " ") {
			if word == "" {
				continue
			}
			skipWords = append(skipWords, word)
		}
	}
	return defaultNormalizer.WithStopWords(skipWords).Normalize(text)
}
