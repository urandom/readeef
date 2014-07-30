package readeef

import (
	"log"
	"time"
)
import "net/http"

type FeedUpdater struct {
	config      Config
	db          DB
	feeds       []Feed
	addFeed     chan Feed
	removeFeed  chan Feed
	updateFeed  chan<- Feed
	done        chan bool
	client      *http.Client
	logger      *log.Logger
	feedTickers map[string]*time.Ticker
}

func NewFeedUpdater(db DB, c Config, l *log.Logger, updateFeed chan<- Feed) FeedUpdater {
	return FeedUpdater{
		db: db, config: c, logger: l, updateFeed: updateFeed,
		addFeed: make(chan Feed, 2), removeFeed: make(chan Feed, 2), done: make(chan bool),
		feedTickers: map[string]*time.Ticker{},
		client:      NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite)}
}

func (fu FeedUpdater) Start() {
	go fu.reactToChanges()

	go fu.scheduleFeeds()
}

func (fu FeedUpdater) AddFeed(f Feed) {
	fu.addFeed <- f
}

func (fu FeedUpdater) RemoveFeed(f Feed) {
	fu.removeFeed <- f
}

func (fu FeedUpdater) AddFeedChannel() chan<- Feed {
	return fu.addFeed
}

func (fu FeedUpdater) removeFeedChannel() chan<- Feed {
	return fu.removeFeed
}

func (fu FeedUpdater) reactToChanges() {
	for {
		select {
		case f := <-fu.addFeed:
			fu.startUpdatingFeed(f)
		case f := <-fu.removeFeed:
			if t, ok := fu.feedTickers[f.Link]; ok {
				t.Stop()
				delete(fu.feedTickers, f.Link)
			}
		case <-fu.done:
			return
		}
	}
}

func (fu FeedUpdater) startUpdatingFeed(f Feed) {
	d := 30 * time.Minute
	if fu.config.Updater.Converted.Interval != 0 {
		if f.TTL != 0 && f.TTL > fu.config.Updater.Converted.Interval {
			d = f.TTL
		} else {
			d = fu.config.Updater.Converted.Interval
		}
	}

	ticker := time.NewTicker(d)

	fu.feedTickers[f.Link] = ticker

	go func() {
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				if !f.SkipHours[now.Hour()] && !f.SkipDays[now.Weekday().String()] {
					// start a goroutine to fetch the content
				}
			case <-fu.done:
				ticker.Stop()
				delete(fu.feedTickers, f.Link)
				return
			}
		}
	}()
}

func (fu FeedUpdater) scheduleFeeds() {
	feeds, err := fu.db.GetUnsubscribedFeed()
	if err != nil {
		fu.logger.Printf("Error fetching unsubscribed feeds: %v\n", err)
		return
	}

	for _, f := range feeds {
		fu.addFeed <- f
	}
}
