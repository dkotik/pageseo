package pageseo

import (
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type headMeta struct {
	CharacterSet string
	Data         map[string]string
	Properties   map[string]string
}

func getHeadMeta(t *testing.T, head *html.Node) (result headMeta) {
	result.Data = make(map[string]string)
	result.Properties = make(map[string]string)
	ok := false

nextNode:
	for node := range head.ChildNodes() {
		switch node.Type {
		case html.ElementNode: // fallthrough
		case html.CommentNode:
			continue // skip
		case html.TextNode:
			if strings.TrimSpace(node.Data) == "" {
				continue
			}
			fallthrough
		default:
			t.Error("unexpected node in document <head>:", node.Data)
		}

		if strings.ToLower(node.Data) != "meta" {
			continue
		}
		name, property, content := "", "", ""
		for _, attr := range node.Attr {
			switch strings.ToLower(attr.Key) {
			case "name":
				if name != "" {
					t.Error(getElementPath(node)+":", "duplicate value for", attr.Val)
				}
				name = strings.ToLower(attr.Val)
			case "property":
				if property != "" {
					t.Error(getElementPath(node), "duplicate property value for", attr.Val)
				}
				property = strings.ToLower(attr.Val)
			case "content":
				if content != "" {
					t.Error(getElementPath(node), "duplicate content value for", attr.Val)
				}
				content = attr.Val
			case "charset":
				if result.CharacterSet != "" {
					t.Error(getElementPath(node), "duplicate character set:", attr.Val)
				}
				result.CharacterSet = attr.Val
				continue nextNode
			}
		}
		content = strings.TrimSpace(content)
		if content == "" {
			t.Errorf(getElementPath(node), "has no content")
		} else {
			if name == "" {
				if property == "" {
					t.Error(getElementPath(node), "name attribute absent")
				} else {
					if _, ok = result.Properties[property]; ok {
						t.Error(getElementPath(node), "duplicate meta property", property)
					}
					result.Properties[property] = content
				}
			} else {
				if property != "" {
					t.Error(getElementPath(node), "name and property are both set")
				}
				if _, ok = result.Data[name]; ok {
					t.Error(getElementPath(node), "duplicate meta content", name)
				}
				result.Data[name] = content
			}
		}
	}
	return result
}

type headRequirements struct {
	FoundValidViewPort    bool
	FoundValidCharset     bool
	FoundValidTitle       bool
	FoundValidDescription bool
	FoundTwitterCard      bool
	FoundOpenGraphCard    bool
}

func (r PageValidator) TestHead(node *html.Node) func(t *testing.T) {
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
				attributes := getAttributes(t, child)
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
						// if !ok {
						// 	t.Errorf("meta tag content for viewport %q is missing width attribute", content)
						// } else if width == "" {
						// 	t.Errorf("meta tag content for viewport %q has empty width attribute", content)
						// }
						if ok && width == "" {
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
					case MetaTwitterCard, MetaTwitterImage, MetaTwitterTitle, MetaTwitterDescription, MetaTwitterSite, MetaTwitterURL:
						found.FoundTwitterCard = true
					}
				} else {
					charset, ok := attributes["charset"]
					if ok {
						if found.FoundValidCharset {
							t.Error("there are multiple meta tags with charset attribute")
						}
						if strings.ToLower(charset) == "utf-8" {
							found.FoundValidCharset = true
						} else {
							t.Errorf("meta tag content for charset has invalid charset attribute: %s", charset)
						}
					}
				}

				property, ok := attributes["property"]
				if ok {
					_, ok := attributes["content"]
					if !ok {
						t.Errorf("meta tag property is missing content attribute: %s/%s", name, property)
					}
					switch property {
					case MetaOpenGraphType, MetaOpenGraphTitle, MetaOpenGraphDescription, MetaOpenGraphImage, MetaOpenGraphURL:
						found.FoundOpenGraphCard = true
					}
				}
			default:
				if child.Type == html.TextNode && len(strings.TrimSpace(child.Data)) == 0 {
					continue
				}
				// t.Logf("found unexpected tag: %v", child.Data)
			}
		}

		if found.FoundOpenGraphCard {
			t.Run("og:card", r.TestOpenGraphCard(node))
		}
		if found.FoundTwitterCard {
			t.Run("twitter:card", r.TestTwitterCard(node))
		}
	}
}
