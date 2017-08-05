package content

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/urandom/readeef/parser"
)

type FeedID int64

type Feed struct {
	ID             FeedID          `json:"id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Link           string          `json:"link"`
	SiteLink       string          `db:"site_link" json:"-"`
	HubLink        string          `db:"hub_link" json:"-"`
	UpdateError    string          `db:"update_error" json:"updateError"`
	SubscribeError string          `db:"subscribe_error" json:"subscribeError"`
	TTL            time.Duration   `json:"-"`
	SkipHours      map[int]bool    `json:"-"`
	SkipDays       map[string]bool `json:"-"`

	parsedArticles []Article
}

func (f Feed) Validate() error {
	if f.ID == 0 {
		return NewValidationError(errors.New("no ID"))
	}

	if f.Link == "" {
		return NewValidationError(errors.New("no link"))
	}

	if u, err := url.Parse(f.Link); err != nil || !u.IsAbs() {
		return NewValidationError(errors.New("no link"))
	}

	return nil
}

func (f *Feed) Refresh(pf parser.Feed) {
	f.Title = pf.Title
	f.Description = pf.Description
	f.SiteLink = pf.SiteLink
	f.HubLink = pf.HubLink

	f.parsedArticles = make([]Article, len(pf.Articles))

	for i := range pf.Articles {
		a := Article{
			Title:       pf.Articles[i].Title,
			Description: pf.Articles[i].Description,
			Link:        pf.Articles[i].Link,
			Date:        pf.Articles[i].Date,
		}
		a.FeedID = f.ID

		if pf.Articles[i].Guid != "" {
			a.Guid.Valid = true
			a.Guid.String = pf.Articles[i].Guid
		}

		f.parsedArticles[i] = a
	}
}

func (f Feed) ParsedArticles() (a []Article) {
	return f.parsedArticles
}

func (f Feed) String() string {
	return fmt.Sprintf("%d: %s", f.ID, f.Title)
}
