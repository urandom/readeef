package readeef

import (
	"errors"
	"net/url"
	"readeef/parser"
)

type Feed struct {
	parser.Feed

	User           User
	HubLink        string `db:"hub_link"`
	UpdateError    string `db:"update_error"`
	SubscribeError string `db:"subscribe_error"`
	Articles       []Article
}

type Article struct {
	parser.Article

	FeedLink string `db:"feed_link"`
	Read     bool
	Favorite bool
}

func (f Feed) UpdateFromParsed(pf parser.Feed) Feed {
	f.Feed = pf
	f.HubLink = pf.HubLink

	newArticles := make([]Article, len(pf.Articles))

	for i, pa := range pf.Articles {
		a := Article{Article: pa}
		a.FeedLink = f.Link
		newArticles[i] = a
	}

	f.Articles = newArticles

	return f
}

func (f Feed) Validate() error {
	if u, err := url.Parse(f.Link); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Feed has no link")}
	}

	return nil
}

func (a Article) Validate() error {
	if a.Id == "" {
		return ValidationError{errors.New("Article has no id")}
	}
	if u, err := url.Parse(a.FeedLink); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Article has no feed link")}
	}

	return nil
}
