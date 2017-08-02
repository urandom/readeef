package content

type FeedMonitor interface {
	FeedUpdated(feed Feed) error
	FeedDeleted(feed Feed) error
}
