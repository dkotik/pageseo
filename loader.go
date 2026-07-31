package pageseo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"slices"
	"sync"
	"time"
)

type Loader interface {
	Load(context.Context, string) ([]byte, string, error)
}

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

type loaderHTTP struct {
	*http.Client
	Headers http.Header
}

func NewClient(client *http.Client, headers http.Header) Loader {
	if client == nil {
		panic("nil HTTP client")
	}
	return loaderHTTP{
		Client:  client,
		Headers: headers,
	}
}

func (web loaderHTTP) Load(ctx context.Context, url string) (data []byte, contentType string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("unable to open <%s>: %w", url, err)
	}
	for key, values := range web.Headers {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}
	resp, err := web.Client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("unable to load <%s>: %w", url, err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("unable to load <%s>: %w", url, err)
	}
	contentTypeRaw := resp.Header.Get(`Content-Type`)
	contentType, _, err = mime.ParseMediaType(contentTypeRaw)
	if err != nil {
		return nil, "", fmt.Errorf("unable to parse header <Content-Type> <%s>: %w", contentTypeRaw, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = fmt.Errorf("HTTP %d error: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	// if contentType == "" {
	// 	contentType = "application/octet-stream"
	// }
	return data, contentType, err
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
