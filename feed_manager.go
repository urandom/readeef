package readeef

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"time"

	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/feed"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw/util"
)

// TODO: split up this struct and modify the api

type FeedManager struct {
	config           config.Config
	repo             content.Repo
	ops              chan func(context.Context, *FeedManager)
	log              Logger
	hubbub           *Hubbub
	scheduler        feed.Scheduler
	parserProcessors []parser.Processor
	feedMonitors     []content.FeedMonitor
}

var (
	commentPattern = regexp.MustCompile("<!--.*?-->")
	linkPattern    = regexp.MustCompile(`<link ([^>]+)>`)

	ErrNoAbsolute = errors.New("Feed link is not absolute")
	ErrNoFeed     = errors.New("Feed not found")

	httpStatusPrefix = "HTTP Status: "
)

func NewFeedManager(repo content.Repo, c config.Config, l Logger) *FeedManager {
	return &FeedManager{
		repo: repo, config: c, log: l,
		ops:       make(chan func(context.Context, *FeedManager)),
		scheduler: feed.NewScheduler(),
	}
}

func (fm *FeedManager) SetHubbub(hubbub *Hubbub) {
	fm.hubbub = hubbub
}

func (fm *FeedManager) AddParserProcessor(p parser.Processor) {
	fm.parserProcessors = append(fm.parserProcessors, p)
}

func (fm *FeedManager) AddFeedMonitor(m content.FeedMonitor) {
	fm.feedMonitors = append(fm.feedMonitors, m)
}

func (fm *FeedManager) Start(ctx context.Context) error {
	fm.log.Infoln("Starting the feed manager")

	go fm.loop(ctx)

	feeds := fm.repo.AllUnsubscribedFeeds()
	if fm.repo.HasErr() {
		return errors.WithMessage(fm.repo.Err(), "getting unsubscribed feeds")
	}

	for _, f := range feeds {
		fm.log.Infoln("Scheduling feed " + f.String())

		fm.AddFeed(f)
	}

	return nil
}

func (fm *FeedManager) AddFeed(f content.Feed) {
	fm.ops <- func(ctx context.Context, fm *FeedManager) {
		fm.startUpdatingFeed(ctx, f)
	}
}

func (fm *FeedManager) RemoveFeed(f content.Feed) {
	fm.ops <- func(ctx context.Context, fm *FeedManager) {
		fm.stopUpdatingFeed(f)
	}
}

func (fm *FeedManager) AddFeedByLink(link string) (content.Feed, error) {
	u, err := url.Parse(link)
	if err == nil {
		if !u.IsAbs() {
			return nil, ErrNoAbsolute
		}
		u.Fragment = ""
		link = u.String()
	} else {
		return nil, err
	}

	f := fm.repo.FeedByLink(link)
	err = f.Err()
	if err != nil && err != content.ErrNoContent {
		return f, err
	}

	if err != nil {
		fm.log.Infoln("Discovering feeds in " + link)

		feeds, err := fm.discoverSecureParserFeeds(u)

		if err != nil {
			return nil, err
		}

		f = feeds[0]

		f.Update()
		if f.HasErr() {
			return f, f.Err()
		}

		// Do not halt the adding process due to slow monitors
		go fm.processFeedUpdateMonitors(f)
	}

	fm.log.Infoln("Adding feed " + f.String() + " to manager")
	fm.AddFeed(f)

	return f, nil
}

func (fm *FeedManager) RemoveFeedByLink(link string) (content.Feed, error) {
	f := fm.repo.FeedByLink(link)
	if f.HasErr() {
		err := f.Err()
		if err == content.ErrNoContent {
			err = nil
		}
		return f, f.Err()
	}

	if f.Validate() != nil {
		return f, nil
	}

	fm.log.Infoln("Removing feed " + f.String() + " from manager")

	fm.RemoveFeed(f)

	return f, nil
}

