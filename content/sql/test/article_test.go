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
	u := createUser(data.User{Login: "scored_article_login"})
	uf := createUserFeed(u, data.Feed{Link: "http://sugr.org/bg/sitemap.xml", Title: "article feed 1"})
	uf.AddArticles([]content.Article{
		createArticle(data.Article{Title: "article1", Date: now, Link: "http://sugr.org/bg/products/gearshift"}),
		createArticle(data.Article{Title: "article2", Date: now.Add(2 * time.Hour), Link: "http://sugr.org/bg/products/readeef"}),
		createArticle(data.Article{Title: "article3", Date: now.Add(-3 * time.Hour), Link: "http://sugr.org/bg/about/us"}),
	})
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())

	ua := uf.AllArticles()
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())
	tests.CheckInt64(t, 3, int64(len(ua)))
	id1, id3 := ua[0].Data().Id, ua[2].Data().Id

	asc1 := createArticleScores(data.ArticleScores{ArticleId: id1, Score1: 2, Score2: 2})
	asc2 := createArticleScores(data.ArticleScores{ArticleId: id3, Score1: 1, Score2: 3})

	sa := repo.Article()
	sa.Data(data.Article{Id: id1})

	tests.CheckInt64(t, asc1.Calculate(), sa.Scores().Calculate())
	tests.CheckBool(t, false, sa.HasErr(), sa.Err())

	sa.Data(data.Article{Id: id3})
	tests.CheckInt64(t, asc2.Calculate(), sa.Scores().Calculate())
	tests.CheckBool(t, false, sa.HasErr(), sa.Err())
}

func TestUserArticle(t *testing.T) {
	now := time.Now()
	u := createUser(data.User{Login: "user_article_login"})
	uf := createUserFeed(u, data.Feed{Link: "http://sugr.org/bg/404", Title: "user article feed 1"})
	uf.AddArticles([]content.Article{
		createArticle(data.Article{Title: "article1", Date: now, Link: "http://sugr.org/bg/products/readeef"}),
	})
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())

	articles := uf.AllArticles()
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())
	tests.CheckInt64(t, 1, int64(len(articles)))

	id := articles[0].Data().Id
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())

	a := u.ArticleById(id)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, true, a.Data().Read)
	tests.CheckBool(t, false, a.Data().Favorite)

	a.Read(false)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, false, a.Data().Read)
	tests.CheckBool(t, false, u.ArticleById(id).Data().Read)

	a.Read(true)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, true, a.Data().Read)
	tests.CheckBool(t, true, u.ArticleById(id).Data().Read)

	a.Favorite(true)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, true, a.Data().Favorite)
	tests.CheckBool(t, true, u.ArticleById(id).Data().Favorite)

	a.Favorite(false)
	tests.CheckBool(t, false, a.HasErr(), a.Err())
	tests.CheckBool(t, false, a.Data().Favorite)
	tests.CheckBool(t, false, u.ArticleById(id).Data().Favorite)
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
	sa = repo.Article()
	sa.Data(d)

	return sa
}
