package logging

import (
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type Service struct {
	repo.Service

	article articleRepo
}

func NewService(s repo.Service, log log.Log) Service {
	return Service{
		s,
		articleRepo{s.ArticleRepo(), log},
	}
}

func (s Service) ArticleRepo() repo.Article {
	return s.article
}