func (fm *FeedManager) DiscoverFeeds(link string) ([]content.Feed, error) {
	feeds := []content.Feed{}

	u, err := url.Parse(link)
	if err == nil {
		if !u.IsAbs() {
			return feeds, ErrNoAbsolute
		}
		link = u.String()
	} else {
		return feeds, err
	}

	f := fm.repo.FeedByLink(link)
	err = f.Err()
	if err != nil && err != content.ErrNoContent {
		return feeds, f.Err()
	} else {
		if err != nil {
			fm.log.Debugln("Discovering feeds in " + link)

			discovered, err := fm.discoverSecureParserFeeds(u)

			if err != nil {
				return feeds, err
			}

			fm.log.Debugf("Discovered %d feeds in %s\n", len(discovered), link)
			feeds = append(feeds, discovered...)
		}
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

func (fm *FeedManager) startUpdatingFeed(ctx context.Context, f content.Feed) {
	data := f.Data()

	if data.HubLink != "" && fm.hubbub != nil {
		err := fm.hubbub.Subscribe(f)

		if err == nil || err == ErrSubscribed {
			return
		}
	}

	d := 30 * time.Minute
	if fm.config.FeedManager.Converted.UpdateInterval != 0 {
		if data.TTL != 0 && data.TTL > fm.config.FeedManager.Converted.UpdateInterval {
			d = data.TTL
		} else {
			d = fm.config.FeedManager.Converted.UpdateInterval
		}
	}

	go fm.scheduleFeed(ctx, f, d)
}

func (fm *FeedManager) scheduleFeed(ctx context.Context, feed content.Feed, update time.Duration) {
	for update := range fm.scheduler.ScheduleFeed(ctx, feed, update) {
		if update.IsErr() {
			data := feed.Data()

			data.UpdateError = update.Error()
			feed.Data(data)
		} else {
			feed.Refresh(fm.processParserFeed(update.Feed))
		}

		fm.updateFeed(feed)
	}
}

func (fm *FeedManager) stopUpdatingFeed(f content.Feed) {
	data := f.Data()

	if data.HubLink != "" && fm.hubbub != nil {
		fm.hubbub.Unsubscribe(f)
	}

	fm.log.Infoln("Stopping feed update for " + data.Link)

	users := f.Users()
	if f.HasErr() {
		fm.log.Printf("Error getting users for feed '%s': %v\n", f, f.Err())
	} else {
		if len(users) == 0 {
			fm.log.Infoln("Removing orphan feed " + f.String() + " from the database")

			for _, m := range fm.feedMonitors {
				if err := m.FeedDeleted(f); err != nil {
					fm.log.Printf(
						"Error invoking monitor '%s' on deleted feed '%s': %v\n",
						reflect.TypeOf(m), f, err)
				}
			}
			f.Delete()
			if f.HasErr() {
				fm.log.Printf("Error deleting feed '%s' from the repository: %v\n", f, f.Err())
			}
		}
	}
}

func (fm FeedManager) discoverSecureParserFeeds(u *url.URL) (feeds []content.Feed, err error) {
	if u.Scheme == "http" {
		fm.log.Debugln("Testing secure link of", u)

		u.Scheme = "https"
		feeds, err = fm.discoverParserFeeds(u.String())
		u.Scheme = "http"
	}

	if u.Scheme != "http" || err != nil {
		feeds, err = fm.discoverParserFeeds(u.String())
	}

	return
}

func (fm FeedManager) discoverParserFeeds(link string) ([]content.Feed, error) {
	fm.log.Debugf("Fetching feed link body %s\n", link)
	resp, err := http.Get(link)
	if err != nil {
		return []content.Feed{}, err
	}
	defer resp.Body.Close()

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.ReadFrom(resp.Body)

	if parserFeed, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
		fm.log.Debugf("Discovering link %s contains feed data\n", link)

		feed := fm.repo.Feed()

		feed.Data(data.Feed{Link: link})
		feed.Refresh(fm.processParserFeed(parserFeed))

		return []content.Feed{feed}, nil
	} else {
		fm.log.Debugf("Searching for html links within the discovering link %s\n", link)

		html := commentPattern.ReplaceAllString(buf.String(), "")
		links := linkPattern.FindAllStringSubmatch(html, -1)

		feeds := []content.Feed{}
		for _, l := range links {
			attrs := l[1]
			if strings.Contains(attrs, `"application/rss+xml"`) || strings.Contains(attrs, `'application/rss+xml'`) {
				index := strings.Index(attrs, "href=")
				attr := attrs[index+6:]
				index = strings.IndexByte(attr, attrs[index+5])
				href := attr[:index]

				if u, err := url.Parse(href); err != nil {
					return []content.Feed{}, ErrNoFeed
				} else {
					if !u.IsAbs() {
						l, _ := url.Parse(link)

						u.Scheme = l.Scheme

						if u.Host == "" {
							u.Host = l.Host
						}

						href = u.String()
					}

					fs, err := fm.discoverParserFeeds(href)
					if err != nil {
						return []content.Feed{}, err
					}

					feeds = append(feeds, fs[0])
				}
			}
		}

		if len(feeds) != 0 {
			return feeds, nil
		}
	}

	return []content.Feed{}, ErrNoFeed
}

func (fm FeedManager) updateFeed(f content.Feed) {
	f.Update()

	if f.HasErr() {
		fm.log.Printf("Error updating feed '%s' database record: %v\n", f, f.Err())
	} else {
		fm.processFeedUpdateMonitors(f)
	}
}

func (fm FeedManager) processFeedUpdateMonitors(f content.Feed) {
	if len(f.NewArticles()) > 0 {
		for _, m := range fm.feedMonitors {
			if err := m.FeedUpdated(f); err != nil {
				fm.log.Printf("Error invoking monitor '%s' on updated feed '%s': %v\n",
					reflect.TypeOf(m), f, err)
			}
		}
	} else {
		fm.log.Infoln("No new articles for " + f.String())
	}
}

func (fm FeedManager) processParserFeed(pf parser.Feed) parser.Feed {
	for _, p := range fm.parserProcessors {
		pf = p.Process(pf)
	}

	return pf
}
