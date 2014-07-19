package readeef

import (
	"errors"
	"readeef/parser"
)

type Feed struct {
	parser.Feed

	User     User
	HubLink  string `db:"hub_link"`
	Articles []Article
}

type Article struct {
	parser.Article
	FeedLink string `db:"feed_link"`
	Read     bool
	Favorite bool

	Feed Feed
}

func (f Feed) Validate() error {
	if f.Link == "" {
		return ValidationError{errors.New("Feed has no link")}
	}

	return nil
}

func (a Article) Validate() error {
	if a.Id == "" {
		return ValidationError{errors.New("Article has no id")}
	}

	return nil
}
