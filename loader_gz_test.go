package pageseo

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"testing"
)

func TestGzipLoader(t *testing.T) {
	html := []byte("<!DOCTYPE html><html><body><p>hello</p></body></html>")
	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?><urlset></urlset>`)
	sentinel := errors.New("boom")

	cases := []struct {
		Name            string
		Inner           staticLoader
		WantData        []byte
		WantContentType string
		WantErr         error
		WantAnyErr      bool
	}{
		{
			Name:            "passes through non-gzip content",
			Inner:           staticLoader{Data: html, ContentType: "text/html"},
			WantData:        html,
			WantContentType: "text/html",
		},
		{
			Name:            "decompresses application/gzip into html",
			Inner:           staticLoader{Data: gzipBytes(t, html), ContentType: "application/gzip"},
			WantData:        html,
			WantContentType: "text/html",
		},
		{
			Name:            "decompresses application/x-gzip into xml",
			Inner:           staticLoader{Data: gzipBytes(t, xml), ContentType: "application/x-gzip"},
			WantData:        xml,
			WantContentType: "text/xml",
		},
		{
			Name:    "propagates wrapped loader error",
			Inner:   staticLoader{Err: sentinel},
			WantErr: sentinel,
		},
		{
			Name:    "propagates Skip",
			Inner:   staticLoader{Err: Skip},
			WantErr: Skip,
		},
		{
			Name:       "fails on corrupt gzip payload",
			Inner:      staticLoader{Data: []byte("not gzip"), ContentType: "application/gzip"},
			WantAnyErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			data, ct, err := NewGzip(tc.Inner).Load(context.Background(), "test")
			switch {
			case tc.WantErr != nil:
				if !errors.Is(err, tc.WantErr) {
					t.Fatalf("expected error %v, got %v", tc.WantErr, err)
				}
				return
			case tc.WantAnyErr:
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			case err != nil:
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(data, tc.WantData) {
				t.Errorf("data mismatch:\n got: %q\nwant: %q", data, tc.WantData)
			}
			if ct != tc.WantContentType {
				t.Errorf("content type: got %q, want %q", ct, tc.WantContentType)
			}
		})
	}
}

type staticLoader struct {
	Data        []byte
	ContentType string
	Err         error
}

func (s staticLoader) Load(context.Context, string) ([]byte, string, error) {
	return s.Data, s.ContentType, s.Err
}

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
