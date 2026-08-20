/*
Package crawler provides an HTTP [pageseo.Loader] that channels
unique [pageseo.Resource]s discovered during page validation.
*/
package crawler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/dkotik/pageseo"
	"github.com/dkotik/pageseo/crawler/repository"
	sqr "github.com/dkotik/pageseo/crawler/repository/sqlite"
)

type Analyzer interface {
	Analyze(context.Context, repository.Target) error
}

type AnalyzerFunc func(context.Context, repository.Target) error

func (f AnalyzerFunc) Analyze(ctx context.Context, t repository.Target) error {
	return f(ctx, t)
}

type Crawler interface {
	pageseo.Loader
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
				if o.SQLiteConn == nil {
					return o, errors.New("SQLite connection is required when a repository is not provided")
				}
				loader := newClientPool(8)
				if o.Delay != 0 || o.DelayFluctuate != 0 {
					loader = pageseo.NewDelay(o.Delay, o.DelayFluctuate).WrapLoader(loader)
				}
				o.Repository, err = sqr.New(
					o.SQLiteConn,
					loader,
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

func (c *crawler) Load(ctx context.Context, URL string) ([]byte, string, error) {
	return c.Repository.Load(ctx, URL)
}
