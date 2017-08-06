package monitor

import "github.com/urandom/readeef/content"

// Feed imeplementations get notified for feed changes by the manager.
type Feed interface {
	FeedUpdated(content.Feed, []content.Article) error
	FeedDeleted(content.Feed) error
}
