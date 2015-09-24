package data

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
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

	Read          bool
	Favorite      bool
	Score         int64
	Thumbnail     string `json:",omitempty"`
	ThumbnailLink string `db:"thumbnail_link" json:",omitempty"`

	IsNew bool `json:"-"`

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

type ArticleThumbnail struct {
	ArticleId ArticleId
	Thumbnail string
	Link      string
	Processed bool
}

type ArticleExtract struct {
	ArticleId ArticleId
	Title     string
	Content   string
	TopImage  string `db:"top_image"`
	Language  string
}

type ArticleQueryOptions struct {
	Limit           int
	Offset          int
	ReadOnly        bool
	UnreadOnly      bool
	UnreadFirst     bool
	FavoriteOnly    bool
	UntaggedOnly    bool
	IncludeScores   bool
	HighScoredFirst bool
	BeforeId        ArticleId
	AfterId         ArticleId
	BeforeDate      time.Time
	AfterDate       time.Time

	SkipProcessors        bool
	SkipSessionProcessors bool
}

type ArticleIdQueryOptions struct {
	UnreadOnly   bool
	FavoriteOnly bool
	UntaggedOnly bool
	BeforeId     ArticleId
	AfterId      ArticleId
	BeforeDate   time.Time
	AfterDate    time.Time
}

type ArticleCountOptions struct {
	UnreadOnly   bool
	FavoriteOnly bool
	UntaggedOnly bool
	BeforeId     ArticleId
	AfterId      ArticleId
	BeforeDate   time.Time
	AfterDate    time.Time
}

type ArticleUpdateStateOptions struct {
	FavoriteOnly bool
	UntaggedOnly bool
	BeforeId     ArticleId
	AfterId      ArticleId
	BeforeDate   time.Time
	AfterDate    time.Time
}

func (id *ArticleId) Scan(src interface{}) error {
	asInt, ok := src.(int64)
	if !ok {
		return fmt.Errorf("Scan source '%#v' (%T) was not of type int64 (ArticleId)", src, src)
	}

	*id = ArticleId(asInt)

	return nil
}

func (id ArticleId) Value() (driver.Value, error) {
	return int64(id), nil
}
