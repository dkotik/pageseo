package pageseo

import (
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/dkotik/pageseo/internal"
	"golang.org/x/net/html"
)

type HeadNodeConstraints struct {
	Title       StringConstraints
	Description StringConstraints
	Keywords    StringConstraints
}

type head struct {
	Title       StringConstraints
	Description StringConstraints
	Keywords    StringConstraints
}

func NewHeadNodeTester(constraints HeadNodeConstraints) NodeTester {
	if constraints.Title.Normalizer == nil {
		constraints.Title.Normalizer = NormalizeTextToNFC
	}
	if constraints.Title.MinimumLength < 1 {
		constraints.Title.MinimumLength = DefaultMinimumTitleLength
	}
	if constraints.Title.MaximumLength < 1 {
		constraints.Title.MaximumLength = DefaultMaximumTitleLength
	}

	if constraints.Description.Normalizer == nil {
		constraints.Description.Normalizer = NormalizeTextToNFC
	}
	if constraints.Description.MinimumLength < 1 {
		constraints.Description.MinimumLength = DefaultMinimumDescriptionLength
	}
	if constraints.Description.MaximumLength < 1 {
		constraints.Description.MaximumLength = DefaultMaximumDescriptionLength
	}

	if constraints.Keywords.Normalizer == nil {
		constraints.Keywords.Normalizer = NormalizeTextToNFC
	}
	if constraints.Keywords.MaximumLength < 1 {
		constraints.Keywords.MaximumLength = DefaultMaximumKeywordsLength
	}

	return &head{
		Title:       constraints.Title,
		Description: constraints.Description,
		Keywords:    constraints.Keywords,
	}
}

func (h *head) Match(t testing.TB, node *html.Node) bool {
	switch node.Type {
	case html.ElementNode:
		return node.Data == "head"
	case html.DocumentNode:
		t.Cleanup(func() {
			var countOfHeadTags uint32
			for child := range node.Descendants() {
				if child.Type == html.ElementNode && child.Data == "head" {
					countOfHeadTags++
				}
			}
			switch countOfHeadTags {
			case 0:
				t.Fatal("document has no <head> node")
			case 1: // as required
			default:
				t.Logf(internal.WP+" document has %d extra <head> nodes", countOfHeadTags-1)
			}
		})
		return false
	default:
		return false
	}
}

func (h *head) ListResourcesForPreloading(origin *url.URL, node *html.Node) []string {
	return nil
}

