package pageseo

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/net/html"
)

func TestGetMeta(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join("testdata", "minimal.html"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := html.Parse(bytes.NewReader(doc))
	if err != nil {
		t.Fatal(err)
	}
	head := findElementNode(tree, "head")
	if head == nil {
		t.Fatal("no head element found")
	}
	meta := getHeadMeta(t, head)
	if meta.CharacterSet != "utf-8" {
		t.Fatal("there is no character set")
	}

	description, ok := meta.Data["description"]
	if !ok {
		t.Error("there is no description meta")
	}
	if description != "lorem ipsum" {
		t.Error("description does not match:", description)
	}

	description, ok = meta.Properties["og:description"]
	if !ok {
		t.Error("there is no og:description meta")
	}
	if description != "Lorem Ipsum Dolor" {
		t.Error("description does not match:", description)
	}
}
