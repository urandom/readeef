package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

var feedKey = contextKey("feed")

func feedContext(repo repo.Feed, log log.Log) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, stop := userFromRequest(w, r)
			if stop {
				return
			}

			id, err := strconv.ParseInt(chi.URLParam(r, "feedID"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			feed, err := repo.Get(content.FeedID(id), user)
			if err != nil {
				if content.IsNoContent(err) {
					http.Error(w, "Not found", http.StatusNotFound)
				} else {
					fatal(w, log, "Error getting feed: %+v", err)
				}
				return
			}

			ctx := context.WithValue(r.Context(), feedKey, feed)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func feedFromRequest(w http.ResponseWriter, r *http.Request) (feed content.Feed, stop bool) {
	var ok bool
	if feed, ok = r.Context().Value(feedKey).(content.Feed); ok {
		return feed, false
	}

	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	return content.Feed{}, true
}

func listFeeds(repo repo.Feed, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feeds, err := repo.ForUser(user)
		if err != nil {
			fatal(w, log, "Error getting feeds: %+v", err)
			return
		}

		args{"feeds": feeds}.WriteJSON(w)
	}
}

type addFeedError struct {
	Link    string `json:"link"`
	Title   string `json:"title"`
	Message string `json:"error"`
}

func (e addFeedError) Error() string {
	return e.Message
}

type feedManager interface {
	AddFeedByLink(link string) (content.Feed, error)
	RemoveFeed(feed content.Feed)
	DiscoverFeeds(link string) ([]content.Feed, error)
}

func addFeed(repo repo.Feed, feedManager feedManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		links := r.Form["link"]

		errs := make([]error, 0, len(links))
		feeds := map[string]content.Feed{}
		for _, link := range links {
			feed, err := addFeedByURL(link, user, repo, feedManager)
			if err == nil {
				feeds[link] = feed
			} else {
				errs = append(errs, err)
			}
		}

		args{"errors": errs, "feeds": feeds, "success": len(errs) < len(links)}.WriteJSON(w)
	}
}

func addFeedByURL(
	link string,
	user content.User,
	repo repo.Feed,
	feedManager feedManager,
) (content.Feed, error) {
	u, err := url.Parse(link)
	if err != nil {
		return content.Feed{}, addFeedError{Link: link, Message: "Not a url"}
	}

	if !u.IsAbs() {
		return content.Feed{}, addFeedError{Link: link, Message: "Link is not absolute"}
	}

	if f, err := feedManager.AddFeedByLink(link); err == nil {
		err = repo.AttachTo(f, user)
		if err != nil {
			return content.Feed{}, addFeedError{Link: link, Title: f.Title, Message: fmt.Sprintf("adding feed to user %s: %s", user, err.Error())}
		}

		tags := strings.SplitN(u.Fragment, ",", -1)
		if u.Fragment != "" && len(tags) > 0 {
			t := make([]*content.Tag, len(tags))
			for i := range tags {
				t[i] = &content.Tag{Value: content.TagValue(tags[i])}
			}

			if err = repo.SetUserTags(f, user, t); err != nil {
				return content.Feed{}, addFeedError{Link: link, Title: f.Title, Message: "adding feed tags to the database: " + err.Error()}
			}
		}

		return f, nil
	} else {
		return content.Feed{}, addFeedError{Link: link, Message: "adding feed to the database: " + err.Error()}
	}

}

func deleteFeed(repo repo.Feed, feedManager feedManager, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feed, stop := feedFromRequest(w, r)
		if stop {
			return
		}

		if err := repo.DetachFrom(feed, user); err != nil {
			fatal(w, log, "Error deleting feed: %+v", err)
			return
		}

		feedManager.RemoveFeed(feed)

		args{"success": true}.WriteJSON(w)
	}
}

func discoverFeeds(repo repo.Feed, discoverer feedManager, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.Form.Get("query")
		if query == "" {
			http.Error(w, "No query", http.StatusBadRequest)
			return
		}

		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feeds, err := discoverFeedsByQuery(query, user, repo, discoverer)
		if err == nil {
			args{"feeds": feeds}.WriteJSON(w)
		} else {
			http.Error(w, "Error discovering feeds: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func discoverFeedsByQuery(query string, user content.User, repo repo.Feed, discoverer feedManager) ([]content.Feed, error) {
	userFeeds, err := repo.ForUser(user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting feeds for user")
	}

	userFeedIdMap := make(map[content.FeedID]bool)
	userFeedLinkMap := make(map[string]bool)
	for _, feed := range userFeeds {
		userFeedIdMap[feed.ID] = true
		userFeedLinkMap[feed.Link] = true

		u, err := url.Parse(feed.Link)
		if err == nil && strings.HasPrefix(u.Host, "www.") {
			u.Host = u.Host[4:]
			userFeedLinkMap[u.String()] = true
		}
	}

	feeds, err := discoverer.DiscoverFeeds(query)
	if err != nil {
		return nil, err
	}

	respFeeds := []content.Feed{}
	for i := range feeds {
		feed := feeds[i]
		if !userFeedIdMap[feed.ID] && !userFeedLinkMap[feed.Link] {
			respFeeds = append(respFeeds, feeds[i])
		}
	}

	return respFeeds, nil
}
