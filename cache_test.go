package pageseo

import (
	"context"
	"testing"
)

type mockLoader struct{}

func (ml mockLoader) Load(context.Context, string) ([]byte, string, error) {
	return []byte(`mock`), "text/html", nil
}

func TestResourceCache(t *testing.T) {
	called := 0
	cache := NewCache(func(r Resource) {
		called++
		t.Log("called cache:", r.URL)
	}).WrapLoader(mockLoader{})

	_, _, _ = cache.Load(nil, "string")
	_, _, _ = cache.Load(nil, "string")
	_, _, _ = cache.Load(nil, "string1")
	if called != 2 {
		t.Fatal("called cache too many times:", called)
	}
}
