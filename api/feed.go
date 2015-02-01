package api

import (
	"encoding/json"
	"errors"
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

func (con Feed) Patterns() map[string]webfw.MethodIdentifierTuple {
	prefix := "/v:version/feed/"

	return map[string]webfw.MethodIdentifierTuple{
		prefix:                     webfw.MethodIdentifierTuple{webfw.MethodGet, "list"},
		prefix + "discover":        webfw.MethodIdentifierTuple{webfw.MethodGet, "discover"},
		prefix + "opml":            webfw.MethodIdentifierTuple{webfw.MethodPost, "opml"},
		prefix + "add":             webfw.MethodIdentifierTuple{webfw.MethodPost, "add"},
		prefix + "remove/:feed-id": webfw.MethodIdentifierTuple{webfw.MethodPost, "remove"},
		prefix + "tags/:feed-id":   webfw.MethodIdentifierTuple{webfw.MethodGet | webfw.MethodPost, "tags"},

		prefix + "read/:feed-id/:timestamp": webfw.MethodIdentifierTuple{webfw.MethodPost, "read"},

		prefix + "articles/:feed-id/:limit/:offset/:newer-first/:unread-only": webfw.MethodIdentifierTuple{webfw.MethodGet, "articles"},
	}
}

func (con Feed) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := webfw.GetMultiPatternIdentifier(c, r)

		var resp responseError
		switch action {
		case "list":
			resp = listFeeds(c, r)
		case "discover":
			resp = discoverFeeds(c, r, con.fm)
		case "opml":
			resp = parseOpml(c, r, con.fm)
		case "add":
			resp = addFeed(c, r, con.fm)
		case "remove":
			resp = removeFeed(c, r, con.fm)
		case "tags":
			resp = feedTags(c, r)
		case "read":
			resp = markFeedAsRead(c, r)
		case "articles":
			resp = getFeedArticles(c, r)
		}

		switch resp.err {
		case readeef.ErrNoAbsolute:
			resp.val["Error"] = true
			resp.val["ErrorType"] = "error-no-absolute"
			resp.err = nil
		case readeef.ErrNoFeed:
			resp.val["Error"] = true
			resp.val["ErrorType"] = "error-no-feed"
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

func listFeeds(c context.Context, r *http.Request) responseError {
	resp := newResponse()
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)

	feeds, err := db.GetUserTagsFeeds(user)

	if err != nil {
		resp.err = err
		return resp
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
	return resp
}

func discoverFeeds(c context.Context, r *http.Request, fm *readeef.FeedManager) responseError {
	resp := newResponse()
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)

	r.ParseForm()

	link := r.FormValue("url")

	if u, err := url.Parse(link); err != nil {
		resp.err = readeef.ErrNoAbsolute
		return resp
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
		return resp
	}

	userFeeds, err := db.GetUserFeeds(user)
	if err != nil {
		resp.err = err
		return resp
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
	return resp
}

func parseOpml(c context.Context, r *http.Request, fm *readeef.FeedManager) responseError {
	resp := newResponse()
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)
	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.ReadFrom(r.Body)

	opml, err := parser.ParseOpml(buf.Bytes())
	if err != nil {
		resp.err = err
		return resp
	}

	userFeeds, err := db.GetUserFeeds(user)
	if err != nil {
		resp.err = err
		return resp
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
	return resp
}

func addFeed(c context.Context, r *http.Request, fm *readeef.FeedManager) responseError {
	resp := newResponse()
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)

	r.ParseForm()
	links := r.Form["url"]
	success := false

	for _, link := range links {
		if u, err := url.Parse(link); err != nil {
			/* TODO: non-fatal error */
			resp.err = err
			return resp
		} else if !u.IsAbs() {
			/* TODO: non-fatal error */
			resp.err = errors.New("Feed has no link")
			return resp
		} else {

			f, err := fm.AddFeedByLink(link)
			if err != nil {
				resp.err = err
				return resp
			}

			f, err = db.CreateUserFeed(user, f)
			if err != nil {
				resp.err = err
				return resp
			}

			tags := strings.SplitN(u.Fragment, ",", -1)
			if u.Fragment != "" && len(tags) > 0 {
				resp.err = db.CreateUserFeedTag(f, tags...)
				return resp
			}

			success = true
		}
	}

	resp.val["Success"] = success
	return resp
}

func removeFeed(c context.Context, r *http.Request, fm *readeef.FeedManager) responseError {
	resp := newResponse()
	params := webfw.GetParams(c, r)
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)

	id, err := strconv.ParseInt(params["feed-id"], 10, 64)
	/* TODO: non-fatal error */
	if err != nil {
		resp.err = err
		return resp
	}

	feed, err := db.GetUserFeed(id, user)
	/* TODO: non-fatal error */
	if err != nil {
		resp.err = err
		return resp
	}

	if resp.err = db.DeleteUserFeed(feed); resp.err != nil {
		/* TODO: non-fatal error */
		return resp
	}

	fm.RemoveFeed(feed)

	resp.val["Success"] = true
	return resp
}

