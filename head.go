package pageseo

import (
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type headRequirements struct {
	FoundValidViewPort    bool
	FoundValidCharset     bool
	FoundValidTitle       bool
	FoundValidDescription bool
}

func (r Requirements) TestHead(node *html.Node) func(t *testing.T) {
	r = r.WithDefaults()
	return func(t *testing.T) {
		found := headRequirements{}
		t.Cleanup(func() {
			if !found.FoundValidViewPort {
				t.Errorf("valid viewport tag not found")
			}
			if !found.FoundValidCharset {
				t.Errorf("valid meta charset tag not found")
			}
			if !found.FoundValidTitle {
				t.Errorf("valid title tag not found")
			}
			if !found.FoundValidDescription {
				t.Errorf("valid meta description tag not found")
			}
		})
		// TODO: implement UniqueConstraint(Validator) Validator
		var err error
		for child := range node.ChildNodes() {
			switch child.Data {
			case "title":
				if found.FoundValidTitle {
					t.Error("there are multiple title tags")
				}
				if child.Type != html.ElementNode {
					t.Errorf("title tag is not of element type: %v", child.Type)
					continue
				}
				if child.FirstChild == nil {
					t.Errorf("title tag is empty")
					continue
				}
				if err = r.Title.Validate(strings.TrimSpace(child.FirstChild.Data)); err != nil {
					t.Errorf("title tag is not valid: %v", err)
					continue
				}
				found.FoundValidTitle = true
			case "meta":
				attributes, err := ParseTagAttributes(child)
				if err != nil {
					t.Errorf("unable to collect tag attributes: %v", err)
				}
				name, ok := attributes["name"]
				if ok {
					content, ok := attributes["content"]
					if !ok {
						t.Errorf("meta tag is missing content attribute: %s", name)
						continue
					}
					switch name {
					case "description":
						if found.FoundValidDescription {
							t.Error("there are multiple description meta tags")
						}
						if err = r.Description.Validate(strings.TrimSpace(content)); err != nil {
							t.Errorf("meta tag content for description is not valid: %v", err)
							continue
						}
						found.FoundValidDescription = true
					case "viewport":
						if found.FoundValidViewPort {
							t.Error("there are multiple viewport meta tags")
						}
						csv, err := ParseCommaSeparatedKeyedValues(content)
						if err != nil {
							t.Errorf("meta tag content for viewport %q is not valid: %v", content, err)
							continue
						}
						width, ok := csv["width"]
						if !ok {
							t.Errorf("meta tag content for viewport %q is missing width attribute", content)
						} else if width == "" {
							t.Errorf("meta tag content for viewport %q has empty width attribute", content)
						}
						scale, ok := csv["initial-scale"]
						if !ok {
							t.Errorf("meta tag content for viewport %q is missing initial scale attribute", content)
						} else if scale == "" {
							t.Errorf("meta tag content for viewport %q has empty initial scale attribute", content)
						}
						if _, err = strconv.ParseFloat(scale, 32); err != nil {
							t.Errorf("meta tag content for viewport scale %q has invalid initial scale attribute: %v", scale, err)
						}
						found.FoundValidViewPort = true
					}
				} else {
					charset, ok := attributes["charset"]
					if ok {
						if found.FoundValidCharset {
							t.Error("there are multiple meta tags with charset attribute")
						}
						if charset == "utf-8" {
							found.FoundValidCharset = true
						} else {
							t.Errorf("meta tag content for charset has invalid charset attribute: %s", charset)
						}
					}
				}
			default:
				if child.Type == html.TextNode && len(strings.TrimSpace(child.Data)) == 0 {
					continue
				}
				t.Logf("found unexpected tag: %v", child.Data)
			}
		}
	}
}

func (r Requirements) TestHeadings(node *html.Node) func(t *testing.T) {
	r = r.WithDefaults()
	return func(t *testing.T) {
		foundValidH1 := false
		t.Cleanup(func() {
			if !foundValidH1 {
				t.Errorf("H1 tag not found under %q tag", node.Data)
			}
		})

		var err error
		for descendant := range node.Descendants() {
			if descendant.Type != html.ElementNode {
				continue
			}
			switch descendant.Data {
			case "h1":
				text := ParseTextContent(descendant)
				if text == "" {
					t.Errorf("H1 tag has no text content")
					continue
				}
				if err = r.Title.Validate(text); err != nil {
					t.Errorf("H1 tag text content is not valid: %v", err)
					continue
				}
				foundValidH1 = true
			case "h2", "h3", "h4", "h5", "h6":
				text := ParseTextContent(descendant)
				if text == "" {
					t.Errorf("heading tag %q has no text content", descendant.Data)
					continue
				}
				if err = r.Title.Validate(text); err != nil {
					t.Errorf("heading tag %q text content is not valid: %v", descendant.Data, err)
					continue
				}
			}
		}
	}
}
