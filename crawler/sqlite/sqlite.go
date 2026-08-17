package sqlite

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dkotik/pageseo"
	"github.com/dkotik/pageseo/crawler"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type cache struct {
	mu         *sync.Mutex
	Conn       *sqlite.Conn
	Loader     pageseo.Loader
	TimeToLive time.Duration

	stmtPush *sqlite.Stmt
	stmtMark *sqlite.Stmt
	stmtPull *sqlite.Stmt
	stmtNext *sqlite.Stmt
}

func chainLikeOperators(prefixes ...string) string {
	if len(prefixes) == 0 {
		return "1"
	}
	b := bytes.Buffer{}
	for _, p := range prefixes {
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		_, _ = b.WriteString(` url LIKE '`)
		_, _ = b.WriteString(p)
		_, _ = b.WriteString(`%' OR`)
	}
	b.Truncate(b.Len() - 4)
	return b.String()
}

func New(
	conn *sqlite.Conn,
	loader pageseo.Loader,
	tableName string,
	timeToLive time.Duration,
	allTargetPrefixes ...string,
) (_ crawler.Cache, err error) {
	if tableName == "" {
		tableName = "pageseo_cache"
	}
	tableName = escapeIdentifier(tableName)

	if err = sqlitex.ExecScript(conn, `
		CREATE TABLE IF NOT EXISTS `+tableName+` (
			url text PRIMARY KEY,
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

	c := &cache{
		mu:         &sync.Mutex{},
		Conn:       conn,
		Loader:     loader,
		TimeToLive: timeToLive * -1,
	}
	c.stmtPush, err = conn.Prepare(`
		INSERT INTO ` + tableName + ` (url, content_type, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET content_type = excluded.content_type, content = excluded.content, updated_at = excluded.updated_at;
		;
	`)
	if err != nil {
		return nil, err
	}
	c.stmtMark, err = conn.Prepare(`
		UPDATE ` + tableName + ` SET analyzed_at=? WHERE url=?;
	`)
	if err != nil {
		return nil, err
	}
	c.stmtPull, err = conn.Prepare(`
		SELECT content_type, content FROM ` + tableName + ` WHERE url=? AND updated_at>?;
	`)
	if err != nil {
		return nil, err
	}
	c.stmtNext, err = conn.Prepare(`
		SELECT url, content_type, content, created_at, updated_at, analyzed_at FROM ` + tableName + ` WHERE analyzed_at>? AND (` + chainLikeOperators(allTargetPrefixes...) + `);
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
