package pageseo

import (
	"bytes"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

func NewCrawler(
	validator PageValidator,
	clients ...*http.Client,
) func(string) func(*testing.T) {
	retry := NewRetry(3)
	// delay := NewDelay(time.Second, time.Millisecond*700)
	loaders := make([]Loader, len(clients))
	for i, client := range clients {
		if client == nil {
			panic("nil HTTP client")
		}
		loaders[i] = retry.WrapLoader(
			// delay.WrapLoader(
			NewClient(client, http.Header{
				"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
				"Accept-Language": []string{"en-US,en;q=0.9"},
			}),
			// ),
		)
	}
	cache := make(map[string]cachedResource, 64)
	mu := &sync.Mutex{}
	semaphore := NewSemaphore(loaders...)
	return func(origin string) func(*testing.T) {
		return func(t *testing.T) {
			_, err := url.Parse(origin)
			if err != nil {
				t.Fatalf("invalid URL <%s>: %v", origin, err)
			}

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
			content, contentType, err := loader.Load(t.Context(), origin)
			if err != nil {
				t.Errorf("unable to load <%s>: %v", origin, err)
			}
			if contentType != "text/html" {
				t.Error("unexpected content type:", contentType)
			}
			if len(content) == 0 {
				t.Error("empty page content")
			}
			validator := validator // copy
			validator.Loader = loader
			validator.TestReader(origin, bytes.NewReader(content))(t)

			// limit := 100
			// next = next[1:] // first item is the page itself
			// for {
			// 	if len(next) == 0 {
			// 		break
			// 	}
			// 	batch := slices.Clone(next)
			// 	next = next[:0]
			// 	for _, next := range batch {
			// 		limit--
			// 		if limit == 0 {
			// 			t.Error("100 attempts ran out")
			// 		}
			// 		nextTarget, err := url.Parse(next.URL)
			// 		if err != nil {
			// 			t.Errorf("invalid URL <%s>: %v", next.URL, err)
			// 			continue
			// 		}
			// 		if nextTarget.Host != root.Host {
			// 			continue // external URL, outside of this domain
			// 		}
			// 		t.Run(
			// 			strings.TrimPrefix(nextTarget.Path, "/"),
			// 			validator.TestReader(
			// 				next.URL,
			// 				bytes.NewReader(next.Content),
			// 			),
			// 		)
			// 	}
			// }
		}
	}
}
