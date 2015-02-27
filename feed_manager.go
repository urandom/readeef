package readeef

import (
	"database/sql"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/popularity"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/util"
)
import (
	"net/http"
	"net/url"
)

type FeedManager struct {
	*UpdateFeedReceiverManager
	config       Config
	repo         content.Repo
	addFeed      chan content.Feed
	removeFeed   chan content.Feed
	scoreArticle chan content.ScoredArticle
	done         chan bool
	client       *http.Client
	logger       webfw.Logger
	activeFeeds  map[info.FeedId]bool
	hubbub       *Hubbub
	searchIndex  SearchIndex
}

var (
	commentPattern = regexp.MustCompile("<!--.*?-->")
	linkPattern    = regexp.MustCompile(`<link ([^>]+)>`)

	ErrNoAbsolute = errors.New("Feed link is not absolute")
	ErrNoFeed     = errors.New("Feed not found")

	httpStatusPrefix = "HTTP Status: "
)

func NewFeedManager(repo content.Repo, c Config, l webfw.Logger, um *UpdateFeedReceiverManager) *FeedManager {
	return &FeedManager{
		UpdateFeedReceiverManager: um,
		repo: repo, config: c, logger: l,
		addFeed: make(chan content.Feed, 2), removeFeed: make(chan content.Feed, 2),
		scoreArticle: make(chan content.ScoredArticle), done: make(chan bool),
		activeFeeds: map[info.FeedId]bool{},
		client:      NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite)}
}

func (fm *FeedManager) SetHubbub(hubbub *Hubbub) {
	fm.hubbub = hubbub
}

func (fm *FeedManager) SetSearchIndex(si SearchIndex) {
	fm.searchIndex = si
}

func (fm *FeedManager) SetClient(c *http.Client) {
	fm.client = c
}

func (fm FeedManager) Start() {
	fm.logger.Infoln("Starting the feed manager")

	go fm.reactToChanges()

	go fm.scheduleFeeds()

	go fm.scoreArticles()
}

func (fm *FeedManager) Stop() {
	fm.logger.Infoln("Stopping the feed manager")

	fm.done <- true
}

func (fm *FeedManager) AddFeed(f content.Feed) {
	if f.Info().HubLink != "" && fm.hubbub != nil {
		err := fm.hubbub.Subscribe(f)

		if err == nil || err == ErrSubscribed {
			return
		}
	}

	fm.addFeed <- f
}

func (fm *FeedManager) RemoveFeed(f content.Feed) {
	fm.removeFeed <- f
}

func (fm *FeedManager) AddFeedByLink(link string) (content.Feed, error) {
	if u, err := url.Parse(link); err == nil {
		if !u.IsAbs() {
			return nil, ErrNoAbsolute
		}
		u.Fragment = ""
		link = u.String()
	} else {
		return nil, err
	}

	f := fm.repo.FeedByLink(link)
	if fm.repo.HasErr() {
		err := fm.repo.Err()
		if err == sql.ErrNoRows {
			fm.logger.Infoln("Discovering feeds in " + link)

			feeds, err := fm.discoverParserFeeds(link)
			if err != nil {
				return nil, err
			}

			f = feeds[0]

			f.Update()
			if f.HasErr() {
				return f, f.Err()
			}

			if fm.searchIndex != EmptySearchIndex {
				go func() {
					fm.searchIndex.UpdateFeed(f)
				}()
			}
		} else {
			return f, err
		}
	}

	fm.logger.Infoln("Adding feed " + f.String() + " to manager")
	fm.AddFeed(f)

	return f, nil
}

func (fm *FeedManager) RemoveFeedByLink(link string) (content.Feed, error) {
	f := fm.repo.FeedByLink(link)
	if fm.repo.HasErr() {
		err := fm.repo.Err()
		if err == sql.ErrNoRows {
			return f, nil
		} else {
			return f, err
		}
	}

	fm.logger.Infoln("Removing feed " + f.String() + " from manager")

	fm.removeFeed <- f

	return f, nil
}

