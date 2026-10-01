package pageseo

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
)

type gzLoader struct {
	Loader
}

// NewGzip wraps a loader and transparently decompresses
// gzip-encoded resources. The content type of the
// decompressed payload is detected from its contents.
func NewGzip(loader Loader) Loader {
	if loader == nil {
		panic("nil loader")
	}
	return gzLoader{Loader: loader}
}

func (g gzLoader) Load(ctx context.Context, url string) (data []byte, contentType string, err error) {
	data, contentType, err = g.Loader.Load(ctx, url)
	if err != nil {
		return data, contentType, err
	}
	switch contentType {
	case "application/gzip", "application/x-gzip":
	default:
		return data, contentType, nil
	}

	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("unable to decompress <%s>: %w", url, err)
	}
	defer func() {
		err = errors.Join(err, r.Close())
	}()
	data, err = io.ReadAll(r)
	if err != nil {
		return nil, "", fmt.Errorf("unable to decompress <%s>: %w", url, err)
	}
	contentType, _, err = mime.ParseMediaType(http.DetectContentType(data))
	if err != nil {
		return nil, "", fmt.Errorf("unable to parse media type of decompressed <%s>: %w", url, err)
	}
	return data, contentType, nil
}
