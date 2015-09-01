package content

import "github.com/urandom/readeef/content/data"

type Extractor interface {
	Extract(link string) (data.ArticleExtract, error)
}

type Thumbnailer interface {
	Process(articles []Article) error
}

type SearchProvider interface {
	ArticleSorting
	IsNewIndex() bool
	IndexAllArticles(repo Repo) error
	UpdateFeed(feed Feed)
	DeleteFeed(feed Feed) error
	Search(term string, u User, feedIds []data.FeedId, limit, offset int) ([]UserArticle, error)
}
