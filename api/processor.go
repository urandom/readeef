package api

import "github.com/urandom/readeef/content"

type ArticleProcessor interface {
	ProcessArticles(ua []content.UserArticle) []content.UserArticle
}
