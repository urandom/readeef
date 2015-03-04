package test

import (
	"testing"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestScoredArticle(t *testing.T) {
	now := time.Now()
	u := createUser(data.User{Login: "user_feed_login"})
	uf := createUserFeed(u, data.Feed{Link: "http://sugr.org/bg/sitemap.xml", Title: "article feed 1"})
	uf.AddArticles([]content.Article{
		createArticle(data.Article{Id: 21, Title: "article1", Date: now, Link: "http://sugr.org/bg/products/gearshift"}),
		createArticle(data.Article{Id: 22, Title: "article2", Date: now.Add(2 * time.Hour), Link: "http://sugr.org/bg/products/readeef"}),
		createArticle(data.Article{Id: 23, Title: "article3", Date: now.Add(-3 * time.Hour), Link: "http://sugr.org/bg/about/us"}),
	})

	asc1 := createArticleScores(data.ArticleScores{ArticleId: 21, Score1: 2, Score2: 2})
	asc2 := createArticleScores(data.ArticleScores{ArticleId: 22, Score1: 1, Score2: 3})

	sa := repo.ScoredArticle()
	sa.Data(data.Article{Id: 21})

	tests.CheckInt64(t, asc1.Calculate(), sa.Scores().Calculate())
	tests.CheckBool(t, false, sa.HasErr(), sa.Err())

	sa.Data(data.Article{Id: 22})
	tests.CheckInt64(t, asc2.Calculate(), sa.Scores().Calculate())
	tests.CheckBool(t, false, sa.HasErr(), sa.Err())
}

func TestUserArticle(t *testing.T) {
	now := time.Now()
	u := createUser(data.User{Login: "user_feed_login"})
	uf := createUserFeed(u, data.Feed{Link: "http://sugr.org/bg/404", Title: "user article feed 1"})
	uf.AddArticles([]content.Article{
		createArticle(data.Article{Id: 31, Title: "article1", Date: now, Link: "http://sugr.org/bg/products/readeef"}),
	})

	a := u.ArticleById(31)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, false, a.Data().Read)
	tests.CheckBool(t, false, a.Data().Favorite)

	a.Read(true)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, true, a.Data().Read)
	tests.CheckBool(t, true, u.ArticleById(31).Data().Read)

	a.Read(false)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, false, a.Data().Read)
	tests.CheckBool(t, false, u.ArticleById(31).Data().Read)

	a.Favorite(true)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, true, a.Data().Favorite)
	tests.CheckBool(t, true, u.ArticleById(31).Data().Favorite)

	a.Favorite(false)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, false, a.Data().Favorite)
	tests.CheckBool(t, false, u.ArticleById(31).Data().Favorite)
}

func createArticle(d data.Article) (a content.Article) {
	a = repo.Article()
	a.Data(d)

	return
}

func createUserArticle(u content.User, d data.Article) (ua content.Article) {
	ua = repo.UserArticle(u)
	ua.Data(d)

	return
}

func createScoredArticle(u content.User, d data.Article) (sa content.Article) {
	sa = repo.ScoredArticle()
	sa.Data(d)

	return sa
}
