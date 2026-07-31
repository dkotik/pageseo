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
				if strings.ToLower(result.CharacterSet) != "utf-8" {
					t.Error("document character set is not UTF-8:", result.CharacterSet)
				}
				continue nextNode
			case "http-equiv":
				if name != "" {
					t.Error(getElementPath(node)+":", "duplicate value for", attr.Val)
				}
				name = "http-equiv:" + strings.ToLower(attr.Val)
			}
		}
		content = strings.TrimSpace(content)
		if content == "" {
			t.Error(getElementPath(node), "has no content")
		} else {
			if name == "" {
				if property == "" {
					t.Error(getElementPath(node), "name attribute absent:", content)
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
	FoundValidCharset     bool
	FoundValidTitle       bool
	FoundValidDescription bool
}

func (r PageValidator) TestViewPort(content string) func(t *testing.T) {
	return func(t *testing.T) {
		if content == "" {
			t.Fatal("meta viewport is empty")
		}
		csv, err := ParseCommaSeparatedKeyedValues(content)
		if err != nil {
			t.Fatalf("meta tag content for viewport %q is not valid: %v", content, err)
		}
		width, ok := csv["width"]
		// if !ok {
		// 	t.Error("meta tag content for viewport %q is missing width attribute", content)
		// } else if width == "" {
		// 	t.Error("meta tag content for viewport %q has empty width attribute", content)
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
	}
}

type head struct {
	Title       string
	Description string
	Meta        headMeta
}

func getHead(t *testing.T, node *html.Node) (h head) {
	for child := range node.ChildNodes() {
		if child.Type != html.ElementNode {
			continue
		}
		if strings.ToLower(child.Data) == "title" {
			if h.Title != "" {
				t.Error(getElementPath(node), "duplicate title tag")
			}
			for child := range child.ChildNodes() {
				if child.Type != html.TextNode {
					t.Error(getElementPath(node), "unexpected child node:", child.Data)
					continue
				}
				if h.Title != "" {
					t.Error(getElementPath(node), "duplicate title:", child.Data)
				}
				h.Title = child.Data
			}
		}
	}
	if h.Title == "" {
		t.Error("valid title tag was not found")
	}
	h.Meta = getHeadMeta(t, node)
	h.Description, _ = h.Meta.Data["description"]
	if h.Description == "" {
		t.Error("valid meta description was not found")
	}
	return h
}

func (r PageValidator) TestHead(node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		// TODO: implement UniqueConstraint(Validator) Validator
		// on title and description
		head := getHead(t, node)
		if head.Meta.CharacterSet == "" {
			t.Error("valid meta charset tag not found")
		}
		viewport, ok := head.Meta.Data["viewport"]
		if ok {
			r.TestViewPort(viewport)(t)
		} else {
			t.Error("head meta viewport definition is absent")
		}

		// TODO: if strict, OpenGraph and Twitter must both run
		for name, _ := range head.Meta.Data {
			if strings.HasPrefix(name, MetaTwitterPrefix) {
				t.Run("twitter:card", r.testTwitterCard(head.Meta))
				break
			}
		}
		for property, _ := range head.Meta.Properties {
			if strings.HasPrefix(property, MetaTwitterPrefix) {
				t.Run("og:card", r.testOpenGraphCard(head.Meta))
				break
			}
		}
	}
}
