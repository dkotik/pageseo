package sqlite

import (
	"context"
	"os"
	"strings"
	"time"
)

func (c *sqliteRepository) load(ctx context.Context, URL string) (content []byte, contentType string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.stmtPull.Reset(); err != nil {
		return nil, "", err
	}
	c.stmtPull.BindText(1, URL)
	c.stmtPull.BindText(2, encodeTime(time.Now().Add(c.TimeToLive)))
	var ok bool
	for {
		ok, err = c.stmtPull.Step()
		if err != nil {
			return nil, "", err
		}
		if !ok {
			break
		}
		contentType = c.stmtPull.ColumnText(0)
		content = make([]byte, c.stmtPull.ColumnLen(1))
		_ = c.stmtPull.ColumnBytes(1, content)
	}
	if contentType == "" {
		return nil, "", os.ErrNotExist
	}
	return content, contentType, nil
}

func (c *sqliteRepository) Load(ctx context.Context, URL string) (content []byte, contentType string, err error) {
	content, contentType, err = c.load(ctx, URL)
	if err == nil || !os.IsNotExist(err) {
		return content, contentType, err
	}

	content, contentType, err = c.Loader.Load(ctx, URL)
	if err != nil {
		return nil, "", err
	}
	if err = c.push(ctx, URL, contentType, content); err != nil {
		return nil, "", err
	}
	return content, contentType, nil
}

func (c *sqliteRepository) push(ctx context.Context, URL, contentType string, content []byte) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.stmtPush.Reset(); err != nil {
		return err
	}

	// url, content_type, content, created_at, updated_at
	t := time.Now()
	c.stmtPush.BindText(1, URL)
	c.stmtPush.BindText(2, strings.ToLower(contentType))
	c.stmtPush.BindBytes(3, content)
	c.stmtPush.BindText(4, encodeTime(t))
	c.stmtPush.BindText(5, encodeTime(t))

	var ok bool
	for {
		ok, err = c.stmtPush.Step()
		if err != nil || !ok {
			break
		}
	}
	return err
}
