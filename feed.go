package readeef

import "readeef/parser"

type Feed struct {
	parser.Feed

	User     User
	Articles []Article
}

type Article struct {
	parser.Article

	Read     bool
	Favorite bool
}
