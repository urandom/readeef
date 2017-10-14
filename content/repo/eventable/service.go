package eventable

import (
	"context"

	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type Service struct {
	repo.Service
	eventBus bus

	article articleRepo
	feed    feedRepo
}

func NewService(ctx context.Context, s repo.Service, log log.Log) Service {
	bus := newBus(ctx)

	return Service{
		s, bus,
		articleRepo{s.ArticleRepo(), bus, log},
		feedRepo{s.FeedRepo(), bus, log},
	}
}

func (s Service) Listener() Stream {
	return s.eventBus.Listener()
}

func (s Service) ArticleRepo() repo.Article {
	return s.article
}

func (s Service) FeedRepo() repo.Feed {
	return s.feed
}
