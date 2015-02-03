package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type Feed struct {
	fm *readeef.FeedManager
}

func NewFeed(fm *readeef.FeedManager) Feed {
	return Feed{fm}
}

type feed struct {
	Id             int64
	Title          string
	Description    string
	Link           string
	Image          parser.Image
	Articles       []readeef.Article
	UpdateError    string
	SubscribeError string
	Tags           []string
}

var (
	errTypeNoAbsolute = "error-no-absolute"
	errTypeNoFeed     = "error-no-feed"
)

func (con Feed) Patterns() []webfw.MethodIdentifierTuple {
	prefix := "/v:version/feed/"

	return []webfw.MethodIdentifierTuple{
		webfw.MethodIdentifierTuple{prefix, webfw.MethodGet, "list"},
		webfw.MethodIdentifierTuple{prefix, webfw.MethodPost, "add"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id", webfw.MethodDelete, "remove"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/tags", webfw.MethodGet | webfw.MethodPost, "tags"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/read/:timestamp", webfw.MethodPost, "read"},
		webfw.MethodIdentifierTuple{prefix + ":feed-id/articles/:limit/:offset/:newer-first/:unread-only", webfw.MethodGet, "articles"},

		webfw.MethodIdentifierTuple{prefix + "discover", webfw.MethodGet, "discover"},
		webfw.MethodIdentifierTuple{prefix + "opml", webfw.MethodPost, "opml"},
	}
}

func (con Feed) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := webfw.GetMultiPatternIdentifier(c, r)
		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		r.ParseForm()

		var resp responseError
		var feedId int64

		params := webfw.GetParams(c, r)

		if resp.err == nil {
			switch action {
			case "list":
				resp = listFeeds(db, user)
			case "discover":
				link := r.FormValue("url")
				resp = discoverFeeds(db, user, con.fm, link)
			case "opml":
				buf := util.BufferPool.GetBuffer()
				defer util.BufferPool.Put(buf)

				buf.ReadFrom(r.Body)

				resp = parseOpml(db, user, con.fm, buf.Bytes())
			case "add":
				links := r.Form["url"]
				resp = addFeed(db, user, con.fm, links)
			case "remove":
				if feedId, resp.err = strconv.ParseInt(params["feed-id"], 10, 64); resp.err == nil {
					resp = removeFeed(db, user, con.fm, feedId)
				}
			case "tags":
				if feedId, resp.err = strconv.ParseInt(params["feed-id"], 10, 64); resp.err == nil {
					if r.Method == "GET" {
						resp = getFeedTags(db, user, feedId)
					} else if r.Method == "POST" {
						decoder := json.NewDecoder(r.Body)

						tags := []string{}
						if resp.err = decoder.Decode(&tags); resp.err != nil && resp.err != io.EOF {
							break
						}

						resp.err = nil
						resp = setFeedTags(db, user, feedId, tags)
					}
				}
			case "read":
				var timestamp int64

				if timestamp, resp.err = strconv.ParseInt(params["timestamp"], 10, 64); resp.err == nil {
					resp = markFeedAsRead(db, user, params["feed-id"], timestamp)
				}
			case "articles":
				var limit, offset int

				if limit, resp.err = strconv.Atoi(params["limit"]); resp.err == nil {
					if offset, resp.err = strconv.Atoi(params["offset"]); resp.err == nil {
						resp = getFeedArticles(db, user, params["feed-id"], limit, offset,
							params["newer-first"] == "true", params["unread-only"] == "true")
					}
				}
			}
		}

		switch resp.err {
		case readeef.ErrNoAbsolute:
			resp.val["Error"] = true
			resp.val["ErrorType"] = errTypeNoAbsolute
			resp.err = nil
		case readeef.ErrNoFeed:
			resp.val["Error"] = true
			resp.val["ErrorType"] = errTypeNoFeed
			resp.err = nil
		}

		var b []byte
		if resp.err == nil {
			b, resp.err = json.Marshal(resp.val)
		}

		if resp.err == nil {
			w.Write(b)
		} else {
			webfw.GetLogger(c).Print(resp.err)

			w.WriteHeader(http.StatusInternalServerError)
		}

	})
}

