package sqlite

import (
	"fmt"
	"testing"
	"time"

	"github.com/dkotik/pageseo/crawler/repository"
	"github.com/dkotik/pageseo/internal"
	"zombiezen.com/go/sqlite"
)

func TestRepository(t *testing.T) {
	conn, err := sqlite.OpenConn(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(
		conn,
		internal.NewMockLoader(func(s string) (string, error) {
			return fmt.Sprintf(`<html><body>%s</body></html>`, s), nil
		}, "text/html"),
		"tableName",
		time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	repository.Test(t, c)

	// basic store and recover
	ctx := t.Context()
	repo := c.(*sqliteRepository)
	if err = repo.push(ctx, "https://example.com/", "text/html", []byte("<html><body>test</body></html>")); err != nil {
		t.Fatal(err)
	}
	data, ct, err := repo.load(ctx, "https://example.com/")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "<html><body>test</body></html>" {
		t.Fatal("unexpected data:", string(data))
	}
	if ct != "text/html" {
		t.Fatal("unexpected content type")
	}

	// collect targets
	targets, err := repo.GetTargetBatch(ctx, repository.Cursor{
		ID:         0,
		LikeFilter: "https://example.com/%",
		Limit:      10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatal("unexpected number of targets:", len(targets))
	}

	for _, target := range targets {
		if err = repo.MarkAsAnalyzed(ctx, target.ID); err != nil {
			t.Fatal(err)
		}
	}

	targets, err = repo.GetTargetBatch(ctx, repository.Cursor{
		ID:         0,
		LikeFilter: "https://example.com/%",
		Limit:      10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 0 {
		t.Fatal("unexpected number of targets:", len(targets))
	}
}
