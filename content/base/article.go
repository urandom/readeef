package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/urandom/readeef/content"
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
		return content.NewValidationError(errors.New("Article has no feed id"))
	}

	if !a.data.Guid.Valid && a.data.Link == "" {
		return content.NewValidationError(errors.New("Article has no guid or link"))
	}

	return nil
}

func (a Article) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(a.data)

	if err == nil {
		return b, nil
	} else {
		return []byte{}, fmt.Errorf("Error marshaling article data for %s: %v", a, err)
	}
}