func (fm *FeedManager) DiscoverFeeds(link string) ([]content.Feed, error) {
	feeds := []content.Feed{}

	if u, err := url.Parse(link); err == nil {
		if !u.IsAbs() {
			return feeds, ErrNoAbsolute
		}
		link = u.String()
	} else {
		return feeds, err
	}

	f := fm.repo.FeedByLink(link)
	if !fm.repo.HasErr() {
		feeds = append(feeds, f)
	} else {
		err := fm.repo.Err()
		if err == sql.ErrNoRows {
			fm.logger.Infoln("Discovering feeds in " + link)

			discovered, err := fm.discoverParserFeeds(link)
			if err != nil {
				return feeds, err
			}

			for _, f := range discovered {
				feeds = append(feeds, f)
			}
		} else {
			return feeds, err
		}
	}

	return feeds, nil
}

func (fm FeedManager) AddFeedChannel() chan<- content.Feed {
	return fm.addFeed
}

func (fm FeedManager) RemoveFeedChannel() chan<- content.Feed {
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

func (fm *FeedManager) startUpdatingFeed(f content.Feed) {
	if f == nil {
		fm.logger.Infoln("No feed provided")
		return
	}

	info := f.Info()

	if info.Id == 0 || fm.activeFeeds[info.Id] {
		fm.logger.Infoln("Feed " + info.Link + " already active")
		return
	}

	d := 30 * time.Minute
	if fm.config.Updater.Converted.Interval != 0 {
		if info.TTL != 0 && info.TTL > fm.config.Updater.Converted.Interval {
			d = info.TTL
		} else {
			d = fm.config.Updater.Converted.Interval
		}
	}

	fm.activeFeeds[info.Id] = true

	go func() {
		fm.requestFeedContent(f)

		ticker := time.After(d)

		fm.logger.Infof("Starting feed scheduler for %s and duration %d\n", f, d)
	TICKER:
		for {
			select {
			case now := <-ticker:
				if !fm.activeFeeds[info.Id] {
					fm.logger.Infof("Feed '%s' no longer active\n", info.Link)
					break TICKER
				}

				if !info.SkipHours[now.Hour()] && !info.SkipDays[now.Weekday().String()] {
					fm.requestFeedContent(f)
				}

				ticker = time.After(d)
				fm.logger.Infof("New feed ticker for '%s' after %d\n", info.Link, d)
			case <-fm.done:
				fm.stopUpdatingFeed(f)
				return
			}
		}
	}()

	go fm.scoreFeedContent(f)
}

func (fm *FeedManager) stopUpdatingFeed(f content.Feed) {
	if f == nil {
		fm.logger.Infoln("No feed provided")
		return
	}

	info := f.Info()

	fm.logger.Infoln("Stopping feed update for " + info.Link)
	delete(fm.activeFeeds, info.Id)

	users := f.Users()
	if f.HasErr() {
		fm.logger.Printf("Error getting users for feed '%s': %v\n", f, f.Err())
	} else {
		if len(users) == 0 {
			fm.logger.Infoln("Removing orphan feed " + f.String() + " from the database")

			if fm.searchIndex != EmptySearchIndex {
				if err := fm.searchIndex.DeleteFeed(f); err != nil {
					fm.logger.Printf(
						"Error deleting articles for feed '%s' from the search index: %v\n",
						f, err)
				}
			}
			f.Delete()
			if f.HasErr() {
				fm.logger.Printf("Error deleting feed '%s' from the repository: %v\n", f, f.Err())
			}
		}
	}
}

func (fm *FeedManager) requestFeedContent(f content.Feed) {
	if f == nil {
		fm.logger.Infoln("No feed provided")
		return
	}

	info := f.Info()

	fm.logger.Infoln("Requesting feed content for " + f.String())

	resp, err := fm.client.Get(info.Link)

	if err != nil {
		info.UpdateError = err.Error()
	} else if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		info.UpdateError = httpStatusPrefix + strconv.Itoa(resp.StatusCode)
	} else {
		defer resp.Body.Close()
		info.UpdateError = ""

		buf := util.BufferPool.GetBuffer()
		defer util.BufferPool.Put(buf)

		if _, err := buf.ReadFrom(resp.Body); err == nil {
			if pf, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
				f.Refresh(pf)
			} else {
				info.UpdateError = err.Error()
			}
		} else {
			info.UpdateError = err.Error()
		}

	}

	if info.UpdateError != "" {
		fm.logger.Printf("Error updating feed '%s': %s\n", f, info.UpdateError)
	}

	f.Info(info)

	select {
	case <-fm.done:
		return
	default:
		f.Update()
		if f.HasErr() {
			fm.logger.Printf("Error updating feed '%s' database record: %v\n", f, f.Err())
		}

		if len(f.NewArticles()) > 0 {
			if fm.searchIndex != EmptySearchIndex {
				fm.searchIndex.UpdateFeed(f)
			}

			fm.logger.Infoln("New articles notification for " + f.String())

			fm.NotifyReceivers(f)
		} else {
			fm.logger.Infoln("No new articles for " + f.String())
		}

		return
	}
}

