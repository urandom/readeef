package logging

import (
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type Service struct {
	repo.Service

	article      articleRepo
	extract      extractRepo
	feed         feedRepo
	scores       scoresRepo
	subscription subscriptionRepo
	tag          tagRepo
	thumbnail    thumbnailRepo
	user         userRepo
}

func NewService(s repo.Service, log log.Log) Service {
	return Service{
		s,
		articleRepo{s.ArticleRepo(), log},
		extractRepo{s.ExtractRepo(), log},
		feedRepo{s.FeedRepo(), log},
		scoresRepo{s.ScoresRepo(), log},
		subscriptionRepo{s.SubscriptionRepo(), log},
		tagRepo{s.TagRepo(), log},
		thumbnailRepo{s.ThumbnailRepo(), log},
		userRepo{s.UserRepo(), log},
	}
}

func (s Service) ArticleRepo() repo.Article {
	return s.article
}

func (s Service) ExtractRepo() repo.Extract {
	return s.extract
}

func (s Service) FeedRepo() repo.Feed {
	return s.feed
}

func (s Service) ScoresRepo() repo.Scores {
	return s.scores
}

func (s Service) SubscriptionRepo() repo.Subscription {
	return s.subscription
}

func (s Service) TagRepo() repo.Tag {
	return s.tag
}

func (s Service) ThumbnailRepo() repo.Thumbnail {
	return s.thumbnail
}

func (s Service) UserRepo() repo.User {
	return s.user
}