func (con Feed) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}

func listFeeds(db readeef.DB, user readeef.User) (resp responseError) {
	resp = newResponse()

	var feeds []readeef.Feed
	if feeds, resp.err = db.GetUserTagsFeeds(user); resp.err != nil {
		return
	}

	respFeeds := []feed{}

	for _, f := range feeds {
		respFeeds = append(respFeeds, feed{
			Id: f.Id, Title: f.Title, Description: f.Description,
			Link: f.Link, Image: f.Image, Tags: f.Tags,
			UpdateError: f.UpdateError, SubscribeError: f.SubscribeError,
		})
	}

	resp.val["Feeds"] = respFeeds
	return
}

func discoverFeeds(db readeef.DB, user readeef.User, fm *readeef.FeedManager, link string) (resp responseError) {
	resp = newResponse()

	if u, err := url.Parse(link); err != nil {
		resp.err = readeef.ErrNoAbsolute
		resp.errType = errTypeNoAbsolute
		return
	} else if !u.IsAbs() {
		u.Scheme = "http"
		if u.Host == "" {
			parts := strings.SplitN(u.Path, "/", 2)
			u.Host = parts[0]
			if len(parts) > 1 {
				u.Path = "/" + parts[1]
			} else {
				u.Path = ""
			}
		}
		link = u.String()
	}

	feeds, err := fm.DiscoverFeeds(link)
	if err != nil {
		resp.val["Feeds"] = []feed{}
		return
	}

	var userFeeds []readeef.Feed
	if userFeeds, resp.err = db.GetUserFeeds(user); resp.err != nil {
		return
	}

	userFeedIdMap := make(map[int64]bool)
	userFeedLinkMap := make(map[string]bool)
	for _, f := range userFeeds {
		userFeedIdMap[f.Id] = true
		userFeedLinkMap[f.Link] = true

		u, err := url.Parse(f.Link)
		if err == nil && strings.HasPrefix(u.Host, "www.") {
			u.Host = u.Host[4:]
			userFeedLinkMap[u.String()] = true
		}
	}

	respFeeds := []feed{}
	for _, f := range feeds {
		if !userFeedIdMap[f.Id] && !userFeedLinkMap[f.Link] {
			respFeeds = append(respFeeds, feed{
				Id: f.Id, Title: f.Title, Description: f.Description,
				Link: f.Link, Image: f.Image,
			})
		}
	}

	resp.val["Feeds"] = respFeeds
	return
}

func parseOpml(db readeef.DB, user readeef.User, fm *readeef.FeedManager, data []byte) (resp responseError) {
	resp = newResponse()

	var opml parser.Opml
	if opml, resp.err = parser.ParseOpml(data); resp.err != nil {
		return
	}

	var userFeeds []readeef.Feed
	if userFeeds, resp.err = db.GetUserFeeds(user); resp.err != nil {
		return
	}

	userFeedMap := make(map[int64]bool)
	for _, f := range userFeeds {
		userFeedMap[f.Id] = true
	}

	var feeds []readeef.Feed
	for _, opmlFeed := range opml.Feeds {
		discovered, err := fm.DiscoverFeeds(opmlFeed.Url)
		if err != nil {
			continue
		}

		for _, f := range discovered {
			if !userFeedMap[f.Id] {
				if len(opmlFeed.Tags) > 0 {
					f.Link += "#" + strings.Join(opmlFeed.Tags, ",")
				}

				feeds = append(feeds, f)
			}
		}
	}

	respFeeds := []feed{}
	for _, f := range feeds {
		respFeeds = append(respFeeds, feed{
			Id: f.Id, Title: f.Title, Description: f.Description,
			Link: f.Link, Image: f.Image,
		})
	}
	resp.val["Feeds"] = respFeeds
	return
}

