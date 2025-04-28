package slug

import (
	"regexp"
	"strings"

	"github.com/alexsergivan/transliterator"
	"github.com/dkotik/pageseo"
)

var (
	defaultNormalizer = Normalizer{
		Transliterator:   transliterator.NewTransliterator(nil),
		LanguageShortTag: "",
		MaximumWordCount: 9,
	}.WithStopWords(StopWords[:])
	reWordSplitter = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

type Normalizer struct {
	Transliterator   *transliterator.Transliterator
	LanguageShortTag string
	StopWords        map[string]struct{}
	MinimumWordCount int
	MaximumWordCount int
}

func (n Normalizer) WithStopWords(stopWords []string) Normalizer {
	cp := make(map[string]struct{}, len(n.StopWords))
	for word := range n.StopWords {
		cp[word] = struct{}{}
	}
	for _, word := range stopWords {
		cp[word] = struct{}{}
	}
	return Normalizer{
		Transliterator:   n.Transliterator,
		LanguageShortTag: n.LanguageShortTag,
		StopWords:        cp,
		MaximumWordCount: n.MaximumWordCount,
	}
}

func (n Normalizer) Normalize(text string) (string, error) {
	b := &strings.Builder{}
	foundWords := 0
	ok := false
	for _, word := range reWordSplitter.Split(
		strings.ToLower(n.Transliterator.Transliterate(text, n.LanguageShortTag)), 12) {
		if len(word) == 0 {
			continue
		}
		if _, ok = n.StopWords[word]; ok {
			continue
		}
		if len(word)+b.Len() > pageseo.DefaultMaximumTitleLength {
			break
		}
		_, _ = b.WriteString(word)
		_, _ = b.WriteRune(Hyphen)
		foundWords++
		if foundWords >= n.MaximumWordCount {
			break
		}
	}
	if b.Len() < pageseo.DefaultMinimumTitleLength || foundWords < n.MinimumWordCount {
		return "", ErrTooShort
	}
	return strings.TrimSuffix(b.String(), string(Hyphen)), nil
}
