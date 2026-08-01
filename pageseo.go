package pageseo

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"regexp"
	"testing"

	"golang.org/x/net/html"
)

//go:generate go run ./testdata/generate.go

type StringConstraints struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

type Requirements struct {
	// Normalizer is passed to all default validator constructors.
	// If you are using custom validators, you should pass your
	// own normalizer to each constructor manually.
	//
	// Default value is [PassthroughNormalizer] that does not do anything.
	Normalizer Normalizer

	DeduplicationNamespace               string
	TitleDeduplicator                    ValidationMiddleware
	DescriptionDeduplicator              ValidationMiddleware
	OpenGraphCardTitleDeduplicator       ValidationMiddleware
	OpenGraphCardDescriptionDeduplicator ValidationMiddleware
	TwitterCardTitleDeduplicator         ValidationMiddleware
	TwitterCardDescriptionDeduplicator   ValidationMiddleware

	Title       Validator
	Description Validator
	Heading     Validator
	Language    Validator

	URL          Validator
	LinkText     Validator
	ImageAltText Validator
	ImageSrc     Validator
}

type PageValidator struct {
	Loader                   Loader
	Title                    Validator
	Description              Validator
	OpenGraphCardTitle       Validator
	OpenGraphCardDescription Validator
	TwitterCardTitle         Validator
	TwitterCardDescription   Validator
	Heading                  Validator
	Language                 Validator

	URL          Validator
	LinkText     Validator
	ImageAltText Validator
	ImageSrc     Validator

	cachedURLs *cachedParsedURLs
}

