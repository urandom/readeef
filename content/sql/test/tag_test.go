package test

import (
	"testing"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestTag(t *testing.T) {
	u := createUser(data.User{Login: "tag_login"})

	tag := repo.Tag(u)
	tests.CheckBool(t, false, tag.HasErr(), tag.Err())

	tf := createTaggedFeed(u, data.Feed{Id: 1, Link: "http://sugr.org"})

	tests.CheckInt64(t, 0, int64(len(tf.Tags())))
	tests.CheckInt64(t, 1, int64(len(tf.Tags([]content.Tag{tag}))))

	tf.UpdateTags()
	tests.CheckBool(t, true, tf.HasErr())
	_, ok := tf.Err().(base.ValidationError)
	tests.CheckBool(t, true, ok)

	tag.Value("tag1")
	tests.CheckString(t, "tag1", tag.String())

	tf.Tags([]content.Tag{tag})
	tf.UpdateTags()
	tests.CheckBool(t, false, tf.HasErr(), tf.Err())

	tf2 := createTaggedFeed(u, data.Feed{Id: 2, Link: "http://sugr.org/products/readeef"})

	tag2 := repo.Tag(u)
	tag2.Value(data.TagValue("tag2"))

	tag3 := repo.Tag(u)
	tag3.Value(data.TagValue("tag3"))

	tests.CheckInt64(t, 2, int64(len(tf2.Tags([]content.Tag{tag2, tag3}))))
	tf2.UpdateTags()

	tags := u.Tags()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	tests.CheckInt64(t, 3, int64(len(tags)))

	feeds := u.AllTaggedFeeds()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	for i := range feeds {
		tags := feeds[i].Tags()
		switch feeds[i].Data().Id {
		case 1:
			tests.CheckInt64(t, 1, int64(len(tags)))
		case 2:
			tests.CheckInt64(t, 2, int64(len(tags)))
		}
	}

	tf.Tags([]content.Tag{tag, tag3})
	tf.UpdateTags()
	tests.CheckBool(t, false, tf.HasErr(), tf.Err())

	feeds = tag.AllFeeds()
	tests.CheckBool(t, false, tag.HasErr(), tag.Err())

	tests.CheckInt64(t, 1, int64(len(feeds)))
	tests.CheckInt64(t, 1, int64(feeds[0].Data().Id))

	feeds = tag3.AllFeeds()
	tests.CheckBool(t, false, tag.HasErr(), tag.Err())

	tests.CheckInt64(t, 2, int64(len(feeds)))

	now := time.Now()

	tf.AddArticles([]content.Article{createArticle(data.Article{Title: "article1", Date: now, Id: 1})})
	tf.AddArticles([]content.Article{createArticle(data.Article{Title: "article2", Link: "http://sugr.org", Date: now.Add(3 * time.Hour), Id: 2})})
	tf2.AddArticles([]content.Article{createArticle(data.Article{Title: "article3", Date: now.Add(-2 * time.Hour), Id: 3})})

	ua := tag3.Articles()
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())

	tests.CheckInt64(t, 3, int64(len(ua)))

	tests.CheckInt64(t, 1, int64(ua[0].Data().Id))
	tests.CheckString(t, "article2", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(-2*time.Hour).Equal(ua[2].Data().Date))

	tag3.SortingByDate()
	ua = tag3.Articles()

	tests.CheckInt64(t, 3, int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(3*time.Hour).Equal(ua[2].Data().Date))

	tag3.Reverse()
	ua = tag3.Articles()

	tests.CheckInt64(t, 2, int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(-2*time.Hour).Equal(ua[2].Data().Date))

	ua[0].Read(true)

	tag3.Reverse()
	tag3.DefaultSorting()

	ua = tag3.UnreadArticles()
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 2, int64(len(ua)))

	tests.CheckInt64(t, 1, int64(ua[0].Data().Id))
	tests.CheckString(t, "article3", ua[1].Data().Title)

	u.ArticleById(data.ArticleId(2)).Read(false)

	ua = tag3.UnreadArticles()
	tests.CheckInt64(t, 3, int64(len(ua)))

	tag3.ReadBefore(now.Add(time.Minute), true)
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())

	ua = tag3.UnreadArticles()
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())
	tests.CheckInt64(t, 1, int64(len(ua)))
	tests.CheckInt64(t, 2, int64(ua[0].Data().Id))
}

func createTag(u content.User, d data.TagValue) (t content.Tag) {
	t = repo.Tag(u)
	t.Value(d)

	return
}