func feedTags(c context.Context, r *http.Request) responseError {
	resp := newResponse()
	params := webfw.GetParams(c, r)
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)

	id, err := strconv.ParseInt(params["feed-id"], 10, 64)
	if err != nil {
		/* TODO: non-fatal error */
		resp.err = err
		return resp
	}

	feed, err := db.GetUserFeed(id, user)
	if err != nil {
		/* TODO: non-fatal error */
		resp.err = err
		return resp
	}

	if r.Method == "GET" {
		resp.val["Tags"] = feed.Tags
	} else if r.Method == "POST" {
		tags, err := db.GetUserFeedTags(user, feed)
		if err != nil {
			resp.err = err
			return resp
		}

		if resp.err = db.DeleteUserFeedTag(feed, tags...); resp.err != nil {
			return resp
		}

		decoder := json.NewDecoder(r.Body)

		tags = []string{}
		if resp.err = decoder.Decode(&tags); resp.err != nil {
			return resp
		}

		if resp.err = db.CreateUserFeedTag(feed, tags...); resp.err != nil {
			return resp
		}

		resp.val["Success"] = true
		resp.val["Id"] = feed.Id
	}

	return resp
}

func markFeedAsRead(c context.Context, r *http.Request) responseError {
	resp := newResponse()
	params := webfw.GetParams(c, r)
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)
	feedId := params["feed-id"]
	timestamp := params["timestamp"]

	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	/* TODO: non-fatal error */
	if err != nil {
		resp.err = err
		return resp
	}

	t := time.Unix(seconds/1000, 0)

	switch {
	case feedId == "tag:__all__":
		resp.err = db.MarkUserArticlesByDateAsRead(user, t, true)
	case feedId == "__favorite__" || strings.HasPrefix(feedId, "popular:"):
		// Favorites are assumbed to have been read already
	case strings.HasPrefix(feedId, "tag:"):
		tag := feedId[4:]
		resp.err = db.MarkUserTagArticlesByDateAsRead(user, tag, t, true)
	default:
		id, err := strconv.ParseInt(feedId, 10, 64)
		/* TODO: non-fatal error */
		if err != nil {
			resp.err = err
			return resp
		}

		feed, err := db.GetUserFeed(id, user)
		/* TODO: non-fatal error */
		if err != nil {
			resp.err = err
			return resp
		}

		resp.err = db.MarkFeedArticlesByDateAsRead(feed, t, true)
	}

	if resp.err == nil {
		resp.val["Success"] = true
	}

	return resp
}

func getFeedArticles(c context.Context, r *http.Request) responseError {
	resp := newResponse()
	params := webfw.GetParams(c, r)
	db := readeef.GetDB(c)
	user := readeef.GetUser(c, r)
	var articles []readeef.Article

	var limit, offset int

	feedId := params["feed-id"]

	limit, resp.err = strconv.Atoi(params["limit"])
	if resp.err != nil {
		/* TODO: non-fatal error */
		return resp
	}

	offset, resp.err = strconv.Atoi(params["offset"])
	/* TODO: non-fatal error */
	if resp.err != nil {
		return resp
	}

	newerFirst := params["newer-first"] == "true"
	unreadOnly := params["unread-only"] == "true"

	if limit > 50 {
		limit = 50
	}

	if feedId == "__favorite__" {
		if newerFirst {
			articles, resp.err = db.GetUserFavoriteArticlesDesc(user, limit, offset)
		} else {
			articles, resp.err = db.GetUserFavoriteArticles(user, limit, offset)
		}
	} else if feedId == "popular:__all__" {
		timeRange := readeef.TimeRange{time.Now().AddDate(0, 0, -5), time.Now()}
		if newerFirst {
			articles, resp.err = db.GetScoredUserArticlesDesc(user, timeRange, limit, offset)
		} else {
			articles, resp.err = db.GetScoredUserArticles(user, timeRange, limit, offset)
		}
	} else if feedId == "tag:__all__" {
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
	} else if strings.HasPrefix(feedId, "popular:") {
		timeRange := readeef.TimeRange{time.Now().AddDate(0, 0, -5), time.Now()}

		if strings.HasPrefix(feedId, "popular:tag:") {
			tag := feedId[12:]

			if newerFirst {
				articles, resp.err = db.GetScoredUserTagArticlesDesc(user, tag, timeRange, limit, offset)
			} else {
				articles, resp.err = db.GetScoredUserTagArticles(user, tag, timeRange, limit, offset)
			}
		} else {
			var f readeef.Feed

			var id int64
			id, resp.err = strconv.ParseInt(feedId[8:], 10, 64)

			if resp.err != nil {
				resp.err = errors.New("Unknown feed id " + feedId)
				return resp
			}

			f, resp.err = db.GetFeed(id)
			if resp.err != nil {
				/* TODO: non-fatal error */
				return resp
			}

			f.User = user

			if newerFirst {
				f, resp.err = db.GetScoredFeedArticlesDesc(f, timeRange, limit, offset)
			} else {
				f, resp.err = db.GetScoredFeedArticles(f, timeRange, limit, offset)
			}

			if resp.err != nil {
				return resp
			}

			articles = f.Articles
		}
	} else if strings.HasPrefix(feedId, "tag:") {
		tag := feedId[4:]
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

		var id int64
		id, resp.err = strconv.ParseInt(feedId, 10, 64)

		if resp.err != nil {
			resp.err = errors.New("Unknown feed id " + feedId)
			return resp
		}

		f, resp.err = db.GetFeed(id)
		if resp.err != nil {
			/* TODO: non-fatal error */
			return resp
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
			return resp
		}

		articles = f.Articles
	}

	if resp.err == nil {
		resp.val["Articles"] = articles
	}
	return resp
}
