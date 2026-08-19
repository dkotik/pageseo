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
	"time"

	"github.com/dkotik/pageseo/crawler/repository"
	sqr "github.com/dkotik/pageseo/crawler/repository/sqlite"
	"zombiezen.com/go/sqlite"
)

type Analyzer interface {
	Analyze(context.Context, repository.Target) error
}

type AnalyzerFunc func(context.Context, repository.Target) error

func (f AnalyzerFunc) Analyze(ctx context.Context, t repository.Target) error {
	return f(ctx, t)
}

type Crawler interface {
	CrawlLocation(context.Context, string) error
}

type crawler struct {
	Analyzer   Analyzer
	Repository repository.Repository
	BatchSize  int
	TimeToLive time.Duration
	Logger     *slog.Logger
}

func New(analyzer Analyzer, withOptions ...Option) (_ Crawler, err error) {
	o := options{}
	for _, option := range append(
		slices.Grow(withOptions, len(withOptions)+1),
		func(o options) (_ options, err error) {
			if o.TimeToLive == 0 {
				o.TimeToLive = 5 * time.Minute
			}
			if o.BatchSize == 0 {
				o.BatchSize = 5
			}
			if o.Logger == nil {
				o, err = WithLogger(slog.New(slog.DiscardHandler))(o)
				if err != nil {
					return o, err
				}
			}

			if o.Repository == nil {
				conn, err := sqlite.OpenConn(":memory:")
				if err != nil {
					return o, err
				}
				o.Repository, err = sqr.New(
					conn,
					newClientPool(8),
					"pageseo_targets",
					o.TimeToLive,
				)
				if err != nil {
					return o, err
				}
			}
			return o, nil
		},
	) {
		o, err = option(o)
		if err != nil {
			return nil, fmt.Errorf("invalid crawler configuration: %w", err)
		}
	}
	c := &crawler{
		Analyzer:   analyzer,
		Repository: o.Repository,
		BatchSize:  o.BatchSize,
		TimeToLive: o.TimeToLive,
		Logger:     o.Logger,
	}
	return c, nil
}
