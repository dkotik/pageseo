package htmltest

import (
	"bytes"
	"fmt"
	"testing"

	"golang.org/x/net/html"
)

func TestTreeMatching(t *testing.T) {
	sample := []byte(`<body><div class="content"><p class="title">Hello</p></div></body>`)
	parsed, err := html.Parse(bytes.NewReader(sample))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	tree := Node{
		Name: "body",
		Children: []Node{
			{
				Name: "div",
				Attributes: map[string]Validator{
					"class": ValidatorFunc(func(s string) error {
						if s != "content" {
							return fmt.Errorf("expected 'content', got '%s'", s)
						}
						return nil
					}),
				},
				Children: []Node{
					{
						Name: "p",
						Attributes: map[string]Validator{
							"class": ValidatorFunc(func(s string) error {
								if s != "title" {
									return fmt.Errorf("expected 'title', got '%s'", s)
								}
								return nil
							}),
						},
						Children: []Node{
							{
								Content: ValidatorFunc(func(s string) error {
									if s != "Hello" {
										return fmt.Errorf("expected 'Hello', got '%s'", s)
									}
									return nil
								}),
							},
						},
					},
				},
			},
		},
	}

	if err := tree.Match(parsed, nil); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}
