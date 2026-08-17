package sqlite

import (
	"context"
	"time"
)

func (c *cache) load(ctx context.Context, URL string) (content []byte, contentType string, err error) {
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
	}
	return content, contentType, nil
}

func (c *cache) Load(ctx context.Context, URL string) (content []byte, contentType string, err error) {
	content, contentType, err = c.load(ctx, URL)
	if err != nil || len(content) > 0 {
		return content, contentType, err
	}

	content, contentType, err = c.Loader.Load(ctx, URL)
	if err != nil {
		return nil, "", err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.stmtPush.Reset(); err != nil {
		return nil, "", err
	}

	// url, content_type, content, created_at, updated_at
	t := time.Now()
	c.stmtPush.BindText(1, URL)
	c.stmtPush.BindText(2, contentType)
	c.stmtPush.BindBytes(3, content)
	c.stmtPush.BindText(4, encodeTime(t))
	c.stmtPush.BindText(5, encodeTime(t))

	var ok bool
	for {
		ok, err = c.stmtPush.Step()
		if err != nil {
			return nil, "", err
		}
		if !ok {
			break
		}
		// _ = c.stmtPull.ColumnBytes(1, content)
		// contentType = c.stmtPull.ColumnText(2)
	}

	return content, contentType, nil
}