func (fm *FeedManager) scoreFeedContent(f content.Feed) {
	if f == nil {
		fm.logger.Infoln("No feed provided")
		return
	}

	info := f.Info()

	if len(fm.config.Popularity.Providers) == 0 {
		fm.logger.Infoln("No popularity providers configured")
		return
	}

	if !fm.activeFeeds[info.Id] {
		fm.logger.Infof("Feed '%s' no longer active for scoring\n", f)
		return
	}

	fm.logger.Infoln("Scoring feed content for " + f.String())

	articles := f.LatestArticles()
	if f.HasErr() {
		fm.logger.Printf("Error getting latest feed articles for '%s': %v\n", f, f.Err())
		return
	}

	for i := range articles {
		sa := fm.repo.ScoredArticle()
		sa.Info(articles[i].Info())
		fm.scoreArticle <- sa
	}

	fm.logger.Infoln("Done scoring feed content for " + f.String())

	select {
	case <-time.After(30 * time.Minute):
		go fm.scoreFeedContent(f)
	case <-fm.done:
		return
	}
}

func (fm *FeedManager) scoreArticles() {
	for {
		select {
		case sa := <-fm.scoreArticle:
			time.Sleep(fm.config.Popularity.Converted.Delay)

			ascc := make(chan content.ArticleScores)
			blankScores := fm.repo.ArticleScores()

			go func() {
				asc := sa.Scores()

				if sa.HasErr() {
					fm.logger.Printf("Error getting scores for article '%s': %v\n", sa, sa.Err())
					ascc <- blankScores
				} else {
					ascc <- asc
				}
			}()

			info := sa.Info()

			fm.logger.Infof("Scoring '%s' using %+v\n", sa, fm.config.Popularity.Providers)
			score, err := popularity.Score(info.Link, info.Description, fm.config.Popularity.Providers)
			if err != nil {
				fm.logger.Printf("Error getting article popularity: %v\n", err)
				continue
			}

			asc := <-ascc
			ai := asc.Info()

			if asc != blankScores {
				age := ageInDays(info.Date)
				switch age {
				case 0:
					ai.Score1 = score
				case 1:
					ai.Score2 = score - ai.Score1
				case 2:
					ai.Score3 = score - ai.Score1 - ai.Score2
				case 3:
					ai.Score4 = score - ai.Score1 - ai.Score2 - ai.Score3
				default:
					ai.Score5 = score - ai.Score1 - ai.Score2 - ai.Score3 - ai.Score4
				}

				asc.Info(ai)
				score := asc.Calculate()
				penalty := float64(time.Now().Unix()-info.Date.Unix()) / (60 * 60) * float64(age)

				if penalty > 0 {
					ai.Score = int64(float64(score) / penalty)
				} else {
					ai.Score = score
				}

				asc.Info(ai)
				asc.Update()
				if asc.HasErr() {
					fm.logger.Printf("Error updating article scores: %v\n", asc.Err())
				}
			}
		case <-fm.done:
			return
		}
	}
}

func (fm *FeedManager) scheduleFeeds() {
	feeds := fm.repo.AllUnsubscribedFeeds()
	if fm.repo.HasErr() {
		fm.logger.Printf("Error fetching unsubscribed feeds: %v\n", fm.repo.Err())
		return
	}

	for _, f := range feeds {
		fm.logger.Infoln("Scheduling feed " + f.String())

		fm.AddFeed(f)
	}
}

func (fm FeedManager) discoverParserFeeds(link string) ([]content.Feed, error) {
	resp, err := http.Get(link)
	if err != nil {
		return []content.Feed{}, err
	}
	defer resp.Body.Close()

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.ReadFrom(resp.Body)

	if parserFeed, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
		feed := fm.repo.Feed()

		feed.Info(info.Feed{Link: link})
		feed.Refresh(parserFeed)

		return []content.Feed{feed}, nil
	} else {
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

func ageInDays(published time.Time) int {
	now := time.Now()
	sub := now.Sub(published)
	return int(sub / (24 * time.Hour))
}
