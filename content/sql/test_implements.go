package sql

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestImplements(t *testing.T) {
	var article content.Article

	article = NewArticle()
	article.Info()

	var userArticle content.UserArticle

	userArticle = NewUserArticle(nil, nil, nil)
	userArticle.Info()

	var scoredArticle content.ScoredArticle

	scoredArticle = NewScoredArticle(nil, nil, nil)
	scoredArticle.Info()

	var feed content.Feed

	feed = NewFeed(nil, nil)
	feed.Info()

	var userFeed content.UserFeed

	userFeed = NewUserFeed(nil, nil, nil)
	userFeed.Info()

	var taggedFeed content.TaggedFeed

	taggedFeed = NewTaggedFeed(nil, nil, nil)
	taggedFeed.Info()

	var repo content.Repo

	repo = NewRepo(nil, nil)
	repo.Err()

	var subscription content.Subscription

	subscription = NewSubscription(nil, nil)
	subscription.Info()

	var tag content.Tag

	tag = NewTag(nil, nil)
	tag.Value()

	var user content.User

	user = NewUser(nil, nil)
	user.Info()
}
