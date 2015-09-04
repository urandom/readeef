package content

import "github.com/urandom/readeef/content/data"

type Extractor interface {
	Extract(link string) (data.ArticleExtract, error)
}

type Thumbnailer interface {
	Generate(article Article) error
}

type SearchProvider interface {
	ArticleSorting
	IsNewIndex() bool
	IndexAllFeeds(repo Repo) error
	Search(term string, u User, feedIds []data.FeedId, limit, offset int) ([]UserArticle, error)
	BatchIndex(articles []Article, op data.IndexOperation) error
}

type FeedMonitor interface {
	FeedUpdated(feed Feed) error
	FeedDeleted(feed Feed) error
}
