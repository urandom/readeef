package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"readeef"
	"readeef/parser"
	"strconv"
	"strings"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type Feed struct {
	webfw.BaseController
	fu FeedUpdater
}

func NewFeed(fu FeedUpdater) Feed {
	return Feed{
		webfw.NewBaseController("/v:version/feed/*action", webfw.MethodAll, ""),
	}
}

type feed struct {
	Title       string
	Description string
	Link        string
	Image       parser.Image
	Articles    []readeef.Article
}

func (con Feed) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		db := readeef.GetDB(c)
		user := readeef.GetUser(c, r)

		actionParam := webfw.GetParams(c, r)
		parts := strings.Split(actionParam["action"], "/")
		action := parts[0]

		switch action {
		case "all":
			var feeds []readeef.Feed
			feeds, err = db.GetUserFeeds(user)

			if err != nil {
				break
			}

			type response struct {
				Feeds []feed
			}

			resp := response{}
			for _, f := range feeds {
				resp.Feeds = append(resp.Feeds, feed{
					Title: f.Title, Description: f.Description, Link: f.Link, Image: f.Image,
				})
			}

			var b []byte
			b, err = json.Marshal(resp)
			if err != nil {
				break
			}

			w.Write(b)
		case "add":
			r.ParseForm()

			link := r.FormValue("url")
			var u *url.URL

			/* TODO: non-fatal error */
			if u, err = url.Parse(f.Link); err != nil {
				break
				/* TODO: non-fatal error */
			} else if !u.IsAbs() {
				err = errors.New("Feed has no link")
				break
			}

			err = fu.AddFeedByLink(link)
			if err != nil {
				break
			}

			type response struct {
				success bool
			}
			resp := response{true}

			var b []byte
			b, err = json.Marshal(resp)
			if err != nil {
				break
			}

			w.Write(b)
		case "remove":
			id, err := strconv.ParseInt(parts[1], 10, 64)

			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			fu.RemoveFeed(id)

			type response struct {
				success bool
			}
			resp := response{true}

			var b []byte
			b, err = json.Marshal(resp)
			if err != nil {
				break
			}

			w.Write(b)
		default:
			id, err := strconv.ParseInt(action, 10, 64)

			if err != nil {
				err = errors.New("Unknown action " + action)
				break
			}

			f, err := db.GetFeed(id)
			/* TODO: non-fatal error */
			if err != nil {
				break
			}

			limit := 50
			offset := 0

			if len(parts) == 3 {
				limit, err = strconv.Atoi(parts[1])
				/* TODO: non-fatal error */
				if err != nil {
					break
				}

				offset, err = strconv.Atoi(parts[2])
				/* TODO: non-fatal error */
				if err != nil {
					break
				}
			}
			if limit > 50 {
				limit = 50
			}

			f, err = db.GetFeedArticles(f, limit, offset)
			if err != nil {
				break
			}

			type response struct {
				Feed feed
			}

			resp := response{Feed: feed{
				Title: f.Title, Description: f.Description, Link: f.Link, Image: f.Image,
				Articles: f.Articles,
			}}

			var b []byte
			b, err = json.Marshal(resp)
			if err != nil {
				break
			}

			w.Write(b)
		}

		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (con Feed) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
