package content

import (
	"database/sql"
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
	IDs             []ArticleID

	SortField sortingField
	SortOrder sortingOrder
}

// Paging sets the article query paging optons.
func Paging(limit, offset int) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.Limit = limit
		o.Offset = offset
	}}
}

// IDRange sets the minimum and maximum ids of returned articles.
func IDRange(before, after ArticleID) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.BeforeID = before
		o.AfterID = after
	}}
}

// IDS limits the query to the specified article ids.
func IDs(ids []ArticleID) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.IDs = ids
	}}
}

// TimeRange sets the minimum and maximum times of returned articles.
func TimeRange(before, after time.Time) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.BeforeDate = before
		o.AfterDate = after
	}}
}

// Sorting sets the query result sorting.
func Sorting(field sortingField, order sortingOrder) QueryOpt {
	return QueryOpt{func(o *QueryOptions) {
		o.SortingField = field
		o.SortingOrder = order
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

	// HighScoredFirst sets the query to return articles with high scores first.
	HighScoredFirst = QueryOpt{func(o *QueryOptions) {
		o.HighScoredFirst = true
	}}

	// UnreadOnly sets the query for unread articles.
	UnreadOnly = QueryOpt{func(o *QueryOptions) {
		o.UnreadOnly = true
	}}
)

// Apply applies the settings from the passed opts to the QueryOptions
func (o *QueryOptions) Apply(opts ...QueryOpt) {
	for _, opt := range opts {
		opt(o)
	}
}

/*
type ArticleSorting interface {
	// Resets the sorting
	DefaultSorting() ArticleSorting

	// Sorts by content id, if available
	SortingById() ArticleSorting
	// Sorts by date, if available
	SortingByDate() ArticleSorting
	// Reverse the order
	Reverse() ArticleSorting

	// Returns the current field
	Field(f ...data.SortingField) data.SortingField

	// Returns the order, as set by Reverse()
	Order(o ...data.Order) data.Order
}

type ArticleRepo interface {
	ArticleSorting

	Articles(...data.ArticleQueryOptions) []UserArticle
	Ids(...data.ArticleIdQueryOptions) []data.ArticleId
	Count(...data.ArticleCountOptions) int64
	ReadState(read bool, o ...data.ArticleUpdateStateOptions)
}

type ArticleSearch interface {
	Query(query string, sp SearchProvider, paging ...int) []UserArticle
}

type Article interface {
	Error
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Data(data ...data.Article) data.Article

	Validate() error

	Update()

	Thumbnail() ArticleThumbnail
	Extract() ArticleExtract
	Scores() ArticleScores
}

type UserArticle interface {
	Article
	UserRelated

	Read(read bool)
	Favorite(favorite bool)
}

type ArticleScores interface {
	Error
	RepoRelated

	fmt.Stringer

	Data(data ...data.ArticleScores) data.ArticleScores

	Validate() error
	Calculate() int64

	Update()
}

type ArticleThumbnail interface {
	Error
	RepoRelated

	fmt.Stringer

	Data(data ...data.ArticleThumbnail) data.ArticleThumbnail

	Validate() error

	Update()
}

type ArticleExtract interface {
	Error
	RepoRelated

	fmt.Stringer

	Data(data ...data.ArticleExtract) data.ArticleExtract

	Validate() error

	Update()
}

type ArticleProcessor interface {
	ProcessArticles(ua []UserArticle) []UserArticle
}
*/
