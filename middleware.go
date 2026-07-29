package pageseo

import (
	"context"
	"math/rand/v2"
	"time"
)

type Middleware interface {
	WrapLoader(Loader) Loader
}

type MiddlewareFunc func(Loader) Loader

func (mf MiddlewareFunc) WrapLoader(l Loader) Loader {
	return mf(l)
}

type retryLoader struct {
	AttemptLimit uint8
	Loader       Loader
}

func NewRetry(limit uint8) Middleware {
	if limit < 2 {
		panic("retry middleware requires at least two attempts")
	}
	return MiddlewareFunc(func(l Loader) Loader {
		return retryLoader{
			AttemptLimit: limit,
			Loader:       l,
		}
	})
}

func (r retryLoader) Load(ctx context.Context, url string) (data []byte, ct string, err error) {
	for range r.AttemptLimit {
		data, ct, err = r.Loader.Load(ctx, url)
		if err == nil {
			return data, ct, nil
		}
	}
	return data, ct, err
}

type delayLoader struct {
	Base   time.Duration
	Random time.Duration
	Loader Loader
}

func NewDelay(base, random time.Duration) Middleware {
	return MiddlewareFunc(func(l Loader) Loader {
		return delayLoader{
			Base:   base,
			Random: random,
			Loader: l,
		}
	})
}

func (d delayLoader) Load(ctx context.Context, url string) (data []byte, ct string, err error) {
	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	case <-time.After(d.Base + time.Duration(rand.NormFloat64()*float64(d.Random))):
		return d.Loader.Load(ctx, url)
	}
}
