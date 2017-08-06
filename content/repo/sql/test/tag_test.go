package test

import (
	"testing"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestTag(t *testing.T) {
	u := createUser(data.User{Login: "tag_login"})

	tag := repo.Tag(u)
	tests.CheckBool(t, false, tag.HasErr(), tag.Err())

	tf := createTaggedFeed(u, data.Feed{Link: "http://sugr.org/"})

	tests.CheckInt64(t, 0, int64(len(tf.Tags())))
	tests.CheckInt64(t, 1, int64(len(tf.Tags([]content.Tag{tag}))))

	tf.UpdateTags()
	tests.CheckBool(t, true, tf.HasErr())
	_, ok := tf.Err().(content.ValidationError)
	tests.CheckBool(t, true, ok)

	tag.Data(data.Tag{Value: "tag1"})
	tests.CheckString(t, "tag1", tag.String())

	tf.Tags([]content.Tag{tag})
	tf.UpdateTags()
	tests.CheckBool(t, false, tf.HasErr(), tf.Err())

	tf2 := createTaggedFeed(u, data.Feed{Link: "http://sugr.org/products/readeef"})

	tag2 := repo.Tag(u)
	tag2.Data(data.Tag{Value: "tag2"})

	tag3 := repo.Tag(u)
	tag3.Data(data.Tag{Value: "tag3"})

	tests.CheckInt64(t, 2, int64(len(tf2.Tags([]content.Tag{tag2, tag3}))))
	tf2.UpdateTags()
	tests.CheckBool(t, false, tf2.HasErr(), tf2.Err())

	tags := u.Tags()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	tests.CheckInt64(t, 3, int64(len(tags)))

	feeds := u.AllTaggedFeeds()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	var fId1 data.FeedId
	for i := range feeds {
		tags := feeds[i].Tags()
		d := feeds[i].Data()
		switch d.Link {
		case "http://sugr.org/":
			fId1 = d.Id
			tests.CheckInt64(t, 1, int64(len(tags)))
		case "http://sugr.org/products/readeef":
			tests.CheckInt64(t, 2, int64(len(tags)))
		default:
			tests.CheckBool(t, false, true, "Unknown feed")
		}
	}

	tf.Tags([]content.Tag{tag, tag3})
	tf.UpdateTags()
	tests.CheckBool(t, false, tf.HasErr(), tf.Err())

	feeds = tag.AllFeeds()
	tests.CheckBool(t, false, tag.HasErr(), tag.Err())

	tests.CheckInt64(t, 1, int64(len(feeds)))
	tests.CheckInt64(t, int64(fId1), int64(feeds[0].Data().Id))

	feeds = tag3.AllFeeds()
	tests.CheckBool(t, false, tag.HasErr(), tag.Err())

	tests.CheckInt64(t, 2, int64(len(feeds)))

	now := time.Now()

	tf.AddArticles([]content.Article{
		createArticle(data.Article{Title: "article1", Date: now, Link: "http://1.example.com"}),
		createArticle(data.Article{Title: "article2", Link: "http://sugr.org", Date: now.Add(3 * time.Hour)}),
	})
	tf2.AddArticles([]content.Article{createArticle(data.Article{Title: "article3", Date: now.Add(-2 * time.Hour), Link: "http://sugr.org/products/readeef"})})

	tag3.SortingById()
	ua := tag3.Articles()
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 3, int64(len(ua)))

	var id1, id2, id3 data.ArticleId
	for i := range ua {
		d := ua[i].Data()
		switch d.Title {
		case "article1":
			id1 = d.Id
		case "article2":
			id2 = d.Id
		case "article3":
			id3 = d.Id
		default:
			tests.CheckBool(t, false, true, "Unknown article")
		}
	}

	tests.CheckInt64(t, int64(id1), int64(ua[0].Data().Id))
	tests.CheckString(t, "article2", ua[1].Data().Title)
	tests.CheckInt64(t, now.Add(-2*time.Hour).Unix(), ua[2].Data().Date.Unix())

	tag3.SortingByDate()
	ua = tag3.Articles()

	tests.CheckInt64(t, int64(id3), int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckInt64(t, now.Add(3*time.Hour).Unix(), ua[2].Data().Date.Unix())

	tag3.Reverse()
	ua = tag3.Articles()

	tests.CheckInt64(t, int64(id2), int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckInt64(t, now.Add(-2*time.Hour).Unix(), ua[2].Data().Date.Unix())

	ua[1].Read(false)
	ua[2].Read(false)

	tag3.Reverse()
	tag3.SortingById()

	ua = tag3.Articles(data.ArticleQueryOptions{UnreadOnly: true})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 2, int64(len(ua)))

	tests.CheckInt64(t, int64(id1), int64(ua[0].Data().Id))
	tests.CheckString(t, "article3", ua[1].Data().Title)

	u.ArticleById(id2).Read(false)

	ua = tag3.Articles(data.ArticleQueryOptions{UnreadOnly: true})
	tests.CheckInt64(t, 3, int64(len(ua)))

	tag3.ReadState(true, data.ArticleUpdateStateOptions{BeforeDate: now.Add(time.Minute)})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())

	ua = tag3.Articles(data.ArticleQueryOptions{UnreadOnly: true})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 1, int64(len(ua)))
	tests.CheckInt64(t, int64(id2), int64(ua[0].Data().Id))

	asc1 := createArticleScores(data.ArticleScores{ArticleId: id1, Score1: 2, Score2: 2})
	asc2 := createArticleScores(data.ArticleScores{ArticleId: id2, Score1: 1, Score2: 3})

	sa := tag3.Articles(data.ArticleQueryOptions{AfterDate: now.Add(-20 * time.Hour), BeforeDate: now.Add(20 * time.Hour), IncludeScores: true, HighScoredFirst: true})

	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 2, int64(len(sa)))

	for i := range sa {
		switch sa[i].Data().Id {
		case 1:
			tests.CheckInt64(t, asc1.Calculate(), sa[i].Data().Score)
		case 2:
			tests.CheckInt64(t, asc2.Calculate(), sa[i].Data().Score)
		}
	}

	ua = tag3.Articles()
	ua[0].Read(true)
	ua[1].Read(true)
	ua[2].Read(false)
	ua[0].Favorite(true)
	ua[1].Favorite(false)
	ua[2].Favorite(true)

	count := tag3.Count()
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 3, count)

	count = tag3.Count(data.ArticleCountOptions{UnreadOnly: true})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 1, count)

	count = tag3.Count(data.ArticleCountOptions{FavoriteOnly: true})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 2, count)

	count = tag3.Count(data.ArticleCountOptions{FavoriteOnly: true, UnreadOnly: true})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 1, count)

	count = tag3.Count(data.ArticleCountOptions{BeforeId: id2})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 1, count)

	count = tag3.Count(data.ArticleCountOptions{AfterId: id1})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 2, count)

	count = tag3.Count(data.ArticleCountOptions{BeforeId: id3, AfterId: id1})
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 1, count)
}
