package htmltest

import "fmt"

type Validator interface {
	Validate(string) error
}

type ValidatorFunc func(string) error

func (f ValidatorFunc) Validate(s string) error {
	return f(s)
}

func NewExactMatch(expected string) Validator {
	return ValidatorFunc(func(s string) error {
		if s != expected {
			return fmt.Errorf("expected %q, got %q", expected, s)
		}
		return nil
	})
}
