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
	t.Helper()
	if c == nil {
		t.Fatal("nil cache")
	}
	ctx := context.Background()
	page, ct, err := c.Load(ctx, "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	if page == nil {
		t.Fatal("nil page")
	}
	if ct != "text/html" {
		t.Fatal("wrong content type:", ct)
	}

	target, err := c.GetNextTarget(ctx)
	if err != nil {
		if errors.Is(err, ErrNoMoreTargets) {
			t.Fatal("nothing returned, got error:", err)
		}
		t.Fatal(err)
	}
	if target.Location != "http://localhost/" {
		t.Fatal("wrong target location:", target.Location)
	}
	if target.ContentType != ct {
		t.Fatal("wrong content type:", target.ContentType)
	}

	if err = c.MarkAsAnalyzed(ctx, target); err != nil {
		t.Fatal(err)
	}

	target, err = c.GetNextTarget(ctx)
	if !errors.Is(err, ErrNoMoreTargets) {
		t.Log("location:", target.Location)
		t.Fatal("expected ErrNoMoreTargets, got error:", err)
	}
}
