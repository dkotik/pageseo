package pageseo

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"
)

func NewCrawler(
	validator *PageValidator,
	clients ...*http.Client,
) func(string) func(*testing.T) {
	retry := NewRetry(3)
	delay := NewDelay(time.Second, time.Millisecond*700)
	loaders := make([]Loader, len(clients))
	for i, client := range clients {
		if client == nil {
			panic("nil HTTP client")
		}
		loaders[i] = retry.WrapLoader(
			delay.WrapLoader(
				NewClient(client, http.Header{
					"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
					"Accept-Language": []string{"en-US,en;q=0.9"},
				}),
			),
		)
	}
	cache := make(map[string]cachedResource, 64)
	mu := &sync.Mutex{}
	semaphore := NewSemaphore(loaders...)
	return func(url string) func(*testing.T) {
		return func(t *testing.T) {
			// t.Parallel()
			var next []Resource
			loader := cachedLoader{
				mu:     mu,
				cache:  cache,
				Loader: semaphore,
				OnGrow: func(r Resource) {
					next = append(next, r)
				},
			}
			content, contentType, err := loader.Load(t.Context(), url)
			if err != nil {
				t.Errorf("unable to load <%s>: %v", url, err)
			}
			if contentType != "text/html" {
				t.Error("unexpected content type:", contentType)
			}
			if len(content) == 0 {
				t.Error("empty page content")
			}
			validator.TestReader(url, bytes.NewReader(content), loader)(t)

			for _, next := range next[1:] { // first item is the page itself
				t.Run(
					next.URL,
					validator.TestReader(next.URL, bytes.NewReader(next.Content), loader),
				)
			}
		}
	}
}
