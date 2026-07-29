package main

import (
	"net/http"
	"time"
)

func newClientHTTP() (client *http.Client) {
	client = &http.Client{
		Timeout: 12 * time.Second,
	}
	return client
}
