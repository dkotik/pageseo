package crawler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dkotik/pageseo"
)

var ErrNoMoreTargets = errors.New("there are no targets left")

type Target struct {
	Location       string
	ContentType    string
	Content        []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastAnalyzedAt time.Time
}

type Cache interface {
	pageseo.Loader
	GetNextTarget(context.Context) (Target, error)
	MarkAsAnalyzed(context.Context, Target) error
}

func TestCache(t *testing.T, c Cache) {
	if c == nil {
		t.Fatal("nil cache")
	}
}
