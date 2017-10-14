package readeef

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"net/url"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/feed"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
)

// TODO: split up this struct and modify the api

type FeedManager struct {
	config           config.Config
	repo             repo.Feed
	ops              chan func(context.Context, *FeedManager)
	log              log.Log
	hubbub           *Hubbub
	scheduler        feed.Scheduler
	parserProcessors []processor.Feed
}

var (
	commentPattern = regexp.MustCompile("<!--.*?-->")
	linkPattern    = regexp.MustCompile(`<link ([^>]+)>`)

	ErrNoFeed = errors.New("Feed not found")

	httpStatusPrefix = "HTTP Status: "
)

func NewFeedManager(repo repo.Feed, c config.Config, l log.Log) *FeedManager {
	return &FeedManager{
		repo: repo, config: c, log: l,
		ops:       make(chan func(context.Context, *FeedManager)),
		scheduler: feed.NewScheduler(l),
	}
}

func (fm *FeedManager) SetHubbub(hubbub *Hubbub) {
	fm.hubbub = hubbub
}

func (fm *FeedManager) AddFeedProcessor(p processor.Feed) {
	fm.parserProcessors = append(fm.parserProcessors, p)
}

func (fm *FeedManager) Start(ctx context.Context) error {
	fm.log.Infoln("Starting the feed manager")

	go fm.loop(ctx)
	go fm.scheduler.Start(ctx)

	feeds, err := fm.repo.Unsubscribed()
	if err != nil {
		return err
	}

	for _, f := range feeds {
		fm.log.Infoln("Scheduling feed " + f.String())

		fm.AddFeed(f)
	}

	return nil
}

func (fm *FeedManager) AddFeed(feed content.Feed) {
	fm.ops <- func(ctx context.Context, fm *FeedManager) {
		fm.startUpdatingFeed(ctx, feed)
	}
}

func (fm *FeedManager) RemoveFeed(feed content.Feed) {
	fm.ops <- func(ctx context.Context, fm *FeedManager) {
		fm.stopUpdatingFeed(feed)
	}
}

func (fm *FeedManager) AddFeedByLink(link string) (content.Feed, error) {
	u, err := url.Parse(link)
	if err == nil {
		if !u.IsAbs() {
			return content.Feed{}, errors.New("link not absolute")
		}
		u.Fragment = ""
		link = u.String()
	} else {
		return content.Feed{}, err
	}

	f, err := fm.repo.FindByLink(link)
	if err != nil && !content.IsNoContent(err) {
		return f, err
	}

	if err != nil {
		fm.log.Infoln("Discovering feeds in " + link)

		parsedFeeds, err := feed.Search(link, fm.log)
		if err != nil {
			return content.Feed{}, errors.WithMessage(err, "searching for feeds")
		}

		for link, parserFeed := range parsedFeeds {
			f.Link = link
			f.Refresh(fm.processParserFeed(parserFeed))

			break
		}

		if _, err = fm.repo.Update(&f); err != nil {
			return content.Feed{}, errors.WithMessage(err, "updating feed with parsed data")
		}
	}

	fm.log.Infoln("Adding feed " + f.String() + " to manager")
	fm.AddFeed(f)

	return f, nil
}

func (fm *FeedManager) RemoveFeedByLink(link string) (content.Feed, error) {
	f, err := fm.repo.FindByLink(link)
	if err != nil && !content.IsNoContent(err) {
		return f, err
	}

	if f.Validate() != nil {
		return f, nil
	}

	fm.log.Infoln("Removing feed " + f.String() + " from manager")

	fm.RemoveFeed(f)

	return f, nil
}

func (fm *FeedManager) DiscoverFeeds(link string) ([]content.Feed, error) {

	parsedFeeds, err := feed.Search(link, fm.log)
	if err != nil {
		return []content.Feed{}, errors.WithMessage(err, "discovering feeds")
	}
	feeds := make([]content.Feed, 0, len(parsedFeeds))

	for link, parserFeed := range parsedFeeds {
		feed := content.Feed{Link: link}
		feed.Refresh(parserFeed)

		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func (fm *FeedManager) loop(ctx context.Context) {
	for {
		select {
		case op := <-fm.ops:
			op(ctx, fm)
		case <-ctx.Done():
			return
		}
	}
}

func (fm *FeedManager) startUpdatingFeed(ctx context.Context, feed content.Feed) {
	if feed.HubLink != "" && fm.hubbub != nil {
		err := fm.hubbub.Subscribe(feed)

		if err != nil && err != ErrSubscribed {
			fm.log.Printf("Error subscribing to feed hublink: %+v\n", err)
		}
	}

	d := 30 * time.Minute
	if fm.config.FeedManager.Converted.UpdateInterval != 0 {
		if feed.TTL != 0 && feed.TTL > fm.config.FeedManager.Converted.UpdateInterval {
			d = feed.TTL
		} else {
			d = fm.config.FeedManager.Converted.UpdateInterval
		}
	}

	go fm.scheduleFeed(ctx, feed, d)
}

func (fm *FeedManager) scheduleFeed(ctx context.Context, feed content.Feed, update time.Duration) {
	fm.log.Infof("Scheduling update of feed %s", feed)
	for update := range fm.scheduler.ScheduleFeed(ctx, feed, update) {
		fm.log.Infof("Update for feed %s", feed)
		if update.IsErr() {
			feed.AddUpdateError(fmt.Sprintf("%s: %s", time.Now().Format(time.UnixDate), update.Error()))
		} else {
			feed.Refresh(fm.processParserFeed(update.Feed))
		}

		fm.updateFeed(feed)
	}
}

func (fm *FeedManager) stopUpdatingFeed(feed content.Feed) {
	if feed.HubLink != "" && fm.hubbub != nil {
		fm.hubbub.Unsubscribe(feed)
	}

	fm.log.Infoln("Stopping feed update for " + feed.Link)

	users, err := fm.repo.Users(feed)
	if err != nil {
		fm.log.Printf("Error getting users for feed '%s': %v\n", feed, err)
	} else {
		if len(users) == 0 {
			fm.log.Infoln("Removing orphan feed " + feed.String() + " from the database")

			if err = fm.repo.Delete(feed); err != nil {
				fm.log.Printf("Error deleting feed '%s' from the repository: %v\n", feed, err)
			}
		}
	}
}

func (fm FeedManager) updateFeed(feed content.Feed) {
	if _, err := fm.repo.Update(&feed); err != nil {
		fm.log.Printf("Error updating feed '%s' database record: %+v", feed, err)
	}

}

func (fm FeedManager) processParserFeed(pf parser.Feed) parser.Feed {
	for _, p := range fm.parserProcessors {
		pf = p.ProcessFeed(pf)
	}

	return pf
}
