package readeef

import (
	"log"
	"readeef/parser"
	"strconv"
	"time"

	"github.com/urandom/webfw/util"
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

func (fu FeedUpdater) SetClient(c *http.Client) {
	fu.client = c
}

func (fu FeedUpdater) Start() {
	go fu.reactToChanges()

	go fu.scheduleFeeds()
}

func (fu FeedUpdater) Stop() {
	fu.done <- true
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
			fu.stopUpdatingFeed(f)
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
		go fu.requestFeedContent(f)

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				if !f.SkipHours[now.Hour()] && !f.SkipDays[now.Weekday().String()] {
					go fu.requestFeedContent(f)
				}
			case <-fu.done:
				fu.stopUpdatingFeed(f)
				return
			}
		}
	}()
}

func (fu FeedUpdater) stopUpdatingFeed(f Feed) {
	if t, ok := fu.feedTickers[f.Link]; ok {
		t.Stop()
		delete(fu.feedTickers, f.Link)
	}
}

func (fu FeedUpdater) requestFeedContent(f Feed) {
	resp, err := fu.client.Get(f.Link)

	if err != nil {
		f.UpdateError = err.Error()
	} else if resp.StatusCode != http.StatusOK {
		f.UpdateError = "HTTP Status: " + strconv.Itoa(resp.StatusCode)
	} else {
		f.UpdateError = ""

		buf := util.BufferPool.GetBuffer()
		defer util.BufferPool.Put(buf)

		if _, err := buf.ReadFrom(resp.Body); err == nil {
			if pf, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
				f = f.UpdateFromParsed(pf)
			} else {
				f.UpdateError = err.Error()
			}
		} else {
			f.UpdateError = err.Error()
		}

	}

	if f.UpdateError != "" {
		fu.logger.Printf("Error updating feed: %s\n", f.UpdateError)
	}

	select {
	case <-fu.done:
		return
	default:
		if err := fu.db.UpdateFeed(f); err != nil {
			fu.logger.Printf("Error updating feed database record: %v\n", err)
		}

		fu.updateFeed <- f
	}
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
