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

func (t *Tag) Value(val ...info.TagValue) info.TagValue {
	if t.HasErr() {
		return t.value
	}

	if len(val) > 0 {
		t.value = val[0]
	}

	return t.value
}
