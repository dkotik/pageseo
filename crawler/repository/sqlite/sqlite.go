/*
Package sqlite provides a SQLite-backed cache for the crawler.
*/
package sqlite

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dkotik/pageseo"
	"github.com/dkotik/pageseo/crawler/repository"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type sqliteRepository struct {
	mu         *sync.Mutex
	Conn       *sqlite.Conn
	Loader     pageseo.Loader
	TimeToLive time.Duration

	stmtPush *sqlite.Stmt
	stmtMark *sqlite.Stmt
	stmtPull *sqlite.Stmt
	stmtNext *sqlite.Stmt
}

func New(
	conn *sqlite.Conn,
	loader pageseo.Loader,
	tableName string,
	timeToLive time.Duration,
) (_ repository.Repository, err error) {
	if tableName == "" {
		tableName = "pageseo_cache"
	}
	tableName = escapeIdentifier(tableName)

	if err = sqlitex.ExecScript(conn, `
		CREATE TABLE IF NOT EXISTS `+tableName+` (
			id integer PRIMARY KEY,
			url text NOT NULL UNIQUE,
			content_type text NOT NULL,
			content blob NOT NULL,
			created_at text NOT NULL,
			updated_at text NOT NULL,
			analyzed_at text
		) STRICT;
	`); err != nil {
		return nil, err
	}
	if timeToLive < 0 {
		panic("negative time to live")
	}

	c := &sqliteRepository{
		mu:         &sync.Mutex{},
		Conn:       conn,
		Loader:     loader,
		TimeToLive: timeToLive * -1,
	}
	c.stmtPush, err = conn.Prepare(`
		INSERT INTO ` + tableName + ` (url, content_type, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET content_type=excluded.content_type, content=excluded.content, updated_at=excluded.updated_at
	`)
	if err != nil {
		return nil, err
	}
	c.stmtMark, err = conn.Prepare(`
		UPDATE ` + tableName + ` SET analyzed_at=? WHERE id=?
	`)
	if err != nil {
		return nil, err
	}
	c.stmtPull, err = conn.Prepare(`
		SELECT content_type, content FROM ` + tableName + ` WHERE url=? AND updated_at>?
	`)
	if err != nil {
		return nil, err
	}
	c.stmtNext, err = conn.Prepare(`
		SELECT id, url, content_type, content, created_at, updated_at, analyzed_at FROM ` + tableName + ` WHERE content_type='text/html' AND (analyzed_at IS NULL OR analyzed_at<?) AND url LIKE ? AND id>? ORDER BY id DESC LIMIT ?
	`)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// escapeIdentifier safely quotes an SQLite table or column name.
func escapeIdentifier(name string) string {
	// Double quotes are escaped by doubling them in SQL identifiers
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

func encodeTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func decodeTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
