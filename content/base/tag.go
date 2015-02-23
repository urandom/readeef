package base

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type Tag struct {
	ArticleSorting
	ArticleSearch
	Error

	value info.TagValue
}

func (t Tag) String() string {
	return string(t.value)
}

func (t *Tag) Set(value info.TagValue) content.Tag {
	if t.Err() != nil {
		return t
	}

	t.value = value

	return t
}

func (t Tag) Value() info.TagValue {
	return t.value
}

func (t Tag) AllFeeds() []content.TaggedFeed {
	panic("Not implemented")
}

func (t Tag) Articles(desc bool, paging ...int) []content.UserArticle {
	panic("Not implemented")
}

func (t Tag) UnreadArticles(desc bool, paging ...int) []content.UserArticle {
	panic("Not implemented")
}

func (t Tag) ReadBefore(date time.Time, read bool) content.Tag {
	panic("Not implemented")
}

func (t Tag) ScoredArticles(from, to time.Time, desc bool, paging ...int) []content.ScoredArticle {
	panic("Not implemented")
}
