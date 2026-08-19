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
}
