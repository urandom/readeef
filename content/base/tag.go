package base

import "github.com/urandom/readeef/content/info"

type Tag struct {
	ArticleSorting
	ArticleSearch
	Error
	UserRelated
	RepoRelated

	value info.TagValue
}

func (t Tag) String() string {
	return string(t.value)
}

func (t Tag) Set(value info.TagValue) {
	if t.Err() != nil {
		return
	}

	t.value = value
}

func (t Tag) Value() info.TagValue {
	return t.value
}
