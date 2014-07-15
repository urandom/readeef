package readeef

import "readeef/parser"

type Feed struct {
	parser.Feed
	Id int

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
