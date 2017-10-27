package logging

import (
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type Service struct {
	repo.Service

	article articleRepo
	feed    feedRepo
}

func NewService(s repo.Service, log log.Log) Service {
	return Service{
		s,
		articleRepo{s.ArticleRepo(), log},
		feedRepo{s.FeedRepo(), log},
	}
}

func (s Service) ArticleRepo() repo.Article {
	return s.article
}

func (s Service) FeedRepo() repo.Feed {
	return s.feed
}
