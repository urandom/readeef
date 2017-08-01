package content

import "github.com/urandom/readeef/content/data"

type Extractor interface {
	Extract(link string) (data.ArticleExtract, error)
}

type Thumbnailer interface {
	Generate(article Article) error
}

type FeedMonitor interface {
	FeedUpdated(feed Feed) error
	FeedDeleted(feed Feed) error
}
