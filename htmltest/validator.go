package htmltest

import "fmt"

type skip struct{}

func (s skip) Validate(string) error {
	return nil
}

func (s skip) Wrap(v Validator) Validator {
	return v
}

var skipSingleton = skip{}

type Validator interface {
	Validate(string) error
}

type ValidatorFunc func(string) error

func (f ValidatorFunc) Validate(s string) error {
	return f(s)
}

// SkipValidator is a validator that never returns an error.
// It is a sentinel value used to skip validation tests and
// reduce the verbosity of test output.
var SkipValidator Validator = skipSingleton

// NewExactMatch creates a validator that checks if the input string matches the expected string.
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

// SkipMiddleware is a middleware that does nothing.
var SkipMiddleware Middleware = skipSingleton