func addFeed(db readeef.DB, user readeef.User, fm *readeef.FeedManager, links []string) (resp responseError) {
	resp = newResponse()

	success := false

	for _, link := range links {
		var u *url.URL
		if u, resp.err = url.Parse(link); resp.err != nil {
			/* TODO: non-fatal error */
			return
		} else if !u.IsAbs() {
			/* TODO: non-fatal error */
			resp.err = errors.New("Feed has no link")
			return
		} else {
			var f readeef.Feed
			if f, resp.err = fm.AddFeedByLink(link); resp.err != nil {
				return
			}

			if f, resp.err = db.CreateUserFeed(user, f); resp.err != nil {
				return
			}

			tags := strings.SplitN(u.Fragment, ",", -1)
			if u.Fragment != "" && len(tags) > 0 {
				if resp.err = db.CreateUserFeedTag(f, tags...); resp.err != nil {
					return
				}
			}

			success = true
		}
	}

	resp.val["Success"] = success
	return
}

func removeFeed(db readeef.DB, user readeef.User, fm *readeef.FeedManager, id int64) (resp responseError) {
	resp = newResponse()

	var feed readeef.Feed
	if feed, resp.err = db.GetUserFeed(id, user); resp.err != nil {
		/* TODO: non-fatal error */
		return
	}

	if resp.err = db.DeleteUserFeed(feed); resp.err != nil {
		/* TODO: non-fatal error */
		return
	}

	fm.RemoveFeed(feed)

	resp.val["Success"] = true
	return
}

func getFeedTags(db readeef.DB, user readeef.User, id int64) (resp responseError) {
	resp = newResponse()

	var feed readeef.Feed
	if feed, resp.err = db.GetUserFeed(id, user); resp.err != nil {
		/* TODO: non-fatal error */
		return
	}

	resp.val["Tags"] = feed.Tags
	return
}

func setFeedTags(db readeef.DB, user readeef.User, id int64, tags []string) (resp responseError) {
	resp = newResponse()

	var feed readeef.Feed
	if feed, resp.err = db.GetUserFeed(id, user); resp.err != nil {
		/* TODO: non-fatal error */
		return
	}

	var current []string
	if current, resp.err = db.GetUserFeedTags(user, feed); resp.err != nil {
		return
	}

	if resp.err = db.DeleteUserFeedTag(feed, current...); resp.err != nil {
		return
	}

	if resp.err = db.CreateUserFeedTag(feed, tags...); resp.err != nil {
		return
	}

	resp.val["Success"] = true
	resp.val["Id"] = feed.Id
	return
}

func markFeedAsRead(db readeef.DB, user readeef.User, id string, timestamp int64) (resp responseError) {
	resp = newResponse()

	t := time.Unix(timestamp/1000, 0)

	switch {
	case id == "tag:__all__":
		if resp.err = db.MarkUserArticlesByDateAsRead(user, t, true); resp.err != nil {
			return
		}
	case id == "__favorite__" || strings.HasPrefix(id, "popular:"):
		// Favorites are assumbed to have been read already
	case strings.HasPrefix(id, "tag:"):
		tag := id[4:]
		if resp.err = db.MarkUserTagArticlesByDateAsRead(user, tag, t, true); resp.err != nil {
			return
		}
	default:
		var feedId int64
		if feedId, resp.err = strconv.ParseInt(id, 10, 64); resp.err != nil {
			/* TODO: non-fatal error */
			return
		}

		var feed readeef.Feed
		if feed, resp.err = db.GetUserFeed(feedId, user); resp.err != nil {
			/* TODO: non-fatal error */
			return
		}

		if resp.err = db.MarkFeedArticlesByDateAsRead(feed, t, true); resp.err != nil {
			return
		}
	}

	resp.val["Success"] = true
	return
}

