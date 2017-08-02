package extract

import "github.com/urandom/readeef/content"

type Generator interface {
	Generate(link string) (content.ArticleExtract, error)
}
