package sqlite

import (
	"context"
	"time"

	"github.com/dkotik/pageseo/crawler/repository"
)

func (c *sqliteRepository) GetTargetBatch(ctx context.Context, cursor repository.Cursor) (targets []repository.Target, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.stmtNext.Reset(); err != nil {
		return nil, err
	}

	c.stmtNext.BindText(1, encodeTime(time.Now().Add(c.TimeToLive)))
	c.stmtNext.BindText(2, cursor.LikeFilter)
	c.stmtNext.BindInt64(3, cursor.ID)
	c.stmtNext.BindInt64(4, cursor.Limit)

	var t repository.Target
	var ok bool
	for {
		ok, err = c.stmtNext.Step()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		// id, url, content_type, content, created_at, updated_at, analyzed_at
		t.ID = c.stmtNext.ColumnInt64(0)
		t.Location = c.stmtNext.ColumnText(1)
		t.ContentType = c.stmtNext.ColumnText(2)
		_ = c.stmtNext.ColumnBytes(3, t.Content)
		t.CreatedAt, err = decodeTime(c.stmtNext.ColumnText(4))
		if err != nil {
			return nil, err
		}
		t.UpdatedAt, err = decodeTime(c.stmtNext.ColumnText(5))
		if err != nil {
			return nil, err
		}
		if !c.stmtNext.ColumnIsNull(6) {
			t.LastAnalyzedAt, err = decodeTime(c.stmtNext.ColumnText(6))
			if err != nil {
				return nil, err
			}
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (c *sqliteRepository) MarkAsAnalyzed(ctx context.Context, id int64) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.stmtMark.Reset(); err != nil {
		return err
	}

	c.stmtMark.BindText(1, encodeTime(time.Now()))
	c.stmtMark.BindInt64(2, id)

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
