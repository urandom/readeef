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
	webfw.BaseController
	fm *readeef.FeedManager
}

func NewFeed(fm *readeef.FeedManager) Feed {
	return Feed{
		webfw.NewBaseController("", webfw.MethodGet, ""),
		fm,
	}
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
		prefix + "list":            webfw.MethodIdentifierTuple{webfw.MethodGet, "list"},
		prefix + "discover":        webfw.MethodIdentifierTuple{webfw.MethodGet, "discover"},
		prefix + "opml":            webfw.MethodIdentifierTuple{webfw.MethodPost, "opml"},
		prefix + "add":             webfw.MethodIdentifierTuple{webfw.MethodPost, "add"},
		prefix + "remove/:feed-id": webfw.MethodIdentifierTuple{webfw.MethodPost, "remove"},
		prefix + "tags/:feed-id":   webfw.MethodIdentifierTuple{webfw.MethodGet | webfw.MethodPost, "tags"},

		prefix + "read/:feed-id/:timestamp": webfw.MethodIdentifierTuple{webfw.MethodPost, "read"},

		prefix + "articles/:feed-id/:limit/:offset/:newer-first/:unread-only": webfw.MethodIdentifierTuple{webfw.MethodGet, "articles"},
	}
}

func (con Feed) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		action := webfw.GetMultiPatternIdentifier(c, r)
		params := webfw.GetParams(c, r)
		resp := make(map[string]interface{})

	SWITCH:
		switch action {
		case "list":
			var feeds []readeef.Feed
			feeds, err = db.GetUserTagsFeeds(user)

			if err != nil {
				break
			}

			respFeeds := []feed{}

			for _, f := range feeds {
				respFeeds = append(respFeeds, feed{
					Id: f.Id, Title: f.Title, Description: f.Description,
					Link: f.Link, Image: f.Image, Tags: f.Tags,
					UpdateError: f.UpdateError, SubscribeError: f.SubscribeError,
				})
			}

			resp["Feeds"] = respFeeds
		case "discover":
			r.ParseForm()

			link := r.FormValue("url")
			var u *url.URL

			if u, err = url.Parse(link); err != nil {
				err = readeef.ErrNoAbsolute
				break
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

			var feeds []readeef.Feed
			feeds, err = con.fm.DiscoverFeeds(link)
			if err != nil {
				break
			}

			var userFeeds []readeef.Feed
			userFeeds, err = db.GetUserFeeds(user)
			if err != nil {
				break
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

			resp["Feeds"] = respFeeds
		case "opml":
			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			buf.ReadFrom(r.Body)

			var opml parser.Opml
			opml, err = parser.ParseOpml(buf.Bytes())
			if err != nil {
				break
			}

			var userFeeds []readeef.Feed
			userFeeds, err = db.GetUserFeeds(user)
			if err != nil {
				break
			}

			userFeedMap := make(map[int64]bool)
			for _, f := range userFeeds {
				userFeedMap[f.Id] = true
			}

			var feeds []readeef.Feed
			for _, opmlFeed := range opml.Feeds {
				var discovered []readeef.Feed

				discovered, err = con.fm.DiscoverFeeds(opmlFeed.Url)
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
			resp["Feeds"] = respFeeds
		case "add":
			r.ParseForm()
			links := r.Form["url"]
			success := false

			for _, link := range links {
				/* TODO: non-fatal error */
				var u *url.URL
				if u, err = url.Parse(link); err != nil {
					break SWITCH
					/* TODO: non-fatal error */
				} else if !u.IsAbs() {
					err = errors.New("Feed has no link")
					break SWITCH
				}

				var f readeef.Feed
				f, err = con.fm.AddFeedByLink(link)
				if err != nil {
					break SWITCH
				}

				f, err = db.CreateUserFeed(readeef.GetUser(c, r), f)
				if err != nil {
					break SWITCH
				}

				tags := strings.SplitN(u.Fragment, ",", -1)
				if u.Fragment != "" && len(tags) > 0 {
					err = db.CreateUserFeedTag(f, tags...)
				}

				success = true
			}

			resp["Success"] = success
		case "remove":
			var id int64
			id, err = strconv.ParseInt(params["feed-id"], 10, 64)

			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			var feed readeef.Feed
			feed, err = db.GetUserFeed(id, user)
			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			err = db.DeleteUserFeed(feed)
			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			con.fm.RemoveFeed(feed)

			resp["Success"] = true
		case "tags":
			var id int64
			id, err = strconv.ParseInt(params["feed-id"], 10, 64)

			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			var feed readeef.Feed
			feed, err = db.GetUserFeed(id, user)
			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			if r.Method == "GET" {
				resp["Tags"] = feed.Tags
			} else if r.Method == "POST" {
				var tags []string

				tags, err = db.GetUserFeedTags(user, feed)
				if err != nil {
					break
				}

				err = db.DeleteUserFeedTag(feed, tags...)
				if err != nil {
					break
				}

				decoder := json.NewDecoder(r.Body)

				tags = []string{}
				err = decoder.Decode(&tags)
				if err != nil {
					break
				}

				err = db.CreateUserFeedTag(feed, tags...)
				if err != nil {
					break
				}

				resp["Success"] = true
				resp["Id"] = feed.Id
			}
		case "read":
			feedId := params["feed-id"]
			timestamp := params["timestamp"]

			var seconds int64
			seconds, err = strconv.ParseInt(timestamp, 10, 64)
			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			t := time.Unix(seconds/1000, 0)

			switch {
			case feedId == "tag:__all__":
				err = db.MarkUserArticlesByDateAsRead(user, t, true)
			case feedId == "__favorite__" || feedId == "__popular__":
				// Favorites are assumbed to have been read already
			case strings.HasPrefix(feedId, "tag:"):
				tag := feedId[4:]
				err = db.MarkUserTagArticlesByDateAsRead(user, tag, t, true)
			default:
				var id int64

				id, err = strconv.ParseInt(feedId, 10, 64)
				/* TODO: non-fatal error */
				if err != nil {
					break SWITCH
				}

				var feed readeef.Feed
				feed, err = db.GetUserFeed(id, user)
				/* TODO: non-fatal error */
				if err != nil {
					break SWITCH
				}

				err = db.MarkFeedArticlesByDateAsRead(feed, t, true)
			}

			if err == nil {
				resp["Success"] = true
			}
		case "articles":
			var articles []readeef.Article

			var limit, offset int

			feedId := params["feed-id"]

			limit, err = strconv.Atoi(params["limit"])
			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			offset, err = strconv.Atoi(params["offset"])
			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			newerFirst := params["newer-first"] == "true"
			unreadOnly := params["unread-only"] == "true"

			if limit > 50 {
				limit = 50
			}

			if feedId == "__favorite__" {
				if newerFirst {
					articles, err = db.GetUserFavoriteArticlesDesc(user, limit, offset)
				} else {
					articles, err = db.GetUserFavoriteArticles(user, limit, offset)
				}
				if err != nil {
					break
				}
			} else if feedId == "__popular__" {
				earlier := time.Now().AddDate(0, 0, -5)
				if newerFirst {
					articles, err = db.GetScoredUserArticlesDesc(user, earlier, limit, offset)
				} else {
					articles, err = db.GetScoredUserArticles(user, earlier, limit, offset)
				}
				if err != nil {
					break
				}
			} else if feedId == "tag:__all__" {
				if newerFirst {
					if unreadOnly {
						articles, err = db.GetUnreadUserArticlesDesc(user, limit, offset)
					} else {
						articles, err = db.GetUserArticlesDesc(user, limit, offset)
					}
				} else {
					if unreadOnly {
						articles, err = db.GetUnreadUserArticles(user, limit, offset)
					} else {
						articles, err = db.GetUserArticles(user, limit, offset)
					}
				}
				if err != nil {
					break
				}
			} else if strings.HasPrefix(feedId, "tag:") {
				tag := feedId[4:]
				if newerFirst {
					if unreadOnly {
						articles, err = db.GetUnreadUserTagArticlesDesc(user, tag, limit, offset)
					} else {
						articles, err = db.GetUserTagArticlesDesc(user, tag, limit, offset)
					}
				} else {
					if unreadOnly {
						articles, err = db.GetUnreadUserTagArticles(user, tag, limit, offset)
					} else {
						articles, err = db.GetUserTagArticles(user, tag, limit, offset)
					}
				}
				if err != nil {
					break
				}
			} else {
				var f readeef.Feed

				var id int64
				id, err = strconv.ParseInt(feedId, 10, 64)

				if err != nil {
					err = errors.New("Unknown feed id " + feedId)
					break
				}

				f, err = db.GetFeed(id)
				/* TODO: non-fatal error */
				if err != nil {
					break
				}

				f.User = user

				if newerFirst {
					if unreadOnly {
						f, err = db.GetUnreadFeedArticlesDesc(f, limit, offset)
					} else {
						f, err = db.GetFeedArticlesDesc(f, limit, offset)
					}
				} else {
					if unreadOnly {
						f, err = db.GetUnreadFeedArticles(f, limit, offset)
					} else {
						f, err = db.GetFeedArticles(f, limit, offset)
					}
				}
				if err != nil {
					break
				}

				articles = f.Articles
			}

			resp["Articles"] = articles
		}

		switch err {
		case readeef.ErrNoAbsolute:
			resp["Error"] = true
			resp["ErrorType"] = "error-no-absolute"
			err = nil
		case readeef.ErrNoFeed:
			resp["Error"] = true
			resp["ErrorType"] = "error-no-feed"
			err = nil
		}

		var b []byte
		if err == nil {
			b, err = json.Marshal(resp)
		}

		if err == nil {
			w.Write(b)
		} else {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
		}

	}
}

func (con Feed) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
