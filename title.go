package pageseo

import (
	"errors"
)

const (
	DefaultMinimumTitleLength = 4
	DefaultMaximumTitleLength = 55
)

func NewTitleValidator(s StringConstraints) Validator {
	if s.Normalizer == nil {
		s.Normalizer = NormalizeLine
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumTitleLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumTitleLength
	}
	return titleValidator(s)
}

type titleValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func (s titleValidator) Validate(value string) error {
	normalized, err := s.Normalizer(value)
	if err != nil {
		return err
	}
	if normalized != value {
		return errors.New("page title is not UTF normalized")
	}

	switch length := len(normalized); {
	case length == 0:
		return errors.New("page title is empty")
	case length < s.MinimumLength:
		return errors.New("page title is too short")
	case length > s.MaximumLength:
		return errors.New("page title is too long")
	default:
		return nil
	}
}
