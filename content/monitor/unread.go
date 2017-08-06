package monitor

import (
	"context"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type Unread struct {
	log log.Log
}

func NewUnread(ctx context.Context, repo repo.Article, log log.Log) Unread {
	go func() {
		repo.DeleteStaleUnreadRecords()

		ticker := time.NewTicker(24 * time.Hour)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			repo.DeleteStaleUnreadRecords()
		}
	}()

	return Unread{log: log}
}

func (i Unread) FeedUpdated(feed content.Feed) error {
	i.log.Infof("Adding 'unread' states for all new articles of %s' for all users\n", feed)

	feed.SetNewArticlesUnread()

	if feed.HasErr() {
		return feed.Err()
	} else {
		return nil
	}
}

func (i Unread) FeedDeleted(feed content.Feed) error {
	return nil
}
