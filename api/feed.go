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
)

func feedContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		id, err := strconv.ParseInt(chi.URLParam(r, "feedId"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		feed := user.FeedById(data.FeedId(id))
		if feed.HasErr() {
			err := feed.Err()
			if err == content.ErrNoContent {
				http.Error(w, "Not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error getting feed: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		ctx := context.WithValue(r.Context(), "feed", feed)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func feedFromRequest(w http.ResponseWriter, r *http.Request) (feed content.UserFeed, stop bool) {
	var ok bool
	if feed, ok = r.Context().Value("feed").(content.UserFeed); ok {
		return feed, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return nil, true
}

func listFeeds(w http.ResponseWriter, r *http.Request) {
	user, stop := userFromRequest(w, r)
	if stop {
		return
	}

	feeds := user.AllTaggedFeeds()
	if user.HasErr() {
		http.Error(w, "Error getting feeds: "+user.Err().Error(), http.StatusInternalServerError)
		return
	}

	args{"feeds": feeds}.WriteJSON(w)
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

func addFeed(feedManager *readeef.FeedManager) http.HandlerFunc {
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
			err := addFeedByURL(link, user, feedManager)
			if err != nil {
				errs = append(errs, err)
			}
		}

		args{"errors": errs, "success": len(errs) < len(links)}.WriteJSON(w)
	}
}

func addFeedByURL(link string, user content.User, feedManager *readeef.FeedManager) error {
	u, err := url.Parse(link)
	if err != nil {
		return addFeedError{Link: link, Message: "Not a url"}
	}

	if !u.IsAbs() {
		return addFeedError{Link: link, Message: "Link is not absolute"}
	}

	if f, err := feedManager.AddFeedByLink(link); err == nil {
		uf := user.AddFeed(f)
		if uf.HasErr() {
			return addFeedError{Link: link, Title: f.Data().Title, Message: "Error adding feed to the database: " + uf.Err().Error()}
		}

		tags := strings.SplitN(u.Fragment, ",", -1)
		if u.Fragment != "" && len(tags) > 0 {
			repo := uf.Repo()
			tf := repo.TaggedFeed(user)
			tf.Data(uf.Data())

			t := make([]content.Tag, len(tags))
			for i := range tags {
				t[i] = repo.Tag(user)
				t[i].Data(data.Tag{Value: data.TagValue(tags[i])})
			}

			tf.Tags(t)
			if tf.UpdateTags(); tf.HasErr() {
				return addFeedError{Link: link, Title: f.Data().Title, Message: "Error adding feed to the database: " + tf.Err().Error()}
			}
		}
	} else {
		return addFeedError{Link: link, Message: "Error adding feed to the database"}
	}

	return nil
}

func deleteFeed(feedManager *readeef.FeedManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if feed, stop := feedFromRequest(w, r); stop {
			return
		} else {
			if feed.Detach(); feed.HasErr() {
				http.Error(w, "Error deleting feed: "+feed.Err().Error(), http.StatusInternalServerError)

				return
			}

			feedManager.RemoveFeed(feed)

			args{"success": true}.WriteJSON(w)
		}
	}
}

func discoverFeeds(feedManager *readeef.FeedManager) http.HandlerFunc {
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

		feeds, err := discoverFeedsByQuery(query, user, feedManager)
		if err == nil {
			args{"feeds": feeds}.WriteJSON(w)
		} else {
			http.Error(w, "Error discovering feeds: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func discoverFeedsByQuery(query string, user content.User, feedManager *readeef.FeedManager) ([]content.Feed, error) {
	feeds, err := feedManager.DiscoverFeeds(query)
	if err != nil {
		return nil, err
	}

	uf := user.AllFeeds()
	if user.HasErr() {
		return nil, errors.Wrap(user.Err(), "getting user feeds")
	}

	userFeedIdMap := make(map[data.FeedId]bool)
	userFeedLinkMap := make(map[string]bool)
	for i := range uf {
		in := uf[i].Data()
		userFeedIdMap[in.Id] = true
		userFeedLinkMap[in.Link] = true

		u, err := url.Parse(in.Link)
		if err == nil && strings.HasPrefix(u.Host, "www.") {
			u.Host = u.Host[4:]
			userFeedLinkMap[u.String()] = true
		}
	}

	respFeeds := []content.Feed{}
	for i := range feeds {
		in := feeds[i].Data()
		if !userFeedIdMap[in.Id] && !userFeedLinkMap[in.Link] {
			respFeeds = append(respFeeds, feeds[i])
		}
	}

	return respFeeds, nil
}
