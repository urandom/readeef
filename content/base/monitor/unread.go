package monitor

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/webfw"
)

type Unread struct {
	logger webfw.Logger
}

func NewUnread(repo content.Repo, l webfw.Logger) Unread {
	go func() {
		repo.DeleteStaleUnreadRecords()

		for range time.Tick(24 * time.Hour) {
			repo.DeleteStaleUnreadRecords()
		}
	}()

	return Unread{logger: l}
}

func (i Unread) FeedUpdated(feed content.Feed) error {
	i.logger.Infof("Adding 'unread' states for all new articles of %s' for all users\n", feed)

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
