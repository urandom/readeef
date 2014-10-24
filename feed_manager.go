package readeef

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/popularity"
	"github.com/urandom/webfw/util"
)
import (
	"net/http"
	"net/url"
)

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
	hubbub      *Hubbub
	searchIndex SearchIndex
}

var (
	commentPattern = regexp.MustCompile("<!--.*?-->")
	linkPattern    = regexp.MustCompile(`<link ([^>]+)>`)

	ErrNoAbsolute = errors.New("Feed link is not absolute")
	ErrNoFeed     = errors.New("Feed not found")

	httpStatusPrefix = "HTTP Status: "
)

func NewFeedManager(db DB, c Config, l *log.Logger, updateFeed chan<- Feed) *FeedManager {
	return &FeedManager{
		db: db, config: c, logger: l, updateFeed: updateFeed,
		addFeed: make(chan Feed, 2), removeFeed: make(chan Feed, 2), done: make(chan bool),
		activeFeeds: map[int64]bool{},
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
	Debug.Println("Starting the feed manager")

	go fm.reactToChanges()

	go fm.scheduleFeeds()
}

func (fm *FeedManager) Stop() {
	Debug.Println("Stopping the feed manager")

	fm.done <- true
}

func (fm *FeedManager) AddFeed(f Feed) {
	if f.HubLink != "" && fm.hubbub != nil {
		err := fm.hubbub.Subscribe(f)

		if err == nil || err == ErrSubscribed {
			return
		}
	}

	fm.addFeed <- f
}

func (fm *FeedManager) RemoveFeed(f Feed) {
	fm.removeFeed <- f
}

func (fm *FeedManager) AddFeedByLink(link string) (Feed, error) {
	if u, err := url.Parse(link); err == nil {
		if !u.IsAbs() {
			return Feed{}, ErrNoAbsolute
		}
		u.Fragment = ""
		link = u.String()
	} else {
		return Feed{}, err
	}

	f, err := fm.db.GetFeedByLink(link)
	if err != nil {
		if err == sql.ErrNoRows {
			Debug.Println("Discovering feeds in " + link)

			feeds, err := discoverParserFeeds(link)
			if err != nil {
				return Feed{}, err
			}

			f = feeds[0]
			f, _, err = fm.db.UpdateFeed(f)
			if err != nil {
				return Feed{}, err
			}

			if fm.searchIndex != EmptySearchIndex {
				go func() {
					fm.searchIndex.UpdateFeed(f)
				}()
			}
		} else {
			return Feed{}, err
		}
	}

	Debug.Println("Adding feed " + f.Link + " to manager")
	fm.AddFeed(f)

	return f, nil
}

func (fm *FeedManager) RemoveFeedByLink(link string) (Feed, error) {
	f, err := fm.db.GetFeedByLink(link)
	if err != nil {
		if err == sql.ErrNoRows {
			return Feed{}, nil
		} else {
			return Feed{}, err
		}
	}

	Debug.Println("Removing feed " + f.Link + " from manager")

	fm.removeFeed <- f

	return f, nil
}

func (fm *FeedManager) DiscoverFeeds(link string) ([]Feed, error) {
	feeds := []Feed{}

	if u, err := url.Parse(link); err == nil {
		if !u.IsAbs() {
			return feeds, ErrNoAbsolute
		}
		link = u.String()
	} else {
		return feeds, err
	}

	f, err := fm.db.GetFeedByLink(link)
	if err == nil {
		feeds = append(feeds, f)
	} else {
		if err == sql.ErrNoRows {
			Debug.Println("Discovering feeds in " + link)

			discovered, err := discoverParserFeeds(link)
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

func (fm FeedManager) AddFeedChannel() chan<- Feed {
	return fm.addFeed
}

func (fm FeedManager) RemoveFeedChannel() chan<- Feed {
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
		Debug.Println("Feed " + f.Link + " already active")
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

		go fm.scoreFeedContent(f)

		ticker := time.After(d)
		scoreTicker := time.After(30 * time.Minute)

		Debug.Printf("Starting feed scheduler for %s and duration %d\n", f.Link, d)
	TICKER:
		for {
			select {
			case now := <-ticker:
				if !fm.activeFeeds[f.Id] {
					break TICKER
				}

				if !f.SkipHours[now.Hour()] && !f.SkipDays[now.Weekday().String()] {
					f = fm.requestFeedContent(f)
				}

				if f.UpdateError != "" && !strings.HasPrefix(f.UpdateError, httpStatusPrefix) {
					rand.Seed(time.Now().Unix())
					secs := rand.Intn(45-15) + 15
					ticker = time.After(time.Duration(secs) * time.Second)
				} else {
					ticker = time.After(d)
				}
			case <-scoreTicker:
				if !fm.activeFeeds[f.Id] {
					break TICKER
				}

				go func() {
					fm.scoreFeedContent(f)
					scoreTicker = time.After(30 * time.Minute)
				}()
			case <-fm.done:
				fm.stopUpdatingFeed(f)
				return
			}
		}
	}()
}

func (fm *FeedManager) stopUpdatingFeed(f Feed) {
	Debug.Println("Stopping feed update for " + f.Link)
	delete(fm.activeFeeds, f.Id)

	users, err := fm.db.GetFeedUsers(f)
	if err != nil {
		fm.logger.Printf("Error getting users for feed '%s': %v\n", f.Link, err)
	} else {
		if len(users) == 0 {
			Debug.Println("Removing orphan feed " + f.Link + " from the database")

			if fm.searchIndex != EmptySearchIndex {
				if err := fm.searchIndex.DeleteFeed(f); err != nil {
					fm.logger.Printf(
						"Error deleting articles for feed '%s' from the search index: %v\n",
						f.Link, err)
				}
			}
			fm.db.DeleteFeed(f)
		}
	}
}

func (fm *FeedManager) requestFeedContent(f Feed) Feed {
	Debug.Println("Requesting feed content for " + f.Link)

	resp, err := fm.client.Get(f.Link)

	if err != nil {
		f.UpdateError = err.Error()
	} else if resp.StatusCode != http.StatusOK {
		f.UpdateError = httpStatusPrefix + strconv.Itoa(resp.StatusCode)
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
		return f
	default:
		f, newArticles, err := fm.db.UpdateFeed(f)
		if err != nil {
			fm.logger.Printf("Error updating feed database record: %v\n", err)
		}

		if newArticles {
			if fm.searchIndex != EmptySearchIndex {
				fm.searchIndex.UpdateFeed(f)
			}

			fm.updateFeed <- f
		}

		return f
	}
}

func (fm *FeedManager) scoreFeedContent(f Feed) {
	Debug.Println("Scoring feed content for " + f.Link)

	articles, err := fm.db.GetLatestFeedArticles(f)
	if err != nil {
		fm.logger.Printf("Error getting latest feed articles: %v\n", err)
		return
	}

	for i := 0; i < len(articles); i++ {
		a := articles[i]
		ascc := make(chan ArticleScores)

		go func() {
			asc, err := fm.db.GetArticleScores(a)

			if err != nil {
				fm.logger.Printf("Error getting article scores: %v\n", err)
				ascc <- EmptyArticleScores
			} else {
				ascc <- asc
			}
		}()

		score, err := popularity.Score(a.Link, a.Description)
		if err != nil {
			fm.logger.Printf("Error getting article popularity: %v\n", err)
			continue
		}

		asc := <-ascc

		if asc != EmptyArticleScores {
			age := ageInDays(a.Date)
			switch age {
			case 0:
				asc.Score1 = score
			case 1:
				asc.Score2 = score - asc.Score1
			case 2:
				asc.Score3 = score - asc.Score1 - asc.Score2
			case 3:
				asc.Score4 = score - asc.Score1 - asc.Score2 - asc.Score3
			default:
				asc.Score5 = score - asc.Score1 - asc.Score2 - asc.Score3 - asc.Score4
			}

			score := asc.CalculateScore()
			penalty := float64(time.Now().Unix()-a.Date.Unix()) / (60 * 60) * float64(age)

			if penalty > 0 {
				asc.Score = int64(float64(score) / penalty)
			} else {
				asc.Score = score
			}

			err := fm.db.UpdateArticleScores(asc)
			if err != nil {
				fm.logger.Printf("Error updating article scores: %v\n", err)
			}
		}
	}
}

func (fm *FeedManager) scheduleFeeds() {
	feeds, err := fm.db.GetUnsubscribedFeed()
	if err != nil {
		fm.logger.Printf("Error fetching unsubscribed feeds: %v\n", err)
		return
	}

	for _, f := range feeds {
		Debug.Println("Scheduling feed " + f.Link)

		fm.AddFeed(f)
	}
}

func discoverParserFeeds(link string) ([]Feed, error) {
	resp, err := http.Get(link)
	if err != nil {
		return []Feed{}, err
	}

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.ReadFrom(resp.Body)

	if parserFeed, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
		feed := Feed{Link: link}
		feed = feed.UpdateFromParsed(parserFeed)

		return []Feed{feed}, nil
	} else {
		html := commentPattern.ReplaceAllString(buf.String(), "")
		links := linkPattern.FindAllStringSubmatch(html, -1)

		feeds := []Feed{}
		for _, l := range links {
			attrs := l[1]
			if strings.Contains(attrs, `"application/rss+xml"`) || strings.Contains(attrs, `'application/rss+xml'`) {
				index := strings.Index(attrs, "href=")
				attr := attrs[index+6:]
				index = strings.IndexByte(attr, attrs[index+5])
				href := attr[:index]

				if u, err := url.Parse(href); err != nil {
					return []Feed{}, ErrNoFeed
				} else {
					if !u.IsAbs() {
						l, _ := url.Parse(link)

						u.Scheme = l.Scheme

						if u.Host == "" {
							u.Host = l.Host
						}

						href = u.String()
					}

					Debug.Printf("Checking if '%s' is a valid feed link\n", href)

					fs, err := discoverParserFeeds(href)
					if err != nil {
						return []Feed{}, err
					}

					feeds = append(feeds, fs[0])
				}
			}
		}

		if len(feeds) != 0 {
			return feeds, nil
		}
	}

	return []Feed{}, ErrNoFeed
}

func ageInDays(published time.Time) int {
	now := time.Now()
	sub := now.Sub(published)
	return int(sub / (24 * time.Hour))
}
