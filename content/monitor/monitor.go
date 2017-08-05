package monitor

// Feed imeplementations get notified for feed changes by the manager.
type Feed interface {
	FeedUpdated(feed Feed) error
	FeedDeleted(feed Feed) error
}
