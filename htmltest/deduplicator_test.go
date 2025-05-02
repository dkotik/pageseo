package htmltest

import (
	"errors"
	"testing"
)

func TestDeduplication(t *testing.T) {
	dd := NewDeduplicator("test").Wrap(
		ValidatorFunc(func(s string) error {
			return nil
		}),
	)

	err := dd.Validate("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = dd.Validate("test")
	if !errors.Is(err, ErrDuplicateValue) {
		t.Errorf("Expected ErrDuplicateValue error, got %T instead: %v", err, err)
	}
}
