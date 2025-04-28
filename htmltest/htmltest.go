/*
Package htmltest provides a set of tools for testing rendered HTML page elements.
*/
package htmltest

import (
	"errors"
	"fmt"
	"iter"
	"path"
	"strings"

	"golang.org/x/net/html"
)

type MatchError struct {
	Path  []string
	Cause error
}

func NewMatchError(path []string, cause error) error {
	if cause == nil {
		return nil
	}
	return MatchError{
		Path:  path,
		Cause: cause,
	}
}

func (e MatchError) Unwrap() error {
	return e.Cause
}

func (e MatchError) Error() string {
	p := path.Join(e.Path...)
	if p == "" {
		p = "/"
	}
	return fmt.Sprintf("HTML %q node does not match test specification: %v", p, e.Cause)
}

func IsPadding(n *html.Node) bool {
	if n == nil {
		return false
	}
	if n.Type != html.TextNode {
		return false
	}
	return strings.TrimSpace(n.Data) == ""
}

type Node struct {
	Name       string
	Content    Validator
	Attributes map[string]Validator
	Children   []Node
}

func (n Node) Match(x *html.Node, path []string) (err error) {
	if x == nil {
		return fmt.Errorf("node is nil")
	}

	switch x.Type {
	case html.TextNode, html.CommentNode:
		content := ParseTextContent(x)
		if n.Content != nil {
			return NewMatchError(path, n.Content.Validate(content))
		}
		return nil
	case html.DocumentNode:
		if x.FirstChild == nil {
			return MatchError{
				Path:  path,
				Cause: errors.New("document node has no children"),
			}
		}
		if x.FirstChild.FirstChild == nil {
			return MatchError{
				Path:  path,
				Cause: errors.New("<HTML> node has no children"),
			}
		}
		// Skip the doctype node, the html node, and the head node
		return n.Match(x.FirstChild.FirstChild.NextSibling, path)
	default:
		path = append(path, x.Data)
		if x.Data != n.Name {
			return MatchError{
				Path:  path,
				Cause: fmt.Errorf("expected tag name %q does not match the discovered tag name: %q", n.Name, x.Data),
			}
		}
	}

	attributes, err := ParseAttributes(x)
	if err != nil {
		return err
	}
	for attribute, validator := range n.Attributes {
		value, ok := attributes[attribute]
		if !ok {
			return MatchError{
				Path:  path,
				Cause: fmt.Errorf("expected attribute %q does not exist", attribute),
			}
		}
		if err := validator.Validate(value); err != nil {
			return MatchError{
				Path:  path,
				Cause: fmt.Errorf("attribute %q is invalid: %w", attribute, err),
			}
		}
	}

	children, closer := iter.Pull[*html.Node](x.ChildNodes())
	defer closer()
	for i, childConstraint := range n.Children {
		child, ok := children()
		if !ok {
			return MatchError{
				Path:  path,
				Cause: fmt.Errorf("expected child node #%d %q does not exist", i, childConstraint.Name),
			}
		}
		if IsPadding(child) {
			continue
		}
		if err := childConstraint.Match(child, append(path, n.Name)); err != nil {
			return MatchError{
				Path:  append(path, childConstraint.Name),
				Cause: fmt.Errorf("child node #%d %q is invalid: %w", i, childConstraint.Name, err),
			}
		}
	}
	return nil
}