func New(loader Loader, r Requirements) PageValidator {
	if loader == nil {
		panic("nil loader")
	}
	if r.Normalizer == nil {
		r.Normalizer = PassthroughNormalizer
	}

	if r.TitleDeduplicator == nil {
		r.TitleDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.DescriptionDeduplicator == nil {
		r.DescriptionDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.OpenGraphCardTitleDeduplicator == nil {
		r.OpenGraphCardTitleDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.OpenGraphCardDescriptionDeduplicator == nil {
		r.OpenGraphCardDescriptionDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.TwitterCardTitleDeduplicator == nil {
		r.TwitterCardTitleDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.TwitterCardDescriptionDeduplicator == nil {
		r.TwitterCardDescriptionDeduplicator = NewDeduplicator(r.DeduplicationNamespace)
	}

	if r.Title == nil {
		r.Title = NewTitleValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.Description == nil {
		r.Description = NewDescriptionValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.Heading == nil {
		r.Heading = NewHeadingValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.Language == nil {
		r.Language = ValidatorFunc(func(s string) error {
			if !regexp.MustCompile(`^\w\w(\-\w\w)?$`).MatchString(s) {
				return errors.New("invalid language code")
			}
			return nil
		})
	}
	if r.URL == nil {
		r.URL = NewURLValidator(StringConstraints{})
	}
	if r.LinkText == nil {
		r.LinkText = NewLinkTextValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.ImageAltText == nil {
		r.ImageAltText = NewImageAltTextValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.ImageSrc == nil {
		r.ImageSrc = NewURLValidator(StringConstraints{
			MaximumLength: DefaultMaximumImageSourceLength,
		})
	}

	return PageValidator{
		Loader:                   loader,
		Title:                    r.TitleDeduplicator.Wrap(r.Title),
		Description:              r.DescriptionDeduplicator.Wrap(r.Description),
		OpenGraphCardTitle:       r.OpenGraphCardTitleDeduplicator.Wrap(r.Title),
		OpenGraphCardDescription: r.OpenGraphCardDescriptionDeduplicator.Wrap(r.Description),
		TwitterCardTitle:         r.TwitterCardTitleDeduplicator.Wrap(r.Title),
		TwitterCardDescription:   r.TwitterCardDescriptionDeduplicator.Wrap(r.Description),
		Heading:                  r.Heading,
		Language:                 r.Language,

		URL:          r.URL,
		LinkText:     r.LinkText,
		ImageAltText: r.ImageAltText,
		ImageSrc:     r.ImageSrc,

		cachedURLs: &cachedParsedURLs{},
	}
}

func NewStrict(loader Loader, r Requirements) PageValidator {
	if r.Normalizer == nil {
		r.Normalizer = NormalizeTextToNFC
	}
	if r.Title == nil {
		r.Title = NewTitleValidator(StringConstraints{Normalizer: NormalizeLineToNFC})
	}
	if r.Heading == nil {
		r.Heading = NewHeadingValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.LinkText == nil {
		r.LinkText = NewLinkTextValidator(StringConstraints{Normalizer: NormalizeLineToNFC})
	}
	if r.ImageAltText == nil {
		r.ImageAltText = NewImageAltTextValidator(StringConstraints{Normalizer: NormalizeLineToNFC})
	}
	return New(loader, r)
}

func (r PageValidator) Test(origin string, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		var ok bool
		originURL, err := r.cachedURLs.Get(origin)
		if err != nil {
			t.Fatal("invalid URL:", err)
		}
		if node.FirstChild == nil {
			t.Fatal("page contains no HTML nodes")
		}
		err = ValidateDoctypeTag(node.FirstChild)
		if err != nil {
			t.Errorf("page has an invalid <DOCTYPE> tag: %v", err)
		}
		TestDocumentRootHasExactlyDoctypeAndHTMLNodes(node)(t)
		attributes := getAttributes(t, node.FirstChild.NextSibling)
		language, ok := attributes["lang"]
		if !ok {
			t.Error("HTML tag is missing a lang attribute")
		}
		if err = r.Language.Validate(language); err != nil {
			t.Errorf("HTML tag has an invalid lang attribute %q: %v", language, err)
		}

		nextChild, closer := iter.Pull[*html.Node](node.FirstChild.NextSibling.ChildNodes())
		defer closer()
		var child *html.Node

		for {
			child, ok = nextChild()
			if !ok {
				t.Error("HTML tag is missing a <HEAD> tag at the top")
				break
			}
			if child.Type != html.ElementNode {
				continue
			}
			if child.Data != "head" {
				t.Errorf("first child element tag is not a <HEAD> tag: %s", child.Data)
				break
			}
			t.Run(getElementPath(child), r.TestHead(child))
			break // found a head tag
		}

		for {
			child, ok := nextChild()
			if !ok {
				t.Fatal("HTML tag is missing a <BODY> tag")
			}
			if child.Type != html.ElementNode {
				continue
			}
			if child.Data != "body" {
				t.Fatalf("second child element tag is not a <BODY> tag: %s", child.Data)
			}
			t.Run(getElementPath(child), r.TestHeadings(child))

			hotSwap := r.preloadResources(t, originURL, child)
			for node := range child.Descendants() {
				switch node.Data {
				case "a":
					t.Run(getElementPath(node), r.testLink(originURL, node, hotSwap))
				case "img":
					t.Run(getElementPath(node), r.testImage(originURL, node, hotSwap))
				}
			}

			break // found a body tag
		}

		child, ok = nextChild()
		if ok {
			t.Errorf("HTML tag contains more than two children: %s", child.Data)
		}

		// for node := range node.Descendants() {
		// 	if node.Type != html.ElementNode {
		// 		continue
		// 	}
		// 	switch node.Data {
		// 	case "a":
		// 		// if r.LinkText == SkipValidator {
		// 		// 	continue
		// 		// }
		// 		t.Run(Path(node), r.TestLink(originURL, node))
		// 	case "img":
		// 		// if (r.ImageAltText == nil || r.ImageAltText == SkipValidator) && (r.ImageSrc == nil || r.ImageSrc == SkipValidator) {
		// 		// 	continue
		// 		// }
		// 		t.Run(Path(node), r.TestImage(origin, node))
		// 		// if err = ValidateImage(node); err != nil {
		// 		// 	t.Errorf("invalid link tag %q: %v", Path(node), err)
		// 		// }
		// 		// case "script":
		// 		// 	t.Run("script tag has valid attributes", r.TestScript(node))
		// 		// case "style":
		// 		// 	t.Run("style tag has valid attributes", r.TestStyle(node))
		// 	}
		// }
	}
}

func (v PageValidator) TestReader(
	origin string,
	r io.Reader,
) func(t *testing.T) {
	return func(t *testing.T) {
		tree, err := html.Parse(r)
		if err != nil {
			t.Fatalf("unable to parse the HTML page: %v", err)
		}
		if tree == nil {
			t.Fatal("no HTML tree found in the reader")
		}
		v.Test(origin, tree)(t)
	}
}

func (v PageValidator) TestFile(p string) func(t *testing.T) {
	return func(t *testing.T) {
		f, err := os.Open(p)
		if err != nil {
			t.Fatalf("unable to open file %q: %v", p, err)
		}
		t.Cleanup(func() {
			if cerr := f.Close(); cerr != nil {
				t.Errorf("unable to close HTML file %q: %v", p, cerr)
			}
		})
		v.TestReader("file://"+p, f)(t)
	}
}

func ValidateDoctypeTag(node *html.Node) error {
	if node == nil {
		return errors.New("!DOCTYPE node is nil")
	}
	// TODO: this was glitching out for some reason
	if node.Type != html.DoctypeNode {
		return errors.New("HTML node is not a DOCTYPE tag")
	}
	if node.Data != "html" {
		return fmt.Errorf("DOCTYPE tag contains unexpected root element: %s", node.Data)
	}
	return nil
}
