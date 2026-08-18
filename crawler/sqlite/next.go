package sqlite

import (
	"context"
	"time"

	"github.com/dkotik/pageseo/crawler"
)

func (c *cache) GetNextTarget(ctx context.Context) (t crawler.Target, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.stmtNext.Reset(); err != nil {
		return t, err
	}

	c.stmtNext.BindText(1, encodeTime(time.Now().Add(c.TimeToLive)))

	var ok bool
	for {
		ok, err = c.stmtNext.Step()
		if err != nil {
			return t, err
		}
		if !ok {
			break
		}
		// url, content_type, content, created_at, updated_at, analyzed_at
		t.Location = c.stmtNext.ColumnText(0)
		t.ContentType = c.stmtNext.ColumnText(1)
		_ = c.stmtNext.ColumnBytes(2, t.Content)
		t.CreatedAt, err = decodeTime(c.stmtNext.ColumnText(3))
		if err != nil {
			return t, err
		}
		t.UpdatedAt, err = decodeTime(c.stmtNext.ColumnText(4))
		if err != nil {
			return t, err
		}
		if !c.stmtNext.ColumnIsNull(6) {
			t.LastAnalyzedAt, err = decodeTime(c.stmtNext.ColumnText(5))
			if err != nil {
				return t, err
			}
		}
	}
	if t.Location == "" {
		return t, crawler.ErrNoMoreTargets
	}
	return t, nil
}

func (c *cache) MarkAsAnalyzed(ctx context.Context, target crawler.Target) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.stmtMark.Reset(); err != nil {
		return err
	}

	c.stmtMark.BindText(1, encodeTime(time.Now()))
	c.stmtMark.BindText(2, target.Location)

	var ok bool
	for {
		ok, err = c.stmtMark.Step()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
}
