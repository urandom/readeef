package base

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type Tag struct {
	ArticleSorting
	ArticleSearch
	Error

	value info.TagValue
	user  content.User
}

func NewTag(user content.User) Tag {
	return Tag{user: user}
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

func (t Tag) User() content.User {
	return t.user
}
