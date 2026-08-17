package pageseo

import (
	"testing"

	"github.com/dkotik/pageseo/internal"
)

func TestResourceCache(t *testing.T) {
	called := 0
	cache := NewCache(func(r Resource) {
		called++
		t.Log("called cache:", r.URL)
	}).WrapLoader(internal.NewMockLoader(nil, ""))

	_, _, _ = cache.Load(nil, "string")
	_, _, _ = cache.Load(nil, "string")
	_, _, _ = cache.Load(nil, "string1")
	if called != 2 {
		t.Fatal("called cache too many times:", called)
	}
}
