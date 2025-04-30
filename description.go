package pageseo

import (
	"errors"

	"github.com/dkotik/pageseo/htmltest"
)

const (
	DefaultMinimumDescriptionLength = 4
	DefaultMaximumDescriptionLength = 150
)

func NewDescriptionValidator(s StringConstraints) htmltest.Validator {
	if s.Normalizer == nil {
		s.Normalizer = NormalizeTextToNFC
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumDescriptionLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumDescriptionLength
	}
	return descriptionValidator(s)
}

type descriptionValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func (d descriptionValidator) Validate(value string) error {
	normalized, err := d.Normalizer.Normalize(value)
	if err != nil {
		return err
	}
	if normalized != value {
		return errors.New("page description is not UTF normalized")
	}

	switch length := len(normalized); {
	case length == 0:
		return errors.New("page description is empty")
	case length < d.MinimumLength:
		return errors.New("page description is too short")
	case length > d.MaximumLength:
		return errors.New("page description is too long")
	default:
		return nil
	}
}
