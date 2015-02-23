package base

import (
	"errors"
	"strconv"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type SortingField int

const (
	DefaultSort SortingField = iota
	SortById
	SortByDate
)

type Article struct {
	Error

	info info.Article
}

type UserArticle struct {
	user content.User
}

func (a Article) String() string {
	return a.info.Title + " " + strconv.FormatInt(int64(a.info.Id), 10)
}

func (a *Article) Set(info info.Article) content.Article {
	if a.Err() != nil {
		return a
	}

	a.info = info

	return a
}

func (a Article) Info() info.Article {
	return a.info
}

func (a Article) Validate() error {
	if a.info.FeedId == 0 {
		return ValidationError{errors.New("Article has no feed id")}
	}

	return nil
}

func (ua UserArticle) User() content.User {
	return ua.user
}
