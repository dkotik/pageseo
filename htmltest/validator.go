package htmltest

import "fmt"

type Validator interface {
	Validate(string) error
}

type ValidatorFunc func(string) error

func (f ValidatorFunc) Validate(s string) error {
	return f(s)
}

// SkipValidator is a validator that never returns an error.
// It is a sentinel value used to skip validation tests.
var SkipValidator Validator = ValidatorFunc(func(s string) error {
	return nil
})

func NewExactMatch(expected string) Validator {
	return ValidatorFunc(func(s string) error {
		if s != expected {
			return fmt.Errorf("expected %q, got %q", expected, s)
		}
		return nil
	})
}

type Middleware interface {
	Wrap(Validator) Validator
}

type MiddlewareFunc func(Validator) Validator

func (f MiddlewareFunc) Wrap(v Validator) Validator {
	return f(v)
}

var SkipMiddleware Middleware = MiddlewareFunc(func(v Validator) Validator {
	return v
})
