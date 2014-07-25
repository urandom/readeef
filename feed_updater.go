package readeef

import "net/http"

type FeedUpdater struct {
	config     Config
	db         DB
	feeds      []Feed
	addFeed    <-chan Feed
	removeFeed <-chan Feed
	updateFeed chan<- Feed
	client     *http.Client
}

func NewFeedUpdater(db DB, c Config) FeedUpdater {
	return FeedUpdater{db: db, config: c,
		client: NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite)}
}

func (fu FeedUpdater) Start() {
	done := make(chan bool)
	go fu.reactToChanges(done)
}

func (fu FeedUpdater) reactToChanges(done <-chan bool) {
	for {
		select {
		case <-fu.addFeed:
		case <-fu.removeFeed:
		case <-done:
			break
		}
	}
}
