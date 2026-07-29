package pageseo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
)

type Loader interface {
	Load(context.Context, string) ([]byte, string, error)
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
