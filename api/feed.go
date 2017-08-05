package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/repo"
)

func feedContext(repo repo.Feed, log readeef.Logger) func(http.Handler) http.Handler {
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
			if err {
				if err == content.IsNoContent(err) {
					http.Error(w, "Not found", http.StatusNotFound)
				} else {
					error(w, log, "Error getting feed: %+v", err)
				}
				return
			}

			ctx := context.WithValue(r.Context(), "feed", feed)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func feedFromRequest(w http.ResponseWriter, r *http.Request) (feed content.UserFeed, stop bool) {
	var ok bool
	if feed, ok = r.Context().Value("feed").(content.UserFeed); ok {
		return feed, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return nil, true
}

func listFeeds(repo repo.Feed, log readeef.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feeds, err := repo.ForUser(user)
		if err != nil {
			error(w, log, "Error getting feeds: %+v", log)
			return
		}

		args{"feeds": feeds}.WriteJSON(w)
	}
}

type addFeedData struct {
	Link  string   `json:"link"`
	Links []string `json:"links"`
}

type addFeedError struct {
	Link    string `json:"link"`
	Title   string `json:"title"`
	Message string `json:"error"`
}

func (e addFeedError) Error() string {
	return e.Message
}

func addFeed(repo repo.Feed, feedManager *readeef.FeedManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		data := addFeedData{}
		if stop = readJSON(w, r.Body, &data); stop {
			return
		}

		links := make([]string, len(data.Links), len(data.Links)+1)
		copy(links, data.Links)
		if data.Link != "" {
			links = append([]string{data.Link}, links...)
		}

		errs := make([]error, 0, len(links))
		for _, link := range links {
			err := addFeedByURL(link, user, repo, feedManager)
			if err != nil {
				errs = append(errs, err)
			}
		}

		args{"errors": errs, "success": len(errs) < len(links)}.WriteJSON(w)
	}
}

func addFeedByURL(
	link string,
	user content.User,
	repo repo.Feed,
	feedManager *readeef.FeedManager,
) error {
	u, err := url.Parse(link)
	if err != nil {
		return addFeedError{Link: link, Message: "Not a url"}
	}

	if !u.IsAbs() {
		return addFeedError{Link: link, Message: "Link is not absolute"}
	}

	if f, err := feedManager.AddFeedByLink(link); err == nil {
		err = repo.Attach(f, user)
		if err != nil {
			return addFeedError{Link: link, Title: f.Data().Title, Message: "Error adding feed to the database: " + err.Error()}
		}

		tags := strings.SplitN(u.Fragment, ",", -1)
		if u.Fragment != "" && len(tags) > 0 {
			t := make([]content.Tag, len(tags))
			for i := range tags {
				t[i] = content.Tag{Value: content.TagValue(tags[i])}
			}

			if err = repo.SetUserTags(f, user, tags); err != nil {
				return addFeedError{Link: link, Title: f.Title, Message: "Error adding feed to the database: " + err.Error()}
			}
		}
	} else {
		return addFeedError{Link: link, Message: "Error adding feed to the database"}
	}

	return nil
}

func deleteFeed(repo repo.Feed, feedManager *readeef.FeedManager, log readeef.Logger) http.HandlerFunc {
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
			error(w, log, "Error deleting feed: %+v", err)
			return
		}

		feedManager.RemoveFeed(feed)

		args{"success": true}.WriteJSON(w)
	}
}

func discoverFeeds(repo repo.Feed, feedManager *readeef.FeedManager, log readeef.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "No query", http.StatusBadRequest)
			return
		}

		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feeds, err := discoverFeedsByQuery(query, user, repo, feedManager)
		if err == nil {
			args{"feeds": feeds}.WriteJSON(w)
		} else {
			http.Error(w, "Error discovering feeds: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func discoverFeedsByQuery(query string, user content.User, repo repo.Feed, feedManager *readeef.FeedManager) ([]content.Feed, error) {
	feeds, err := feedManager.DiscoverFeeds(query)
	if err != nil {
		return nil, err
	}

	feeds, err := repo.ForUser(user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting feeds for user")
	}

	userFeedIdMap := make(map[data.FeedId]bool)
	userFeedLinkMap := make(map[string]bool)
	for i := range feeds {
		feed := feeds[i]
		userFeedIdMap[feed.Id] = true
		userFeedLinkMap[feed.Link] = true

		u, err := url.Parse(feed.Link)
		if err == nil && strings.HasPrefix(u.Host, "www.") {
			u.Host = u.Host[4:]
			userFeedLinkMap[u.String()] = true
		}
	}

	respFeeds := []content.Feed{}
	for i := range feeds {
		feed := feeds[i]
		if !userFeedIdMap[feed.Id] && !userFeedLinkMap[feed.Link] {
			respFeeds = append(respFeeds, feeds[i])
		}
	}

	return respFeeds, nil
}
