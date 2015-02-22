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
	Highlight(highlight bool) ArticleSearch
	Query(query string) []UserArticle
}

type Article interface {
	Error

	fmt.Stringer

	Set(info info.Article)
	Info() info.Article
}

type UserArticle interface {
	Article

	User() info.User

	Read(read bool)
	Favorite(favorite bool)
}

type ScoredArticle interface {
	UserArticle

	SetScores(asc info.ArticleScores)
	Scores() info.ArticleScores
}
