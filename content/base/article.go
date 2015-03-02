package base

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/urandom/readeef/content/data"
)

type Article struct {
	Error
	RepoRelated

	data data.Article
}

type UserArticle struct {
	UserRelated
}

func (a Article) String() string {
	return a.data.Title + " (" + strconv.FormatInt(int64(a.data.Id), 10) + ")"
}

func (a *Article) Data(d ...data.Article) data.Article {
	if a.HasErr() {
		return data.Article{}
	}

	if len(d) > 0 {
		a.data = d[0]
	}

	return a.data
}

func (a Article) Validate() error {
	if a.data.FeedId == 0 {
		return ValidationError{errors.New("Article has no feed id")}
	}

	return nil
}

func (a Article) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.data)
}
