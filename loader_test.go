package pageseo

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

type mockLoader struct {
	Prefix string
}

func (ml mockLoader) Load(_ context.Context, URL string) ([]byte, string, error) {
	return []byte(`<html>` + ml.Prefix + URL + `</html>`), "text/html", nil
}

func TestPreloader(t *testing.T) {
	delay := time.Second * 3
	preloader := newPreloader(NewDelay(delay, 0).WrapLoader(mockLoader{}))
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		preloader.Preload(ctx, "1")
		<-time.After(delay + (time.Millisecond * 300))
		preloader.Preload(ctx, "2")

		at := time.Now()
		first, _, err := preloader.Load(ctx, "1")
		if err != nil {
			t.Fatal("unable to load the first item:", string(first))
		}
		if time.Since(at) > time.Millisecond {
			t.Fatal("loading the first item took too long")
		}

		// <-time.After(delay + (time.Millisecond * 300)) // fails the test
		at = time.Now()
		second, _, err := preloader.Load(ctx, "2")
		if err != nil {
			t.Fatal("unable to load the second item:", string(second))
		}
		if time.Since(at) < time.Millisecond {
			t.Fatal("loading the second item took too short")
		}
	})
}
