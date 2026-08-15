package crawler

import (
	"context"

	"github.com/dkotik/pageseo"
	"golang.org/x/sync/singleflight"
)

type singleFlightLoader struct {
	pageseo.Loader
	*singleflight.Group
}

func NewSingleFlightLoader(loader pageseo.Loader) *singleFlightLoader {
	if loader == nil {
		panic("nil loader")
	}
	return &singleFlightLoader{
		Loader: loader,
		Group:  &singleflight.Group{},
	}
}

func (l *singleFlightLoader) Load(ctx context.Context, URL string) ([]byte, string, error) {
	result, err, _ := l.Group.Do(URL, func() (any, error) {
		data, ct, err := l.Loader.Load(ctx, URL)
		return pageseo.Resource{
			URL:         URL,
			ContentType: ct,
			Content:     data,
		}, err
	})
	r := result.(pageseo.Resource)
	return r.Content, r.ContentType, err
}