func getFeedArticles(db readeef.DB, user readeef.User, id string, limit int, offset int, newerFirst bool, unreadOnly bool) (resp responseError) {
	resp = newResponse()
	var articles []readeef.Article

	if limit > 50 {
		limit = 50
	}

	if id == "__favorite__" {
		if newerFirst {
			articles, resp.err = db.GetUserFavoriteArticlesDesc(user, limit, offset)
		} else {
			articles, resp.err = db.GetUserFavoriteArticles(user, limit, offset)
		}
	} else if id == "popular:__all__" {
		timeRange := readeef.TimeRange{time.Now().AddDate(0, 0, -5), time.Now()}
		if newerFirst {
			articles, resp.err = db.GetScoredUserArticlesDesc(user, timeRange, limit, offset)
		} else {
			articles, resp.err = db.GetScoredUserArticles(user, timeRange, limit, offset)
		}
	} else if id == "tag:__all__" {
		if newerFirst {
			if unreadOnly {
				articles, resp.err = db.GetUnreadUserArticlesDesc(user, limit, offset)
			} else {
				articles, resp.err = db.GetUserArticlesDesc(user, limit, offset)
			}
		} else {
			if unreadOnly {
				articles, resp.err = db.GetUnreadUserArticles(user, limit, offset)
			} else {
				articles, resp.err = db.GetUserArticles(user, limit, offset)
			}
		}
	} else if strings.HasPrefix(id, "popular:") {
		timeRange := readeef.TimeRange{time.Now().AddDate(0, 0, -5), time.Now()}

		if strings.HasPrefix(id, "popular:tag:") {
			tag := id[12:]

			if newerFirst {
				articles, resp.err = db.GetScoredUserTagArticlesDesc(user, tag, timeRange, limit, offset)
			} else {
				articles, resp.err = db.GetScoredUserTagArticles(user, tag, timeRange, limit, offset)
			}
		} else {
			var f readeef.Feed

			var feedId int64
			feedId, resp.err = strconv.ParseInt(id[8:], 10, 64)

			if resp.err != nil {
				resp.err = errors.New("Unknown feed id " + id)
				return
			}

			f, resp.err = db.GetFeed(feedId)
			if resp.err != nil {
				/* TODO: non-fatal error */
				return
			}

			f.User = user

			if newerFirst {
				f, resp.err = db.GetScoredFeedArticlesDesc(f, timeRange, limit, offset)
			} else {
				f, resp.err = db.GetScoredFeedArticles(f, timeRange, limit, offset)
			}

			if resp.err != nil {
				return
			}

			articles = f.Articles
		}
	} else if strings.HasPrefix(id, "tag:") {
		tag := id[4:]
		if newerFirst {
			if unreadOnly {
				articles, resp.err = db.GetUnreadUserTagArticlesDesc(user, tag, limit, offset)
			} else {
				articles, resp.err = db.GetUserTagArticlesDesc(user, tag, limit, offset)
			}
		} else {
			if unreadOnly {
				articles, resp.err = db.GetUnreadUserTagArticles(user, tag, limit, offset)
			} else {
				articles, resp.err = db.GetUserTagArticles(user, tag, limit, offset)
			}
		}
	} else {
		var f readeef.Feed

		var feedId int64
		feedId, resp.err = strconv.ParseInt(id, 10, 64)

		if resp.err != nil {
			resp.err = errors.New("Unknown feed id " + id)
			return
		}

		f, resp.err = db.GetFeed(feedId)
		if resp.err != nil {
			/* TODO: non-fatal error */
			return
		}

		f.User = user

		if newerFirst {
			if unreadOnly {
				f, resp.err = db.GetUnreadFeedArticlesDesc(f, limit, offset)
			} else {
				f, resp.err = db.GetFeedArticlesDesc(f, limit, offset)
			}
		} else {
			if unreadOnly {
				f, resp.err = db.GetUnreadFeedArticles(f, limit, offset)
			} else {
				f, resp.err = db.GetFeedArticles(f, limit, offset)
			}
		}
		if resp.err != nil {
			return
		}

		articles = f.Articles
	}

	if resp.err == nil {
		resp.val["Articles"] = articles
	}
	return
}
