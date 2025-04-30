//go:build popular

package pageseo

import (
	"io/fs"
	"testing"
)

func TestPopularPages(t *testing.T) {
	sub, err := fs.Sub(testData, "testdata/popular")
	if err != nil {
		t.Fatal("unable to load popular pages directory", err)
	}
	reqs := Requirements{}.WithDefaults()

	err = fs.WalkDir(sub, ".", fs.WalkDirFunc(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			f, err := sub.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			t.Run(path, reqs.TestReader(f))
		}
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
}
