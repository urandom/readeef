package sql

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestImplements(t *testing.T) {
	var article content.Article

	r := NewRepo(nil, nil)

	article = r.Article()
	article.Info()

	var userArticle content.UserArticle

	userArticle = r.UserArticle(nil)
	userArticle.Info()

	var scoredArticle content.ScoredArticle

	scoredArticle = r.ScoredArticle(nil)
	scoredArticle.Info()

	var feed content.Feed

	feed = r.Feed()
	feed.Info()

	var userFeed content.UserFeed

	userFeed = r.UserFeed(nil)
	userFeed.Info()

	var taggedFeed content.TaggedFeed

	taggedFeed = r.TaggedFeed(nil)
	taggedFeed.Info()

	r.Err()

	var subscription content.Subscription

	subscription = r.Subscription()
	subscription.Info()

	var tag content.Tag

	tag = r.Tag(nil)
	tag.Value()

	var user content.User

	user = r.User()
	user.Info()
}
