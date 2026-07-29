package pageseo

import (
	"context"
	"sync"
)

type Resource struct {
	URL         string
	ContentType string
	Content     []byte
	Error       error
}

type cachedResource struct {
	Content     []byte
	ContentType string
	Error       error
}

type cachedLoader struct {
	Loader Loader
	OnGrow func(Resource)

	mu    *sync.Mutex
	cache map[string]cachedResource
}

func NewCache(onGrow func(Resource)) Middleware {
	cache := make(map[string]cachedResource, 64)
	mu := &sync.Mutex{}
	return MiddlewareFunc(func(l Loader) Loader {
		return cachedLoader{
			Loader: l,
			OnGrow: onGrow,
			mu:     mu,
			cache:  cache,
		}
	})
}

func (cl cachedLoader) Load(ctx context.Context, url string) ([]byte, string, error) {
	cl.mu.Lock()
	cached, ok := cl.cache[url]
	cl.mu.Unlock()
	if ok {
		return cached.Content, cached.ContentType, cached.Error
	}
	fresh, freshContentType, err := cl.Loader.Load(ctx, url)
	cl.mu.Lock()
	_, ok = cl.cache[url] // check if resource maybe loaded in parallel
	cl.cache[url] = cachedResource{
		Content:     fresh,
		ContentType: freshContentType,
		Error:       err,
	}
	cl.mu.Unlock()
	if !ok { // true unique
		cl.OnGrow(Resource{
			URL:         url,
			Content:     fresh,
			ContentType: freshContentType,
			Error:       err,
		})
	}
	return fresh, freshContentType, err
}
