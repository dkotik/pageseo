package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/dkotik/pageseo"
)

type loaderHTTP struct {
	*http.Client
	Headers http.Header
}

func NewClient(client *http.Client, headers http.Header) pageseo.Loader {
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
