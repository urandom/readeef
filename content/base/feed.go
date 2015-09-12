package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/parser"
)

type Feed struct {
	Error
	RepoRelated

	data           data.Feed
	parsedArticles []Article
}

type UserFeed struct {
	ArticleSorting
	UserRelated
}

func (f Feed) String() string {
	return f.data.Title + " (" + strconv.FormatInt(int64(f.data.Id), 10) + ")"
}

func (f *Feed) Data(d ...data.Feed) data.Feed {
	if f.HasErr() {
		return data.Feed{}
	}

	if len(d) > 0 {
		f.data = d[0]
	}

	return f.data
}

func (f Feed) Validate() error {
	if f.data.Link == "" {
		return content.NewValidationError(errors.New("Feed has no link"))
	}

	if u, err := url.Parse(f.data.Link); err != nil || !u.IsAbs() {
		return content.NewValidationError(errors.New("Feed has no link"))
	}

	return nil
}

func (f *Feed) Refresh(pf parser.Feed) {
	if f.HasErr() {
		return
	}

	d := f.Data()

	d.Title = pf.Title
	d.Description = pf.Description
	d.SiteLink = pf.SiteLink
	d.HubLink = pf.HubLink

	f.parsedArticles = make([]Article, len(pf.Articles))

	for i := range pf.Articles {
		ai := data.Article{
			Title:       pf.Articles[i].Title,
			Description: pf.Articles[i].Description,
			Link:        pf.Articles[i].Link,
			Date:        pf.Articles[i].Date,
		}
		ai.FeedId = d.Id

		if pf.Articles[i].Guid != "" {
			ai.Guid.Valid = true
			ai.Guid.String = pf.Articles[i].Guid
		}

		f.parsedArticles[i] = Article{data: ai}
	}

	f.Data(d)
}

func (f *Feed) ParsedArticles() (a []Article) {
	if f.HasErr() {
		return
	}

	return f.parsedArticles
}

func (f Feed) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(f.data)

	if err == nil {
		return b, nil
	} else {
		return []byte{}, fmt.Errorf("Error marshaling feed data for %s: %v", f, err)
	}
}

func (uf UserFeed) Validate() error {
	if uf.user == nil || uf.user.Data().Login == "" {
		return content.NewValidationError(errors.New("UserFeed has no user"))
	}

	return nil
}
