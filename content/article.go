package content

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"time"
)

type ArticleID int64

type Article struct {
	ID     ArticleID      `json:"id"`
	Guid   sql.NullString `json:"-"`
	FeedID FeedID         `db:"feed_id" json:"feedID"`

	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Date        time.Time `json:"date"`

	Read          bool   `json:"read"`
	Favorite      bool   `json:"favorite"`
	Score         int64  `json:"score,omitempty"`
	Thumbnail     string `json:"thumbnail,omitempty"`
	ThumbnailLink string `db:"thumbnail_link" json:"thumbnailLink,omitempty"`

	IsNew bool `json:"-"`

	Hit struct {
		Fragments map[string][]string `json:"fragments,omitempty"`
	} `json:"hits"`
}

type ArticleExtract struct {
	ArticleID ArticleID
	Title     string
	Content   string
	TopImage  string `db:"top_image"`
	Language  string
}
type sortingField int
type sortingOrder int

const (
	DefaultSort sortingField = iota
	SortByID
	SortByDate
)

const (
	AscendingOrder sortingOrder = iota
	DescendingOrder
)

// QueryOpt is a single query option.
type QueryOpt struct {
	f func(*QueryOptions)
}

// QueryOptions is the full range of options for querying articles.
type QueryOptions struct {
	Limit           int
	Offset          int
	ReadOnly        bool
	UnreadOnly      bool
	UnreadFirst     bool
	FavoriteOnly    bool
	UntaggedOnly    bool
	IncludeScores   bool
	HighScoredFirst bool
	BeforeID        ArticleID
	AfterID         ArticleID
	BeforeDate      time.Time
	AfterDate       time.Time
	BeforeScore     int64
	AfterScore      int64
	IDs             []ArticleID
	FeedIDs         []FeedID
	Filters         []Filter

	SortField sortingField
	SortOrder sortingOrder
}

type Filter struct {
	TagID        TagID    `json:"tagID"`
	FeedIDs      []FeedID `json:"feedIDs"`
	InverseFeeds bool     `json:"inverseFeeds"`
	URLTerm      string   `json:"urlTerm"`
	TitleTerm    string   `json:"titleTerm"`
	InverseURL   bool     `json:"inverseURL"`
	InverseTitle bool     `json:"inverseTitle"`
}

func (f Filter) Valid() bool {
	return (f.URLTerm != "" || f.TitleTerm != "") && (f.TagID == 0 || len(f.FeedIDs) > 0)
}

// Paging sets the article query paging optons.
func Paging(limit, offset int) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.Limit = limit
		o.Offset = offset
	}}
}

// IDRange sets the valid article ids between the two provided.
func IDRange(after, before ArticleID) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.AfterID = after
		o.BeforeID = before
	}}
}

// IDS limits the query to the specified article ids.
func IDs(ids []ArticleID) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.IDs = ids
	}}
}

// FeedIDS limits the query to the specified feed ids.
func FeedIDs(ids []FeedID) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.FeedIDs = ids
	}}
}

// TimeRange sets the minimum and maximum times of returned articles.
func TimeRange(after, before time.Time) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.AfterDate = after
		o.BeforeDate = before
	}}
}

// ScoreRange sets the minimum and maximum scores of returned articles.
func ScoreRange(after, before int64) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.AfterScore = after
		o.BeforeScore = before
	}}
}

// Filters ads configurable filters to limit the query result.
func Filters(filters []Filter) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.Filters = filters
	}}
}

// Sorting sets the query result sorting.
func Sorting(field sortingField, order sortingOrder) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.SortField = field
		o.SortOrder = order
	}}
}

var (
	// ReadOnly sets the query for read articles.
	ReadOnly = QueryOpt{func(o *QueryOptions) {
		o.ReadOnly = true
	}}

	// UnreadOnly sets the query for unread articles.
	UnreadOnly = QueryOpt{func(o *QueryOptions) {
		o.UnreadOnly = true
	}}

	// UnreadFirst sets the query to return unread articles first.
	UnreadFirst = QueryOpt{func(o *QueryOptions) {
		o.UnreadFirst = true
	}}

	// FavoriteOnly sets the query for favorite articles.
	FavoriteOnly = QueryOpt{func(o *QueryOptions) {
		o.FavoriteOnly = true
	}}

	// UntaggedOnly sets the query for untagged articles.
	UntaggedOnly = QueryOpt{func(o *QueryOptions) {
		o.UntaggedOnly = true
	}}

	// IncludeScores sets the query to return articles' score information.
	IncludeScores = QueryOpt{func(o *QueryOptions) {
		o.IncludeScores = true
	}}

	// HighScoredFirst sets the query to return articles with high scores first.
	HighScoredFirst = QueryOpt{func(o *QueryOptions) {
		o.HighScoredFirst = true
	}}
)

// Apply applies the settings from the passed opts to the QueryOptions
func (o *QueryOptions) Apply(opts []QueryOpt) {
	for _, opt := range opts {
		opt.f(o)
	}
}

func (a Article) String() string {
	return fmt.Sprintf("%d: %s", a.ID, a.Title)
}

func (a Article) Validate() error {
	if a.ID == 0 {
		return NewValidationError(errors.New("no ID"))
	}

	if a.FeedID == 0 {
		return NewValidationError(errors.New("no feed ID"))
	}

	if a.Link == "" {
		return NewValidationError(errors.New("no link"))
	}

	if u, err := url.Parse(a.Link); err != nil || !u.IsAbs() {
		return NewValidationError(errors.New("no link"))
	}

	return nil
}

func (id *ArticleID) Scan(src interface{}) error {
	asInt, ok := src.(int64)
	if !ok {
		return fmt.Errorf("Scan source '%#v' (%T) was not of type int64 (ArticleId)", src, src)
	}

	*id = ArticleID(asInt)

	return nil
}

func (id ArticleID) Value() (driver.Value, error) {
	return int64(id), nil
}

func GetUserFilters(u User) []Filter {
	if filters, ok := u.ProfileData["filters"].([]Filter); ok {
		return filters
	}

	return nil
}