func (h *head) TestNode(t testing.TB, origin *url.URL, node *html.Node, loader Loader) {
	title := ""
	characterSet := ""
	metaData := make(map[string]string)
	metaProperties := make(map[string]string)
	ok, hasTwitter := false, false

nextNode:
	for node := range node.ChildNodes() {
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

		switch node.Data {
		case "meta": // fallthrough
		case "title":
			if title != "" {
				t.Error("duplicate <title>:", internal.GetText(node))
			}
			title = internal.GetText(node)
			continue
		default:
			continue
		}

		name, property, content := "", "", ""
		for _, attr := range node.Attr {
			switch attr.Key {
			case "name":
				if name != "" {
					t.Error("<meta[name]>: duplicate value for", attr.Val)
				}
				name = strings.ToLower(attr.Val)
			case "property":
				if property != "" {
					t.Error("<meta[property]>: duplicate property value for", attr.Val)
				}
				property = strings.ToLower(attr.Val)
			case "content":
				if content != "" {
					t.Error("<meta[content]>: duplicate content value for", attr.Val)
				}
				content = attr.Val
			case "charset":
				if characterSet != "" {
					t.Error("<meta[charset]>: duplicate character set:", attr.Val)
				}
				characterSet = attr.Val
				if strings.ToLower(characterSet) != "utf-8" {
					t.Error("<meta[charset]>: document character set is not UTF-8:", characterSet)
				}
				continue nextNode
			case "http-equiv":
				if name != "" {
					t.Error("<meta[http-equiv]>: duplicate value for", attr.Val)
				}
				name = "http-equiv:" + strings.ToLower(attr.Val)
			}
		}
		content = strings.TrimSpace(content)
		if content == "" {
			t.Errorf("<meta[name=%q]>: has no content", name)
		} else {
			if name == "" {
				if property == "" {
					t.Errorf("<meta[content=%q]>: name attribute absent", content)
				} else {
					if strings.HasPrefix(property, MetaTwitterPrefix) {
						t.Log("<meta[property]> must be <meta[content]> for OpenGraph data")
					}
					if _, ok = metaProperties[property]; ok {
						t.Error("duplicate <meta[property]>:", property)
					}
					metaProperties[property] = content
				}
			} else {
				if property != "" {
					t.Logf("<meta[name=%q]>: name and property are both set", name)
					if _, ok = metaProperties[property]; ok {
						t.Errorf("<meta[property=%q]>: duplicate meta property", property)
					}
					metaProperties[property] = content
				}
				if _, ok = metaData[name]; ok {
					t.Errorf("<meta[name=%q]>: duplicate meta content", name)
				}
				if strings.HasPrefix(name, MetaTwitterPrefix) {
					hasTwitter = true
				}
				if strings.HasPrefix(property, MetaOpenGraphPrefix) {
					t.Log("<meta[content]> must be <meta[property]> for OpenGraph data")
				}
				metaData[name] = content
			}
		}
	}

	if characterSet == "" {
		t.Error("<head> meta charset tag is absent")
	}
	viewport, ok := metaData["viewport"]
	if ok {
		TestViewPort(t, viewport)
	} else {
		t.Error("<head> meta viewport definition is absent")
	}

	if title == "" {
		t.Error("head <title> is absent")
	} else {
		normalized, err := h.Title.Normalizer.Normalize(title)
		if err != nil {
			t.Log(internal.WP, "head <title> normalization error:", err)
		}
		if title != normalized {
			t.Log(internal.WP, "title text is not normalized")
		}
		length := len(title)
		if length == 0 {
			t.Error("head <title> text is empty")
		} else if length > h.Title.MaximumLength {
			t.Error("head <title> text is too long:", length, "vs", h.Title.MaximumLength)
		} else if length < h.Title.MinimumLength {
			t.Error("head <title> text is too short:", length, "vs", h.Title.MinimumLength)
		}
	}

	description, ok := metaData["description"]
	if !ok {
		t.Error("head <description> is absent")
	} else {
		normalized, err := h.Description.Normalizer.Normalize(description)
		if err != nil {
			t.Log(internal.WP, "head <description> normalization error:", err)
		}
		if description != normalized {
			t.Log(internal.WP, "description text is not normalized")
		}
		length := len(description)
		if length == 0 {
			t.Error("head <description> text is empty")
		} else if length > h.Description.MaximumLength {
			t.Error("head <description> text is too long:", length, "vs", h.Description.MaximumLength)
		} else if length < h.Description.MinimumLength {
			t.Error("head <description> text is too short:", length, "vs", h.Description.MinimumLength)
		}
	}

	keywords, ok := metaData["keywords"]
	if ok {
		normalized, err := h.Keywords.Normalizer.Normalize(keywords)
		if err != nil {
			t.Log(internal.WP, "head <keywords> normalization error:", err)
		}
		if keywords != normalized {
			t.Log(internal.WP, "keywords text is not normalized")
		}
		length := len(keywords)
		if length == 0 {
			t.Log("head <keywords> text is empty")
		} else if length > h.Keywords.MaximumLength {
			t.Log("head <keywords> text is too long:", length, "vs", h.Keywords.MaximumLength)
		} else if length < h.Keywords.MinimumLength {
			t.Log("head <keywords> text is too short:", length, "vs", h.Keywords.MinimumLength)
		}
	}

	// if hasOpenGraph {
	TestOpenGraphMeta(t, metaProperties, HeadNodeConstraints(*h))
	// } else {
	// 	t.Error("there is no open graph <head> meta data")
	// }
	if hasTwitter {
		TestTwitterMeta(t, metaData, HeadNodeConstraints(*h))
	} else {
		t.Log(internal.WP, "there is no Twitter (or `X`) <head> meta data")
	}
}

func TestViewPort(t testing.TB, content string) {
	if content == "" {
		t.Fatal("meta viewport is empty")
	}
	csv, err := internal.ParseCommaSeparatedKeyedValues(content)
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
