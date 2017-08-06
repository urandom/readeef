package monitor

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type Unread struct {
	articleRepo repo.Article
	userRepo    repo.User
	log         log.Log
}

func NewUnread(ctx context.Context, service repo.Service, log log.Log) Unread {
	repo := service.ArticleRepo()
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

	return Unread{repo: repo, userRepo: services.UserRepo(), log: log}
}

func (i Unread) FeedUpdated(feed content.Feed, articles []content.Article) error {
	i.log.Infof("Adding 'unread' states for all new articles of %s' for all users\n", feed)

	ids := make([]content.ArticleID, len(articles))
	for i := range articles {
		ids[i] = articles[i].ID
	}

	users, err := i.userRepo.All()
	if err != nil {
		return errors.WithMessage(err, "getting all users")
	}

	for _, user := range users {
		if err := i.articleRepo.Read(false, user, content.IDs(ids)); err != nil {
			return errors.WithMessage(err, "marking new articles as unread")
		}
	}

	return nil
}

func (i Unread) FeedDeleted(feed content.Feed) error {
	return nil
}
