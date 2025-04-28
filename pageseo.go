package pageseo

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"regexp"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

type StringConstraints struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

type Requirements struct {
	Title       htmltest.Validator
	Description htmltest.Validator
	Language    htmltest.Validator
}

func (r Requirements) WithDefaults() Requirements {
	if r.Title == nil {
		r.Title = NewTitleValidator(StringConstraints{})
	}
	if r.Description == nil {
		r.Description = NewDescriptionValidator(StringConstraints{})
	}
	if r.Language == nil {
		r.Language = htmltest.ValidatorFunc(func(s string) error {
			if !regexp.MustCompile(`^\w\w(\-\w\w)?$`).MatchString(s) {
				return errors.New("invalid language code")
			}
			return nil
		})
	}
	return r
}

func (r Requirements) Test(node *html.Node) func(t *testing.T) {
	r = r.WithDefaults()
	return func(t *testing.T) {
		if node.FirstChild == nil {
			t.Fatal("page contains no HTML nodes")
		}
		err := ValidateDoctypeTag(node.FirstChild)
		if err != nil {
			t.Errorf("page has an invalid <DOCTYPE> tag: %v", err)
		}
		if node.FirstChild.NextSibling.NextSibling != nil {
			t.Log("found an unexpected third root tag:", node.FirstChild.NextSibling.NextSibling.Data)
			t.Fatal("page has an un expected number of root tags: should include only <DOCTYPE> and <HTML> tags")
		}
		if node == nil {
			t.Fatal("HTML node is nil")
		}
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
	}
}

func (v Requirements) TestReader(r io.Reader) func(t *testing.T) {
	return func(t *testing.T) {
		tree, err := html.Parse(r)
		if err != nil {
			t.Fatalf("unable to parse the HTML page: %v", err)
		}
		if tree == nil {
			t.Fatal("no HTML tree found in the reader")
		}
		v.Test(tree)(t)
	}
}

func (v Requirements) TestFile(p string) func(t *testing.T) {
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
		v.TestReader(f)(t)
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
