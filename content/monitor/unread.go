package monitor

import (
	"context"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/log"
)

func Unread(ctx context.Context, service eventable.Service, log log.Log) {
	// Grab the non-eventable article repo. We don't want to notify on the
	// initial unread mark.
	articleRepo := service.Service.ArticleRepo()

	go func() {
		articleRepo.RemoveStaleUnreadRecords()

		ticker := time.NewTicker(24 * time.Hour)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			articleRepo.RemoveStaleUnreadRecords()
		}
	}()

	userRepo := service.UserRepo()
	for event := range service.Listener() {
		switch data := event.Data.(type) {
		case eventable.FeedUpdateData:
			log.Infof("Setting new feed %s articles to unread", data.Feed)

			ids := make([]content.ArticleID, len(data.NewArticles))
			for i := range data.NewArticles {
				ids[i] = data.NewArticles[i].ID
			}

			users, err := userRepo.All()
			if err != nil {
				log.Printf("Error getting all users: %+v", err)
				continue
			}

			for _, user := range users {
				if err := articleRepo.Read(
					false, user, content.IDs(ids),
					content.Filters(content.GetUserFilters(user)),
				); err != nil {
					log.Printf("Error marking new articles as unread: %+v", err)
				}
			}
		}
	}
}
