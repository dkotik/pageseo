package main

import (
	"net/http"
	"time"

	"github.com/dkotik/pageseo"
)

var headers = http.Header{
	"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
	"Accept-Language": []string{"en-US,en;q=0.9"},
}

func newClientHTTP() (client *http.Client) {
	client = &http.Client{
		Timeout: 12 * time.Second,
	}
	return client
}

func newClientPool(depth uint8) pageseo.Loader {
	retry := pageseo.NewRetry(3)
	// delay := NewDelay(time.Second, time.Millisecond*700)
	switch depth {
	case 0:
		panic("zero clients")
	case 1:
		return retry.WrapLoader(pageseo.NewClient(newClientHTTP(), headers))
	}

	loaders := make([]pageseo.Loader, depth)
	for i := range depth {
		loaders[i] = retry.WrapLoader(
			// delay.WrapLoader(
			pageseo.NewClient(newClientHTTP(), headers),
			// ),
		)
	}
	return pageseo.NewSemaphore(loaders...)
}
