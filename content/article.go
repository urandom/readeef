package content

import (
	"fmt"

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
	Highlight(highlight string)
	Query(query string) []UserArticle
}

type Article interface {
	Error
	RepoRelated

	fmt.Stringer

	Info(info ...info.Article) info.Article

	Validate() error
}

type UserArticle interface {
	Article
	UserRelated

	Read(read bool)
	Favorite(favorite bool)
}

type ScoredArticle interface {
	UserArticle

	Scores(asc ...info.ArticleScores) info.ArticleScores
}
