package base

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/urandom/readeef/content/info"
)

type Article struct {
	Error
	RepoRelated

	info info.Article
}

type UserArticle struct {
	UserRelated
}

func (a Article) String() string {
	return a.info.Title + " " + strconv.FormatInt(int64(a.info.Id), 10)
}

func (a *Article) Info(in ...info.Article) info.Article {
	if a.HasErr() {
		return info.Article{}
	}

	if len(in) > 0 {
		a.info = in[0]
	}

	return a.info
}

func (a Article) Validate() error {
	if a.info.FeedId == 0 {
		return ValidationError{errors.New("Article has no feed id")}
	}

	return nil
}

func (a Article) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.info)
}
