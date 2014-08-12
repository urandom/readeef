package readeef

import (
	"database/sql"
	"log"
	"readeef/parser"
	"strconv"
	"time"

	"github.com/urandom/webfw/util"
)
import "net/http"

type FeedManager struct {
	config      Config
	db          DB
	feeds       []Feed
	addFeed     chan Feed
	removeFeed  chan Feed
	updateFeed  chan<- Feed
	done        chan bool
	client      *http.Client
	logger      *log.Logger
	activeFeeds map[int64]bool
}

func NewFeedManager(db DB, c Config, l *log.Logger, updateFeed chan<- Feed) *FeedManager {
	return &FeedManager{
		db: db, config: c, logger: l, updateFeed: updateFeed,
		addFeed: make(chan Feed, 2), removeFeed: make(chan Feed, 2), done: make(chan bool),
		activeFeeds: map[int64]bool{},
		client:      NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite)}
}

func (fm *FeedManager) SetClient(c *http.Client) {
	fm.client = c
}

func (fm FeedManager) Start() {
	go fm.reactToChanges()

	go fm.scheduleFeeds()
}

func (fm *FeedManager) Stop() {
	fm.done <- true
}

func (fm *FeedManager) AddFeed(f Feed) {
	fm.addFeed <- f
}

func (fm *FeedManager) RemoveFeed(f Feed) {
	fm.removeFeed <- f
}

func (fm *FeedManager) AddFeedByLink(link string) error {
	f, err := fm.db.GetFeedByLink(link)
	if err != nil {
		if err == sql.ErrNoRows {
			f = Feed{Feed: parser.Feed{Link: link}}
			f, err = fm.db.UpdateFeed(f)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	fm.addFeed <- f

	return nil
}

func (fm *FeedManager) RemoveFeedByLink(link string) error {
	f, err := fm.db.GetFeedByLink(link)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else {
			return err
		}
	}

	fm.removeFeed <- f

	return nil
}

func (fm FeedManager) AddFeedChannel() chan<- Feed {
	return fm.addFeed
}

func (fm FeedManager) removeFeedChannel() chan<- Feed {
	return fm.removeFeed
}

func (fm *FeedManager) reactToChanges() {
	for {
		select {
		case f := <-fm.addFeed:
			fm.startUpdatingFeed(f)
		case f := <-fm.removeFeed:
			fm.stopUpdatingFeed(f)
		case <-fm.done:
			return
		}
	}
}

func (fm *FeedManager) startUpdatingFeed(f Feed) {
	if f.Id == 0 || fm.activeFeeds[f.Id] {
		return
	}

	d := 30 * time.Minute
	if fm.config.Updater.Converted.Interval != 0 {
		if f.TTL != 0 && f.TTL > fm.config.Updater.Converted.Interval {
			d = f.TTL
		} else {
			d = fm.config.Updater.Converted.Interval
		}
	}

	fm.activeFeeds[f.Id] = true

	go func() {
		fm.requestFeedContent(f)

	ticker:
		for {
			select {
			case now := <-time.After(d):
				if !fm.activeFeeds[f.Id] {
					break ticker
				}

				if !f.SkipHours[now.Hour()] && !f.SkipDays[now.Weekday().String()] {
					fm.requestFeedContent(f)
				}
			case <-fm.done:
				fm.stopUpdatingFeed(f)
				return
			}
		}
	}()
}

func (fm *FeedManager) stopUpdatingFeed(f Feed) {
	delete(fm.activeFeeds, f.Id)
}

func (fm *FeedManager) requestFeedContent(f Feed) {
	resp, err := fm.client.Get(f.Link)

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
		fm.logger.Printf("Error updating feed: %s\n", f.UpdateError)
	}

	select {
	case <-fm.done:
		return
	default:
		if _, err := fm.db.UpdateFeed(f); err != nil {
			fm.logger.Printf("Error updating feed database record: %v\n", err)
		}

		fm.updateFeed <- f
	}
}

func (fm *FeedManager) scheduleFeeds() {
	feeds, err := fm.db.GetUnsubscribedFeed()
	if err != nil {
		fm.logger.Printf("Error fetching unsubscribed feeds: %v\n", err)
		return
	}

	for _, f := range feeds {
		fm.addFeed <- f
	}
}
