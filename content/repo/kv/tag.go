package kv

import "github.com/urandom/readeef/content"

type tagUser struct {
	ID    int           `storm:"increment"`
	TagID content.TagID `storm:"index"`
	Login content.Login `storm:"index"`
}

const (
	tagsBucket      = "tags"
	tagsUsersBucket = "users"
)
