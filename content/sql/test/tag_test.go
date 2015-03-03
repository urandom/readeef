package test

import (
	"testing"

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

	tf.AddArticles([]content.Article{createArticle(data.Article{Title: "article1"})})
	tf.AddArticles([]content.Article{createArticle(data.Article{Title: "article2", Link: "http://sugr.org"})})
	tf2.AddArticles([]content.Article{createArticle(data.Article{Title: "article3"})})

	ua := tag3.Articles()
	tests.CheckBool(t, false, tag3.HasErr(), tag3.Err())

	tests.CheckInt64(t, 3, int64(len(ua)))
}

func createTag(u content.User, d data.TagValue) (t content.Tag) {
	t = repo.Tag(u)
	t.Value(d)

	return
}
