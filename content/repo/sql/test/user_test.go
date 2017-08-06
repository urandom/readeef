package test

import (
	"testing"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestUser(t *testing.T) {
	u := repo.User()

	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u.Update()
	tests.CheckBool(t, true, u.HasErr())

	err := u.Err()
	_, ok := err.(content.ValidationError)
	tests.CheckBool(t, true, ok, err)

	u.Data(data.User{Login: data.Login("login")})

	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u.Update()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u2 := repo.UserByLogin(data.Login("login"))
	tests.CheckBool(t, false, u2.HasErr(), u2.Err())
	tests.CheckString(t, "login", string(u2.Data().Login))

	u.Delete()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u2 = repo.UserByLogin(data.Login("login"))
	tests.CheckBool(t, true, u2.HasErr())
	tests.CheckBool(t, true, u2.Err() == content.ErrNoContent)

	u = createUser(data.User{Login: data.Login("login")})

	now := time.Now()
	uf := createUserFeed(u, data.Feed{Link: "http://sugr.org/en/sitemap.xml", Title: "User feed 1"})
	uf.AddArticles([]content.Article{
		createArticle(data.Article{Title: "article1", Date: now, Link: "http://sugr.org/bg/products/gearshift"}),
		createArticle(data.Article{Title: "article2", Date: now.Add(2 * time.Hour), Link: "http://sugr.org/bg/products/readeef"}),
		createArticle(data.Article{Title: "article3", Date: now.Add(-3 * time.Hour), Link: "http://sugr.org/bg/about/us"}),
	})

	u.AddFeed(uf)

	var id1, id2, id3 data.ArticleId

	for _, a := range uf.AllArticles() {
		d := a.Data()
		switch d.Title {
		case "article1":
			id1 = d.Id
		case "article2":
			id2 = d.Id
		case "article3":
			id3 = d.Id
		default:
			tests.CheckBool(t, true, false, "Unknown article")
		}
	}

	tests.CheckBool(t, false, uf.HasErr(), uf.Err())
	tests.CheckInt64(t, 1, int64(len(u.AllFeeds())))
	tests.CheckString(t, "http://sugr.org/en/sitemap.xml", u.AllFeeds()[0].Data().Link)
	tests.CheckString(t, "User feed 1", u.AllFeeds()[0].Data().Title)

	a := u.ArticleById(10000000)
	tests.CheckBool(t, true, a.Err() == content.ErrNoContent)

	a = u.ArticleById(id1)
	tests.CheckBool(t, false, a.HasErr(), a.Err())

	tests.CheckString(t, "article1", a.Data().Title)

	a2 := u.ArticlesById([]data.ArticleId{100000000, id1, id2})
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	tests.CheckInt64(t, 2, int64(len(a2)))

	for i := range a2 {
		d := a2[i].Data()
		switch d.Title {
		case "article1":
		case "article2":
		default:
			tests.CheckBool(t, false, true, "Unknown article")
		}
	}

	u.SortingById()
	ua := u.Articles()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	tests.CheckInt64(t, 3, int64(len(ua)))

	tests.CheckInt64(t, int64(id1), int64(ua[0].Data().Id))
	tests.CheckString(t, "article2", ua[1].Data().Title)
	tests.CheckInt64(t, now.Add(-3*time.Hour).Unix(), ua[2].Data().Date.Unix())

	u.SortingByDate()
	ua = u.Articles()

	tests.CheckInt64(t, int64(id3), int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckInt64(t, now.Add(2*time.Hour).Unix(), ua[2].Data().Date.Unix())

	u.Reverse()
	ua = u.Articles()

	tests.CheckInt64(t, int64(id2), int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckInt64(t, now.Add(-3*time.Hour).Unix(), ua[2].Data().Date.Unix())

	ua[1].Read(false)
	ua[2].Read(false)

	u.Reverse()
	u.SortingById()

	ua = u.Articles(data.ArticleQueryOptions{UnreadOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, int64(len(ua)))

	tests.CheckInt64(t, int64(id1), int64(ua[0].Data().Id))
	tests.CheckString(t, "article3", ua[1].Data().Title)

	u.ArticleById(id2).Read(false)

	ua = u.Articles(data.ArticleQueryOptions{UnreadOnly: true})
	tests.CheckInt64(t, 3, int64(len(ua)))

	u.ReadState(true, data.ArticleUpdateStateOptions{BeforeDate: now.Add(time.Minute)})
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	ua = u.Articles(data.ArticleQueryOptions{UnreadOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, int64(len(ua)))
	tests.CheckInt64(t, int64(id2), int64(ua[0].Data().Id))

	u.ArticleById(id1).Read(false)

	u.ReadState(true, data.ArticleUpdateStateOptions{AfterDate: now.Add(time.Minute)})
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	ua = u.Articles(data.ArticleQueryOptions{UnreadOnly: true})
	tests.CheckInt64(t, 1, int64(len(ua)))
	tests.CheckInt64(t, int64(id1), int64(ua[0].Data().Id))

	u.ArticleById(id1).Favorite(true)
	u.ArticleById(id3).Favorite(true)

	uIds := u.Ids(data.ArticleIdQueryOptions{UnreadOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, int64(len(uIds)))
	tests.CheckInt64(t, int64(id1), int64(uIds[0]))

	fIds := u.Ids(data.ArticleIdQueryOptions{FavoriteOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, int64(len(fIds)))

	for i := range fIds {
		switch fIds[i] {
		case id1:
		case id3:
		default:
			tests.CheckBool(t, false, true, "Unknown article id")
		}
	}

	tests.CheckInt64(t, 3, u.Count())
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	ua = u.Articles(data.ArticleQueryOptions{FavoriteOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, int64(len(ua)))

	for i := range ua {
		d := ua[i].Data()
		switch d.Id {
		case id1:
		case id3:
		default:
			tests.CheckBool(t, false, true, "Unknown article id")
		}
	}

	u.SortingById()
	ua = u.Articles(data.ArticleQueryOptions{AfterId: id1})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, int64(len(ua)))

	for i := range ua {
		d := ua[i].Data()
		switch d.Id {
		case id2:
		case id3:
		default:
			tests.CheckBool(t, false, true, "Unknown article id")
		}
	}

	u.Reverse()
	ua = u.Articles(data.ArticleQueryOptions{BeforeId: id2})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, int64(len(ua)))
	tests.CheckString(t, "article1", ua[0].Data().Title)

	asc1 := createArticleScores(data.ArticleScores{ArticleId: id1, Score1: 2, Score2: 2})
	asc2 := createArticleScores(data.ArticleScores{ArticleId: id2, Score1: 1, Score2: 3})

	sa := u.Articles(data.ArticleQueryOptions{AfterDate: now.Add(-20 * time.Hour), BeforeDate: now.Add(20 * time.Hour), IncludeScores: true, HighScoredFirst: true})

	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, int64(len(sa)))

	for i := range sa {
		switch sa[i].Data().Id {
		case 1:
			tests.CheckInt64(t, asc1.Calculate(), sa[i].Data().Score)
		case 2:
			tests.CheckInt64(t, asc2.Calculate(), sa[i].Data().Score)
		}
	}

	ua = u.Articles()
	ua[0].Read(true)
	ua[1].Read(true)
	ua[2].Read(false)
	ua[0].Favorite(true)
	ua[1].Favorite(false)
	ua[2].Favorite(true)

	count := u.Count()
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 3, count)

	count = u.Count(data.ArticleCountOptions{UnreadOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, count)

	count = u.Count(data.ArticleCountOptions{FavoriteOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, count)

	count = u.Count(data.ArticleCountOptions{FavoriteOnly: true, UnreadOnly: true})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, count)

	count = u.Count(data.ArticleCountOptions{BeforeId: id2})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, count)

	count = u.Count(data.ArticleCountOptions{AfterId: id1})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, count)

	count = u.Count(data.ArticleCountOptions{BeforeId: id3, AfterId: id1})
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, count)
}

func createUser(d data.User) (u content.User) {
	u = repo.User()
	u.Data(d)
	u.Update()

	return
}
