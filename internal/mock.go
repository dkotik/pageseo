package internal

import "context"

type mockLoader struct {
	Responder   func(string) (string, error)
	ContentType string
}

func NewMockLoader(
	r func(string) (string, error),
	contentType string,
) *mockLoader {
	if r == nil {
		r = func(s string) (string, error) { return s, nil }
	}
	if contentType == "" {
		contentType = "text/html"
	}
	return &mockLoader{
		Responder:   r,
		ContentType: contentType,
	}
}

func (ml mockLoader) Load(_ context.Context, URL string) ([]byte, string, error) {
	s, err := ml.Responder(URL)
	return []byte(s), ml.ContentType, err
}
