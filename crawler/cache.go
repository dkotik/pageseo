package crawler

import (
	"context"
	"testing"
	"time"

	"github.com/dkotik/pageseo"
)

type Target struct {
	ID             int64
	Location       string
	ContentType    string
	Content        []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastAnalyzedAt time.Time
}

type Cursor struct {
	LikeFilter string
	ID         int64
	Limit      int64
}

type Cache interface {
	pageseo.Loader
	GetTargetBatch(context.Context, Cursor) ([]Target, error)
	MarkAsAnalyzed(context.Context, int64) error
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

	targets, err := c.GetTargetBatch(ctx, Cursor{
		LikeFilter: "http://localhost/%",
		Limit:      1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatal("expected 1 target, got:", len(targets))
	}
	target := targets[0]
	if target.Location != "http://localhost/" {
		t.Fatal("wrong target location:", target.Location)
	}
	if target.ContentType != ct {
		t.Fatal("wrong content type:", target.ContentType)
	}

	if err = c.MarkAsAnalyzed(ctx, target.ID); err != nil {
		t.Fatal(err)
	}

	targets, err = c.GetTargetBatch(ctx, Cursor{
		ID:         target.ID,
		LikeFilter: "http://localhost/%",
		Limit:      1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 0 {
		t.Fatal("expected 0 targets, got:", len(targets))
	}
}
