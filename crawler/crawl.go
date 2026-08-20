package crawler

import (
	"context"
	"strings"
	"time"

	"github.com/dkotik/pageseo/crawler/repository"
)

func (c *crawler) CrawlLocation(ctx context.Context, URL string) (err error) {
	if !strings.HasSuffix(URL, "/") {
		URL += "/"
	}
	cursor := repository.Cursor{
		LikeFilter: URL + "%",
		ID:         0,
		Limit:      int64(c.BatchSize),
	}
	// seed the repository with the initial URL
	_, _, err = c.Repository.Load(ctx, URL)
	if err != nil {
		return err
	}

	var t repository.Target
	for {
		batch, err := c.Repository.GetTargetBatch(ctx, cursor)
		if err != nil || len(batch) == 0 {
			break
		}
		for _, t = range batch {
			if time.Now().Sub(t.UpdatedAt) > c.TimeToLive {
				t.Content, t.ContentType, err = c.Repository.Load(ctx, t.Location)
				if err != nil {
					return err
				}
				t.UpdatedAt = time.Now()
			}
			if err = c.Analyzer.Analyze(ctx, t); err != nil {
				return err
			}
			if err = c.Repository.MarkAsAnalyzed(ctx, t.ID); err != nil {
				return err
			}
		}
		cursor.ID = t.ID // set cursor ID to last target ID
	}
	return err
}
