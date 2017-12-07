package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_feedContext(t *testing.T) {
	tests := []struct {
		name         string
		hasUser      bool
		feedID       string
		hasFeedIDErr bool
		feed         content.Feed
		feedErr      error
	}{
		{name: "no user"},
		{name: "no feed id", hasUser: true, hasFeedIDErr: true},
		{name: "invalid feed id", hasUser: true, feedID: "foo", hasFeedIDErr: true},
		{name: "feed err", hasUser: true, feedID: "12", feedErr: errors.New("feed err")},
		{name: "no such feed", hasUser: true, feedID: "12", feedErr: content.ErrNoContent},
		{name: "success", hasUser: true, feedID: "12", feed: content.Feed{ID: 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			feedRepo := mock_repo.NewMockFeed(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusNoContent
			switch {
			default:
				var user content.User
				if tt.hasUser {
					user = content.User{Login: "test"}
					r = r.WithContext(context.WithValue(r.Context(), userKey, user))
				} else {
					code = http.StatusBadRequest
					break
				}

				r = addChiParam(r, "feedID", tt.feedID)

				if tt.hasFeedIDErr {
					code = http.StatusBadRequest
					break
				}

				id, err := strconv.ParseInt(tt.feedID, 10, 63)
				if err != nil {
					t.Fatal(err)
				}

				feedRepo.EXPECT().Get(content.FeedID(id), userMatcher{user}).Return(tt.feed, tt.feedErr)

				if content.IsNoContent(tt.feedErr) {
					code = http.StatusNotFound
				} else if tt.feedErr != nil {
					code = http.StatusInternalServerError
				}
			}

			feedContext(feedRepo, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if feed, stop := feedFromRequest(w, r); !stop {
					if !reflect.DeepEqual(feed, tt.feed) {
						t.Errorf("feedContext() feed = %v, want %v", feed, tt.feed)
						return
					}
					w.WriteHeader(http.StatusNoContent)
				}
			})).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("feedContext() code = %v, want %v", code, w.Code)
			}
		})
	}
}

func Test_listFeeds(t *testing.T) {
	tests := []struct {
		name     string
		hasUser  bool
		feeds    []content.Feed
		feedsErr error
	}{
		{name: "no user"},
		{name: "feed list err", hasUser: true, feedsErr: errors.New("err")},
		{name: "feed list", hasUser: true, feeds: []content.Feed{{ID: 1}, {ID: 2, Link: "http://example.com"}}},
	}

	type data struct {
		Feeds []content.Feed `json:"feeds"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			feedRepo := mock_repo.NewMockFeed(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusOK
			switch {
			default:
				var user content.User
				if tt.hasUser {
					user = content.User{Login: "test"}
					r = r.WithContext(context.WithValue(r.Context(), userKey, user))
				} else {
					code = http.StatusBadRequest
					break
				}

				feedRepo.EXPECT().ForUser(userMatcher{user}).Return(tt.feeds, tt.feedsErr)
				if tt.feedsErr != nil {
					code = http.StatusInternalServerError
				}
			}

			listFeeds(feedRepo, logger).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("listFeeds() code = %v, want %v", code, w.Code)
				return
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (code == http.StatusOK) {
				t.Errorf("listFeeds() body = %s", w.Body)
				return
			}

			if !reflect.DeepEqual(got.Feeds, tt.feeds) {
				t.Errorf("listFeeds() got = %v, want %v", got.Feeds, tt.feeds)
				return
			}
		})
	}
}

func Test_addFeed(t *testing.T) {
	tests := []struct {
		name         string
		form         url.Values
		noUser       bool
		addedFeed    []content.Feed
		addedFeedErr []error
		attachErr    []error
	}{
		{name: "no user", noUser: true},
		{name: "no links"},
		{
			name:         "one success link",
			form:         url.Values{"link": []string{"http://example.com"}},
			addedFeed:    []content.Feed{{ID: 1, Link: "http://example.com"}},
			addedFeedErr: []error{nil},
			attachErr:    []error{nil},
		},
		{
			name:         "multi links",
			form:         url.Values{"link": []string{"http://example.com", "http://broken.com"}},
			addedFeed:    []content.Feed{{ID: 1, Link: "http://example.com"}, {}},
			addedFeedErr: []error{nil, errors.New("err")},
			attachErr:    []error{nil},
		},
	}

	type addErr struct {
		Link  string `json:"link"`
		Title string `json:"title"`
		Error string `json:"error"`
	}
	type data struct {
		Errors  []addErr                `json:"errors"`
		Feeds   map[string]content.Feed `json:"feeds"`
		Success bool                    `json:"success"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			feedRepo := mock_repo.NewMockFeed(ctrl)
			feedManager := NewMockfeedManager(ctrl)

			r := httptest.NewRequest("POST", "/", strings.NewReader(tt.form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.ParseForm()
			w := httptest.NewRecorder()

			code := http.StatusOK
			want := data{}
			switch {
			default:
				var user content.User
				if tt.noUser {
					code = http.StatusBadRequest
					break
				} else {
					user = content.User{Login: "test"}
					r = r.WithContext(context.WithValue(r.Context(), userKey, user))
				}

				want.Errors = []addErr{}
				want.Feeds = map[string]content.Feed{}
				for i, link := range tt.form["link"] {
					feedManager.EXPECT().AddFeedByLink(link).Return(tt.addedFeed[i], tt.addedFeedErr[i])

					if tt.addedFeedErr[i] != nil {
						want.Errors = append(want.Errors, addErr{Link: link, Error: "adding feed to the database: " + tt.addedFeedErr[i].Error()})
						continue
					}

					feedRepo.EXPECT().AttachTo(tt.addedFeed[i], userMatcher{user}).Return(tt.attachErr[i])
					if tt.attachErr[i] != nil {
						want.Errors = append(want.Errors, addErr{Link: link, Error: "adding feed to user test: " + tt.attachErr[i].Error()})
						continue
					}

					want.Feeds[link] = tt.addedFeed[i]
				}

				want.Success = len(want.Errors) < len(tt.form["link"])
			}

			addFeed(feedRepo, feedManager).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("addFeed() code = %v, want %v", code, w.Code)
				return
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (code == http.StatusOK) {
				t.Errorf("addFeed() body = %s, error = %+v", w.Body, err)
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("addFeed() got = %v, want = %v", got, want)
			}
		})
	}
}

