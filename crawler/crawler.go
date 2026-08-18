/*
Package crawler provides an HTTP [pageseo.Loader] that channels
unique [pageseo.Resource]s discovered during page validation.
*/
package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"

	"github.com/dkotik/pageseo"
)

type Crawler interface {
	IsBusy() bool
}

type crawler struct {
	ActiveTaskCount *atomic.Int32
	Filter          Filter
	Discovered      chan pageseo.Resource
	Logger          *slog.Logger
}

func New(withOptions ...Option) (_ Crawler, _ <-chan pageseo.Resource, err error) {
	o := options{}
	for _, option := range append(
		slices.Grow(withOptions, len(withOptions)+1),
		withDefaults(),
	) {
		o, err = option(o)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid crawler configuration: %w", err)
		}
	}
	c := &crawler{
		ActiveTaskCount: &atomic.Int32{},
		Discovered:      make(chan pageseo.Resource, 64),
		Logger:          o.Logger,
	}
	ch := make(chan pageseo.Resource)
	return c, ch, nil
}

func (c *crawler) Issue(r pageseo.Resource) {
	// TODO: check if unique first
	// TODO: store into cache
	select {
	case c.Discovered <- r:
		// resource was sent to the channel
	default:
		// channel was full, so spawn a goroutine to send the resource
		// later to prevent contention lock
		if c.ActiveTaskCount.Add(1) > 256 {
			c.Logger.Warn(
				"crawler is chocking on data",
				slog.String("state", "skipped"),
				slog.String("URL", r.URL),
			)
			c.ActiveTaskCount.Add(-1) // undo the add
			return
		}

		go func() {
			c.Discovered <- r
			c.ActiveTaskCount.Add(-1)
		}()
	}
}

func (c *crawler) IsBusy() bool {
	return c.ActiveTaskCount.Load() > 0 || len(c.Discovered) > 0
}

func (c *crawler) Load(ctx context.Context, URL string) pageseo.Resource {
	c.ActiveTaskCount.Add(1)
	defer c.ActiveTaskCount.Add(-1)
	// return nil, "", errors.New("not implemented")
	return pageseo.Resource{}
}

// func (c *crawler) Wait(ctx context.Context) error {
// 	select {
// 	case <-c.Closer:
// 		return nil
// 	case <-ctx.Done():
// 		return ctx.Err()
// 	}
// }
