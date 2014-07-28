package readeef

import "log"
import "net/http"

type FeedUpdater struct {
	config     Config
	db         DB
	feeds      []Feed
	addFeed    <-chan Feed
	removeFeed <-chan Feed
	updateFeed chan<- Feed
	done       chan bool
	client     *http.Client
	logger     *log.Logger
}

func NewFeedUpdater(db DB, c Config, l *log.Logger) FeedUpdater {
	return FeedUpdater{db: db, config: c, done: make(chan bool), logger: l,
		client: NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite)}
}

func (fu FeedUpdater) Start() {
	go fu.reactToChanges()

	go fu.getFeeds()
}

func (fu FeedUpdater) reactToChanges() {
	for {
		select {
		case <-fu.addFeed:
		case <-fu.removeFeed:
		case <-fu.done:
			break
		}
	}
}

func (fu FeedUpdater) getFeeds() {
	feeds, err := fu.db.GetUnsubscribedFeed()
	if err != nil {
		fu.logger.Printf("Error fetching unsubscribed feeds: %v\n", err)
		return
	}

	for _, f := range feeds {
		fu.addFeed <- f
	}
}
