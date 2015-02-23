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
}

type ArticleSearch interface {
	Highlight(highlight string) ArticleSearch
	Query(query string) []UserArticle
}

type Article interface {
	Error

	fmt.Stringer

	Set(info info.Article) Article
	Info() info.Article

	Validate() error
}

type UserArticle interface {
	Article

	User() User

	Read(read bool) UserArticle
	Favorite(favorite bool) UserArticle
}

type ScoredArticle interface {
	UserArticle

	SetScores(asc info.ArticleScores) ScoredArticle
	Scores() info.ArticleScores
}
