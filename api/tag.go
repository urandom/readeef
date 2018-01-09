package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

var tagKey = contextKey("tag")

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

func getTagFeedIDs(repo repo.Tag, log log.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		tag, stop := tagFromRequest(w, r)
		if stop {
			return
		}

		ids, err := repo.FeedIDs(tag, user)
		if err != nil {
			fatal(w, log, "Error getting tag feed ids: %+v", err)
			return
		}

		args{"feedIDs": ids}.WriteJSON(w)
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
		for i := range tags {
			t[i] = tags[i].String()
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

		tagValues := make([]content.TagValue, len(r.Form["tag"]))
		tagIDs := make([]content.TagID, len(r.Form["id"]))
		for i, t := range r.Form["tag"] {
			tagValues[i] = content.TagValue(t)
		}

		for i, v := range r.Form["id"] {
			id, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid id: %s", v), http.StatusBadRequest)
				break
			}

			tagIDs[i] = content.TagID(id)
		}

		tags := make([]*content.Tag, 0, len(tagValues)+len(tagIDs))
		for i := range tagValues {
			if tagValues[i] != "" {
				tags = append(tags, &content.Tag{Value: tagValues[i]})
			}
		}

		for i := range tagIDs {
			if tagIDs[i] > 0 {
				tags = append(tags, &content.Tag{ID: tagIDs[i]})
			}
		}

		if err := repo.SetUserTags(feed, user, tags); err != nil {
			fatal(w, log, "Error updating feed tags: %+v", err)
			return
		}

		args{"success": true}.WriteJSON(w)
	}
}

func tagContext(repo repo.Tag, log log.Log) func(http.Handler) http.Handler {
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
					http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				} else {
					fatal(w, log, "Error getting tag: %+v", err)
				}
				return
			}

			ctx := context.WithValue(r.Context(), tagKey, tag)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func tagFromRequest(w http.ResponseWriter, r *http.Request) (tag content.Tag, stop bool) {
	var ok bool
	if tag, ok = r.Context().Value(tagKey).(content.Tag); ok {
		return tag, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return content.Tag{}, true
}
