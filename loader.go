package pageseo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/html"
)

// Skip is a sentinel error that indicates a page resource
// should not be validated.
var Skip = errors.New("do not validate this resource")

// Loader fetches the resource from a given location.
// Implementations should handle caching and other
// performance optimizations.
//
// Loader should return [Skip] if a resource
// should not be validated without raising an error
// or a warning.
type Loader interface {
	Load(context.Context, string) ([]byte, string, error)
}

type skipAllLoader struct{}

var skipAllLoaderSingleton Loader = skipAllLoader{}

func (skipAllLoader) Load(context.Context, string) ([]byte, string, error) {
	return nil, "", Skip
}

type Resource struct {
	URL         string
	ContentType string
	Content     []byte
	Error       error
}

func (r PageValidator) preloadResources(
	t *testing.T,
	origin *url.URL,
	body *html.Node,
) Loader {
	URLs := make([]string, 0, 16)
	addURL := func(URL string) {
		relative, _, err := r.cachedURLs.GetRelativeTo(origin, URL)
		if err == nil {
			URLs = append(URLs, relative.String())
		} else {
			t.Errorf("invalid URL <%s>: %v", URL, err)
		}
	}

	for node := range body.Descendants() {
		if node.Type != html.ElementNode {
			continue
		}
		switch node.Data {
		case "a":
			href, ok := getAttribute(node, `href`)
			if ok {
				addURL(href)
			}
		case "img":
			for _, src := range GetPictureSourceList(node.Parent) {
				addURL(src)
			}
			fallthrough
		case "script", "style":
			src, ok := getAttribute(node, `src`)
			if ok {
				addURL(src)
			}
		}
	}
	return NewHotSwap(t.Context(), r.Loader, URLs)
}

// TODO: deprecate in favor of hotSwapLoader
type preloader struct {
	Loader
	Mutex        sync.Mutex
	Cursor       int
	PendingTasks []string
	Preloaded    []Resource
}

func newPreloader(loader Loader) *preloader {
	if loader == nil {
		panic("nil loader")
	}
	return &preloader{
		Loader: loader,
	}
}

func (p *preloader) Preload(ctx context.Context, URL string) {
	go func(ctx context.Context) {
		p.Mutex.Lock()
		if slices.Index(p.PendingTasks, URL) != -1 {
			p.Mutex.Unlock()
			return
		}
		p.PendingTasks = append(p.PendingTasks, URL)
		p.Mutex.Unlock()

		content, contentType, err := p.Loader.Load(ctx, URL)

		p.Mutex.Lock()
		p.Preloaded = append(
			p.Preloaded,
			Resource{
				URL:         URL,
				Content:     content,
				ContentType: contentType,
				Error:       err,
			})
		p.PendingTasks = slices.Delete(p.PendingTasks, slices.Index(p.PendingTasks, URL), 1)
		p.Mutex.Unlock()
	}(ctx)
}

func (p *preloader) Load(ctx context.Context, URL string) ([]byte, string, error) {
	var i, pending int
	var r Resource
	for {
		p.Mutex.Lock()
		// search forward from cursor
		for i = p.Cursor; i < len(p.Preloaded); i++ {
			r = p.Preloaded[i]
			if r.URL == URL {
				p.Cursor = i + 1 // next time begin iteration from same point
				p.Mutex.Unlock()
				return r.Content, r.ContentType, nil
			}
		}
		// search backward from cursor
		for i = p.Cursor - 1; i >= 0; i-- {
			r = p.Preloaded[i]
			if r.URL == URL {
				p.Mutex.Unlock()
				return r.Content, r.ContentType, nil
			}
		}
		pending = len(p.PendingTasks)
		p.Mutex.Unlock()

		if pending == 0 {
			// fallback on slow loader
			return p.Loader.Load(ctx, URL)
		}

		select { // try again after a short break
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(time.Millisecond * 300):
		}
	}
}

type fsLoader struct {
	fs.FS
}

func NewFS(fs fs.FS) Loader {
	if fs == nil {
		panic("nil file system")
	}
	return fsLoader{
		FS: fs,
	}
}

func (fs fsLoader) Load(_ context.Context, url string) ([]byte, string, error) {
	r, err := fs.FS.Open(url)
	if err != nil {
		return nil, "", fmt.Errorf("unable to open <%s>: %w", url, err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", fmt.Errorf("unable to load <%s>: %w", url, err)
	}
	return data, http.DetectContentType(data), nil
}

type semaphoreLoader struct {
	Loaders chan Loader
}

func NewSemaphore(loaders ...Loader) Loader {
	total := len(loaders)
	switch total {
	case 1:
		if loaders[0] == nil {
			panic("nil loader in the semaphore")
		}
		return loaders[0]
	case 0:
		panic("no semaphore loaders")
	}
	stack := make(chan Loader, total)
	for _, loader := range loaders {
		if loader == nil {
			panic("nil loader in the semaphore")
		}
		stack <- loader
	}

	return semaphoreLoader{
		Loaders: stack,
	}
}

func (s semaphoreLoader) Load(ctx context.Context, url string) (data []byte, contentType string, err error) {
	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	case loader := <-s.Loaders:
		data, contentType, err = loader.Load(ctx, url)
		s.Loaders <- loader
		return
	}
}

type hotSwapLoader struct {
	Cursor    int
	Preloaded []Resource
	Loader    Loader
}

func NewHotSwap(ctx context.Context, loader Loader, URLs []string) Loader {
	resources := make([]Resource, len(URLs))
	wg := sync.WaitGroup{}
	for i, url := range URLs {
		wg.Add(1)
		go func(ctx context.Context, i int, url string) {
			data, contentType, err := loader.Load(ctx, url)
			resources[i] = Resource{
				URL:         url,
				ContentType: contentType,
				Content:     data,
				Error:       err,
			}
			wg.Done()
		}(ctx, i, url)
	}
	wg.Wait()
	return hotSwapLoader{
		Loader:    loader,
		Cursor:    0,
		Preloaded: resources,
	}
}

func (h hotSwapLoader) Load(ctx context.Context, URL string) ([]byte, string, error) {
	var i int

	// search forward from cursor
	for i = h.Cursor; i < len(h.Preloaded); i++ {
		r := h.Preloaded[i]
		if r.URL == URL {
			h.Cursor = i + 1 // next time begin iteration from same point
			return r.Content, r.ContentType, r.Error
		}
	}

	// search backward from cursor
	for i = h.Cursor - 1; i >= 0; i-- {
		r := h.Preloaded[i]
		if r.URL == URL {
			return r.Content, r.ContentType, r.Error
		}
	}

	// fallback on loader
	return h.Loader.Load(ctx, URL)
}
