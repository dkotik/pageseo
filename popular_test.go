//go:build golden

package pageseo

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestPopularPages(t *testing.T) {
	popular := os.DirFS(filepath.Join("testdata", "popular"))
	files, err := fs.ReadDir(popular, ".")
	pageSEO := New(nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		contents, err := fs.ReadFile(popular, file.Name())
		if err != nil {
			t.Fatal(err)
		}
		if len(contents) == 0 {
			t.Fatalf("file %s is empty", file.Name())
		}
		t.Run(file.Name(), pageSEO.TestPage(file.Name(), bytes.NewReader(contents)))
	}
	// t.Fail()
}
