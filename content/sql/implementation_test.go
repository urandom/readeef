package sql

import (
	"testing"

	"github.com/urandom/readeef/content"
)

func TestImplements(t *testing.T) {
	var article content.Article

	r := NewRepo(nil, nil)

	article = r.Article()
	article.Data()

	var userArticle content.UserArticle

	userArticle = r.UserArticle(nil)
	userArticle.Data()

	var scoredArticle content.Article

	scoredArticle = r.Article()
	scoredArticle.Data()

	var feed content.Feed

	feed = r.Feed()
	feed.Data()

	var userFeed content.UserFeed

	userFeed = r.UserFeed(nil)
	userFeed.Data()

	var taggedFeed content.TaggedFeed

	taggedFeed = r.TaggedFeed(nil)
	taggedFeed.Data()

	r.HasErr()

	var subscription content.Subscription

	subscription = r.Subscription()
	subscription.Data()

	var tag content.Tag

	tag = r.Tag(nil)
	tag.Data()

	var user content.User

	user = r.User()
	user.Data()
}
