package crawler

import (
	"errors"
	"log/slog"
	"slices"
	"strings"

	"github.com/dkotik/pageseo"
)

type Filter interface {
	IsAllowed(pageseo.Resource) bool
}

type FilterFunc func(pageseo.Resource) bool

func (f FilterFunc) IsAllowed(r pageseo.Resource) bool {
	return f(r)
}

type options struct {
	Targets []string
	Filter  Filter
	Logger  *slog.Logger
}

type Option func(options) (options, error)

func WithTargets(URLs ...string) Option {
	return func(o options) (options, error) {
		o.Targets = slices.Grow(o.Targets, len(URLs))
		for _, URL := range URLs {
			URL = strings.TrimSpace(URL)
			if URL == "" {
				return o, errors.New("empty URL provided as target")
			}
			o.Targets = append(o.Targets, URL)
		}
		return o, nil
	}
}

func WithFilter(filter Filter) Option {
	return func(o options) (options, error) {
		if filter == nil {
			return o, errors.New("nil filter")
		}
		if o.Filter != nil {
			return o, errors.New("filter is already set")
		}
		o.Filter = filter
		return o, nil
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(o options) (options, error) {
		if logger == nil {
			return o, errors.New("nil logger")
		}
		if o.Logger != nil {
			return o, errors.New("logger is already set")
		}
		o.Logger = logger
		return o, nil
	}
}

func withDefaults() Option {
	return func(o options) (_ options, err error) {
		if o.Filter == nil {
			if len(o.Targets) == 0 {
				o.Filter = FilterFunc(func(r pageseo.Resource) bool {
					return true
				})
			} else {
				o.Filter, err = NewTargetFilter(o.Targets...)
				if err != nil {
					return o, err
				}
			}
		}
		if o.Logger == nil {
			o.Logger = slog.Default()
		}
		return o, nil
	}
}
