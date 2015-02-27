package info

import (
	"database/sql"
	"time"
)

type ArticleId int64
type SortingField int
type Order int

const (
	DefaultSort SortingField = iota
	SortById
	SortByDate
)

const (
	AscendingOrder Order = iota
	DescendingOrder
)

type Article struct {
	Id     ArticleId
	Guid   sql.NullString
	FeedId FeedId `db:"feed_id"`

	Title       string
	Description string
	Link        string
	Date        time.Time

	Read     bool
	Favorite bool
	Score    int64

	Hit struct {
		Fragments map[string][]string `json:"fragments,omitempty"`
	}
}

type ArticleScores struct {
	ArticleId ArticleId
	Score     int64
	Score1    int64
	Score2    int64
	Score3    int64
	Score4    int64
	Score5    int64
}

type ArticleFormatting struct {
	Content  string
	Title    string
	TopImage string
	Language string
}