func Test_deleteFeed(t *testing.T) {
	tests := []struct {
		name      string
		noUser    bool
		noFeed    bool
		detachErr error
	}{
		{name: "no user", noUser: true},
		{name: "no feed", noFeed: true},
		{name: "detach err", detachErr: errors.New("err")},
		{name: "success"},
	}

	type data struct {
		Success bool `json:"success"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			feedRepo := mock_repo.NewMockFeed(ctrl)
			feedManager := NewMockfeedManager(ctrl)

			r := httptest.NewRequest("DELETE", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusOK
			want := data{}
			switch {
			default:
				var user content.User
				var feed content.Feed
				if tt.noUser {
					code = http.StatusBadRequest
					break
				} else {
					user = content.User{Login: "test"}
					r = r.WithContext(context.WithValue(r.Context(), userKey, user))
				}

				if tt.noFeed {
					code = http.StatusBadRequest
					break
				} else {
					feed = content.Feed{ID: 1, Link: "http://example.com"}
					r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))
				}

				feedRepo.EXPECT().DetachFrom(feed, user).Return(tt.detachErr)

				if tt.detachErr != nil {
					code = http.StatusInternalServerError
					break
				}

				feedManager.EXPECT().RemoveFeed(feed)

				want.Success = true
			}

			deleteFeed(feedRepo, feedManager, logger).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("deleteFeed() code = %v, want %v", w.Code, code)
				return
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (code == http.StatusOK) {
				t.Errorf("deleteFeed() body = %s, error = %+v", w.Body, err)
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("deleteFeed() got = %v, want = %v", got, want)
			}
		})
	}
}

func Test_discoverFeeds(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		noUser           bool
		noQuery          bool
		userFeeds        []content.Feed
		userFeedsErr     error
		discoverFeeds    []content.Feed
		discoverFeedsErr error
	}{
		{name: "no user", url: "/?query=test", noUser: true},
		{name: "no query", url: "/?query2=test", noQuery: true},
		{name: "user feeds err", url: "/?query=test", userFeedsErr: errors.New("err")},
		{name: "discover feeds err", url: "/?query=test", userFeeds: []content.Feed{{ID: 1, Link: "http://example.com"}, {ID: 2, Link: "http://www.example2.com"}}, discoverFeedsErr: errors.New("err")},
		{name: "some discovered feeds", url: "/?query=test", userFeeds: []content.Feed{{ID: 1, Link: "http://example.com"}, {ID: 2, Link: "http://www.example2.com"}}, discoverFeeds: []content.Feed{{Link: "http://example3.com"}, {Link: "http://example4.com"}}},
		{name: "no discovered feeds", url: "/?query=test", userFeeds: []content.Feed{{ID: 1, Link: "http://example.com"}, {ID: 2, Link: "http://www.example2.com"}}, discoverFeeds: []content.Feed{}},
	}

	type data struct {
		Feeds []content.Feed `json:"feeds"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			feedRepo := mock_repo.NewMockFeed(ctrl)
			feedManager := NewMockfeedManager(ctrl)

			r := httptest.NewRequest("GET", tt.url, nil)
			r.ParseForm()
			w := httptest.NewRecorder()

			code := http.StatusOK
			want := data{}
			switch {
			default:
				var user content.User
				var query string
				if tt.noUser {
					code = http.StatusBadRequest
					break
				} else {
					user = content.User{Login: "test"}
					r = r.WithContext(context.WithValue(r.Context(), userKey, user))
				}

				if tt.noQuery {
					code = http.StatusBadRequest
					break
				} else {
					query = r.Form.Get("query")
				}

				feedRepo.EXPECT().ForUser(userMatcher{user}).Return(tt.userFeeds, tt.userFeedsErr)

				if tt.userFeedsErr != nil {
					code = http.StatusInternalServerError
					break
				}

				feedManager.EXPECT().DiscoverFeeds(query).Return(tt.discoverFeeds, tt.discoverFeedsErr)

				if tt.discoverFeedsErr != nil {
					code = http.StatusInternalServerError
					break
				}

				want.Feeds = tt.discoverFeeds
			}

			discoverFeeds(feedRepo, feedManager, logger).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("discoverFeed() code = %v, want %v", w.Code, code)
				return
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (code == http.StatusOK) {
				t.Errorf("discoverFeed() body = %s, error = %+v", w.Body, err)
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("discoverFeed() got = %v, want = %v", got, want)
			}
		})
	}
}
