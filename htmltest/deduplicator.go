package htmltest

import (
	"errors"
	"sync"
)

var (
	ErrDuplicateValue = errors.New("duplicate value")
)

type deduplicator struct {
	nameSpace string
	next      Validator
	mu        *sync.Mutex
	known     map[string]struct{}
}

func (d *deduplicator) Validate(key string) error {
	key = d.nameSpace + key
	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.known[key]
	if !ok {
		d.known[key] = struct{}{}
		return d.next.Validate(key)
	}
	return ErrDuplicateValue
}

type deduplicatorMiddleware struct {
	nameSpace string
	mu        *sync.Mutex
	known     map[string]struct{}
}

func (d *deduplicatorMiddleware) Wrap(next Validator) Validator {
	return &deduplicator{
		nameSpace: d.nameSpace,
		next:      next,
		mu:        &sync.Mutex{},
		known:     make(map[string]struct{}),
	}
}

func NewDeduplicator(nameSpace string) Middleware {
	return &deduplicatorMiddleware{
		nameSpace: nameSpace,
		mu:        &sync.Mutex{},
		known:     make(map[string]struct{}),
	}
}
