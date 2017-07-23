package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func getFeedTags(w http.ResponseWriter, r *http.Request) {
	user, stop := userFromRequest(w, r)
	if stop {
		return
	}

	feed, stop := feedFromRequest(w, r)
	if stop {
		return
	}

	tf := user.Repo().TaggedFeed(user)
	tf.Data(feed.Data())

	tags := tf.Tags()
	if tf.HasErr() {
		http.Error(w, "Error getting feed tags: "+tf.Err().Error(), http.StatusInternalServerError)
		return
	}

	t := make([]string, len(tags))
	for _, tag := range tags {
		t = append(t, tag.String())
	}

	args{"tags": t}.WriteJSON(w)
}

func setFeedTags(w http.ResponseWriter, r *http.Request) {
	user, stop := userFromRequest(w, r)
	if stop {
		return
	}

	feed, stop := feedFromRequest(w, r)
	if stop {
		return
	}

	var tagValues []data.TagValue
	if stop := readJSON(w, r.Body, &tagValues); stop {
		return
	}

	tf := user.Repo().TaggedFeed(user)
	tf.Data(feed.Data())

	filtered := make([]data.TagValue, 0, len(tagValues))
	for _, v := range tagValues {
		v = data.TagValue(strings.TrimSpace(string(v)))
		if v != "" {
			filtered = append(filtered, v)
		}
	}

	tags := make([]content.Tag, len(filtered))
	for i := range filtered {
		tags[i] = user.Repo().Tag(user)
		tags[i].Data(data.Tag{Value: filtered[i]})
	}

	tf.Tags(tags)
	if tf.UpdateTags(); tf.HasErr() {
		http.Error(w, "Error updating feed tags: "+tf.Err().Error(), http.StatusInternalServerError)
		return
	}

	args{"success": true}.WriteJSON(w)
}

func tagContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		id, err := strconv.ParseInt(chi.URLParam(r, "tagId"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tag := user.TagById(data.TagId(id))
		if tag.HasErr() {
			err := tag.Err()
			if err == content.ErrNoContent {
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

func tagFromRequest(w http.ResponseWriter, r *http.Request) (tag content.Tag, stop bool) {
	var ok bool
	if tag, ok = r.Context().Value("tag").(content.Tag); ok {
		return tag, false
	}

	http.Error(w, "Bad Request", http.StatusBadRequest)
	return nil, true
}
