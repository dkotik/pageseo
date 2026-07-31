package pageseo

import "testing"

func TestNor(t *testing.T) {
	good := []string{
		"http://example.com",
		"https://example.com",
		"www.cnn.com",
		"www.wikipedia.org/sdf/sdf/sdf/sdf.html?sdfkjl=43dsf&sdf=dfsd#hash",
	}

	validator := NewURLValidator(StringConstraints{})
	for _, url := range good {
		if err := validator.Validate(url); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}
}
