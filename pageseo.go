package pageseo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
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
	TitleDeduplicator                    htmltest.Middleware
	DescriptionDeduplicator              htmltest.Middleware
	OpenGraphCardTitleDeduplicator       htmltest.Middleware
	OpenGraphCardDescriptionDeduplicator htmltest.Middleware
	TwitterCardTitleDeduplicator         htmltest.Middleware
	TwitterCardDescriptionDeduplicator   htmltest.Middleware

	Title       htmltest.Validator
	Description htmltest.Validator
	Heading     htmltest.Validator
	Language    htmltest.Validator

	URL          htmltest.Validator
	LinkText     htmltest.Validator
	ImageAltText htmltest.Validator
	ImageSrc     htmltest.Validator
}

type PageValidator struct {
	Title                    htmltest.Validator
	Description              htmltest.Validator
	OpenGraphCardTitle       htmltest.Validator
	OpenGraphCardDescription htmltest.Validator
	TwitterCardTitle         htmltest.Validator
	TwitterCardDescription   htmltest.Validator
	Heading                  htmltest.Validator
	Language                 htmltest.Validator

	URL          htmltest.Validator
	LinkText     htmltest.Validator
	ImageAltText htmltest.Validator
	ImageSrc     htmltest.Validator

	mu                 *sync.Mutex
	deduplicationIndex map[string]struct{}
}

func New(r Requirements) *PageValidator {
	if r.Normalizer == nil {
		r.Normalizer = PassthroughNormalizer
	}

	if r.TitleDeduplicator == nil {
		r.TitleDeduplicator = htmltest.NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.DescriptionDeduplicator == nil {
		r.DescriptionDeduplicator = htmltest.NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.OpenGraphCardTitleDeduplicator == nil {
		r.OpenGraphCardTitleDeduplicator = htmltest.NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.OpenGraphCardDescriptionDeduplicator == nil {
		r.OpenGraphCardDescriptionDeduplicator = htmltest.NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.TwitterCardTitleDeduplicator == nil {
		r.TwitterCardTitleDeduplicator = htmltest.NewDeduplicator(r.DeduplicationNamespace)
	}
	if r.TwitterCardDescriptionDeduplicator == nil {
		r.TwitterCardDescriptionDeduplicator = htmltest.NewDeduplicator(r.DeduplicationNamespace)
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
		r.Language = htmltest.ValidatorFunc(func(s string) error {
			if !regexp.MustCompile(`^\w\w(\-\w\w)?$`).MatchString(s) {
				return errors.New("invalid language code")
			}
			return nil
		})
	}
	if r.URL == nil {
		r.URL = htmltest.NewURLValidator(nil, http.StatusOK)
	}
	if r.LinkText == nil {
		r.LinkText = NewLinkTextValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	if r.ImageAltText == nil {
		r.ImageAltText = NewImageAltTextValidator(StringConstraints{Normalizer: r.Normalizer})
	}
	// if r.ImageSrc == nil {
	// 	r.ImageSrc = NewImageSrcValidator(StringConstraints{Normalizer: r.Normalizer})
	// }

	return &PageValidator{
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

		mu:                 &sync.Mutex{},
		deduplicationIndex: make(map[string]struct{}),
	}
}

func NewStrict(r Requirements) *PageValidator {
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
	return New(r)
}

func (r PageValidator) Test(origin string, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		if node.FirstChild == nil {
			t.Fatal("page contains no HTML nodes")
		}
		err := ValidateDoctypeTag(node.FirstChild)
		if err != nil {
			t.Errorf("page has an invalid <DOCTYPE> tag: %v", err)
		}
		TestDocumentRootHasExactlyDoctypeAndHTMLNodes(node)(t)
		attributes, err := htmltest.ParseAttributes(node.FirstChild.NextSibling)
		if err != nil {
			t.Errorf("failed to parse <HTML> tag attributes: %v", err)
		}
		language, ok := attributes["lang"]
		if !ok {
			t.Error("HTML tag is missing a lang attribute")
		}
		if err = r.Language.Validate(language); err != nil {
			t.Errorf("HTML tag has an invalid lang attribute %q: %v", language, err)
		}

		children, closer := iter.Pull[*html.Node](node.FirstChild.NextSibling.ChildNodes())
		defer closer()

		for {
			child, ok := children()
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
			t.Run("<HEAD> tag contains every required element", r.TestHead(child))
			break // found a head tag
		}

		for {
			child, ok := children()
			if !ok {
				t.Fatal("HTML tag is missing a <BODY> tag")
			}
			if child.Type != html.ElementNode {
				continue
			}
			if child.Data != "body" {
				t.Fatalf("second child element tag is not a <BODY> tag: %s", child.Data)
			}
			t.Run("<BODY> tag contains valid headings", r.TestHeadings(child))
			break // found a body tag
		}

		child, ok := children()
		if ok {
			t.Errorf("HTML tag contains more than two children: %s", child.Data)
		}

		for node := range node.Descendants() {
			if node.Type != html.ElementNode {
				continue
			}
			switch node.Data {
			case "a":
				if r.LinkText == htmltest.SkipValidator {
					continue
				}
				t.Run(htmltest.Path(node), r.TestLink(origin, node))
			case "img":
				// if (r.ImageAltText == nil || r.ImageAltText == htmltest.SkipValidator) && (r.ImageSrc == nil || r.ImageSrc == htmltest.SkipValidator) {
				// 	continue
				// }
				t.Run(htmltest.Path(node), r.TestImage(origin, node))
				// if err = ValidateImage(node); err != nil {
				// 	t.Errorf("invalid link tag %q: %v", htmltest.Path(node), err)
				// }
				// case "script":
				// 	t.Run("script tag has valid attributes", r.TestScript(node))
				// case "style":
				// 	t.Run("style tag has valid attributes", r.TestStyle(node))
			}
		}
	}
}

func (v PageValidator) TestReader(origin string, r io.Reader) func(t *testing.T) {
	return func(t *testing.T) {
		tree, err := html.Parse(r)
		if err != nil {
			t.Fatalf("unable to parse the HTML page: %v", err)
		}
		if tree == nil {
			t.Fatal("no HTML tree found in the reader")
		}
		t.Run(origin, v.Test(origin, tree))
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

func (v PageValidator) TestURL(ctx context.Context, url string) func(t *testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatal("unable to create request", err)
		}
		resp, err := http.DefaultClient.Do(req.WithContext(t.Context()))
		if err != nil {
			t.Fatalf("unable to fetch URL %q: %v", url, err)
		}
		t.Cleanup(func() {
			if cerr := resp.Body.Close(); cerr != nil {
				t.Errorf("unable to close response body for URL %q: %v", url, cerr)
			}
		})
		v.TestReader(url, resp.Body)(t)
	}
}

func ValidateDoctypeTag(node *html.Node) error {
	if node == nil {
		return errors.New("HTML node is nil")
	}
	if node.Type != html.DoctypeNode {
		return errors.New("HTML node is not a DOCTYPE tag")
	}
	if node.Data != "html" {
		return fmt.Errorf("DOCTYPE tag contains unexpected root element: %s", node.Data)
	}
	return nil
}
