package pageseo

import (
	"bytes"
	"os"
	"testing"
)

var testData = NewFS(os.DirFS("testdata"))

func TestMinimalPage(t *testing.T) {
	f, _, err := testData.Load(t.Context(), "minimal.html")
	if err != nil {
		t.Fatal(err)
	}

	NewStrict(testData, Requirements{
		// LinkText: htmltest.SkipValidator,
		// LinkText: NewLinkTextValidator(StringConstraints{
		// 	MinimumLength: 1,
		// 	MaximumLength: 100,
		// }),
	}).TestReader(t.Name(), bytes.NewReader(f))(t)
}
