package content

import "github.com/urandom/readeef/content/data"

type Extractor interface {
	Extract(link string) (data.ArticleExtract, error)
}

type Thumbnailer interface {
	Process(articles []Article)
}
