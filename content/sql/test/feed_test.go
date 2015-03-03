package test

import (
	"testing"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestFeed(t *testing.T) {
	f := repo.Feed()
	f.Data(data.Feed{Title: "feed title", Link: "http://sugr.org/en/products/gearshift"})

	tests.CheckInt64(t, 0, int64(f.Data().Id))

	f.Update()

	tests.CheckBool(t, false, f.HasErr(), f.Err())
	tests.CheckBool(t, false, f.Data().Id == 0)
	tests.CheckInt64(t, 0, int64(len(f.NewArticles())))

	now := time.Now()

	f.AddArticles([]content.Article{
		createArticle(data.Article{Title: "article1", Date: now, Id: 1}),
		createArticle(data.Article{Title: "article2", Date: now.Add(2 * time.Hour), Id: 2}),
		createArticle(data.Article{Title: "article3", Date: now.Add(-3 * time.Hour), Id: 3}),
	})
	tests.CheckBool(t, false, f.HasErr(), f.Err())

	tests.CheckInt64(t, 3, int64(len(f.NewArticles())))
}

func createFeed(d data.Feed) (f content.Feed) {
	f = repo.Feed()
	f.Data(d)
	f.Update()

	return
}

func createUserFeed(u content.User, d data.Feed) (uf content.UserFeed) {
	uf = repo.UserFeed(u)
	uf.Data(d)
	uf.Update()

	u.AddFeed(uf)

	return
}

func createTaggedFeed(u content.User, d data.Feed) (tf content.TaggedFeed) {
	tf = repo.TaggedFeed(u)
	tf.Data(d)
	tf.Update()

	u.AddFeed(tf)

	return
}
