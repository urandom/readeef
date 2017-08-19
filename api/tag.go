package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

func listTags(repo repo.Tag, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		tags, err := repo.ForUser(user)
		if err != nil {
			fatal(w, log, "Error getting tags: %+v", err)
			return
		}

		args{"tags": tags}.WriteJSON(w)
	}
}

type tagsFeedIDs struct {
	Tag content.Tag      `json:"tag"`
	IDs []content.FeedID `json:"ids"`
}

func getTagsFeedIDs(repo repo.Tag, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		tags, err := repo.ForUser(user)
		if err != nil {
			fatal(w, log, "Error getting user tags: %+v", err)
			return
		}

		resp := []tagsFeedIDs{}

		for _, tag := range tags {
			ids, err := repo.FeedIDs(tag, user)
			if err != nil {
				fatal(w, log, "Error getting tag feed ids: %+v", err)
				return
			}

			resp = append(resp, tagsFeedIDs{tag, ids})
		}

		args{"tagFeeds": resp}.WriteJSON(w)
	}
}

func getFeedTags(repo repo.Tag, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feed, stop := feedFromRequest(w, r)
		if stop {
			return
		}

		tags, err := repo.ForFeed(feed, user)
		if err != nil {
			fatal(w, log, "Error getting feed tags: %+v", err)
			return
		}

		t := make([]string, len(tags))
		for _, tag := range tags {
			t = append(t, tag.String())
		}

		args{"tags": t}.WriteJSON(w)
	}
}

func setFeedTags(repo repo.Feed, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feed, stop := feedFromRequest(w, r)
		if stop {
			return
		}

		var tagValues []content.TagValue
		if stop := readJSON(w, r.Body, &tagValues); stop {
			return
		}

		tags := make([]content.Tag, 0, len(tagValues))
		for i := range tagValues {
			if tagValues[i] != "" {
				tags = append(tags, content.Tag{Value: tagValues[i]})
			}
		}

		if err := repo.SetUserTags(feed, user, tags); err != nil {
			fatal(w, log, "Error updating feed tags: %+v", err)
			return
		}

		args{"success": true}.WriteJSON(w)
	}
}

func tagContext(repo repo.Tag) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, stop := userFromRequest(w, r)
			if stop {
				return
			}

			id, err := strconv.ParseInt(chi.URLParam(r, "tagID"), 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			tag, err := repo.Get(content.TagID(id), user)
			if err != nil {
				if content.IsNoContent(err) {
					http.Error(w, "Not found", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			ctx := context.WithValue(r.Context(), "tag", tag)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func tagFromRequest(w http.ResponseWriter, r *http.Request) (tag content.Tag, stop bool) {
	var ok bool
	if tag, ok = r.Context().Value("tag").(content.Tag); ok {
		return tag, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return content.Tag{}, true
}
