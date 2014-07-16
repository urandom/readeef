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

	FeedId   string `db:"feed_id"`
	Read     bool
	Favorite bool
}

func (f Feed) Validate() error {
	if f.Link == "" {
		return ValidationError{errors.New("Feed has no link")}
	}

	return nil
}
