package content

import (
	"fmt"

	"github.com/blevesearch/bleve"
	"github.com/urandom/readeef/content/info"
)

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
	Field() info.SortingField

	// Returns the order, as set by Reverse()
	Order() info.Order
}

type ArticleSearch interface {
	UserRelated

	Highlight(highlight ...string) string
	Query(query string, index bleve.Index, paging ...int) []UserArticle
}

type Article interface {
	Error
	RepoRelated

	fmt.Stringer

	Info(info ...info.Article) info.Article

	Validate() error

	Format(templateDir, readabilityKey string) info.ArticleFormatting
}

type ScoredArticle interface {
	Article

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

	Info(info ...info.ArticleScores) info.ArticleScores

	Validate() error
	Calculate() int64

	Update()
}
