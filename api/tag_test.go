package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_listTags(t *testing.T) {
	tests := []struct {
		name    string
		hasUser bool
		listErr bool
	}{
		{"no user", false, false},
		{"success list", true, false},
		{"list error", true, true},
	}

	type data struct {
		Tags []content.Tag `json:"tags"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tagRepo := mock_repo.NewMockTag(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			tags := []content.Tag{}
			if tt.hasUser {
				u := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, u))

				if tt.listErr {
					code = http.StatusInternalServerError
				} else {
					code = http.StatusOK
				}

				var err error
				if tt.listErr {
					err = errors.New("list err")
				} else {
					tags = append(tags, content.Tag{ID: 1}, content.Tag{ID: 2})
				}

				tagRepo.EXPECT().ForUser(userMatcher{u}).DoAndReturn(func(u content.User) ([]content.Tag, error) {
					return tags, err
				})
			}

			listTags(tagRepo, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("listTags() code = %v, want %v", w.Code, code)
				return
			}

			var got data
			err := json.Unmarshal(w.Body.Bytes(), &got)
			if (err != nil) && (w.Code == http.StatusOK) {
				t.Errorf("listTags() error = %v, code = %v", err, w.Code)
				return
			}

			var want data
			if tt.hasUser && !tt.listErr {
				want.Tags = tags
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("listTags() got = %v, want = %v", got, want)
			}
		})
	}
}

func Test_getTagsFeedIDs(t *testing.T) {
	tests := []struct {
		name       string
		hasUser    bool
		listTags   []content.Tag
		listErr    error
		feedIDs    [][]content.FeedID
		feedIDsErr error
	}{
		{"no user", false, nil, nil, nil, nil},
		{"list tags err", true, nil, errors.New("list tags err"), nil, nil},
		{"feed ids err", true, []content.Tag{{ID: 1}}, nil, [][]content.FeedID{nil}, errors.New("feed ids err")},
		{"tag list", true, []content.Tag{{ID: 1}, {ID: 2, Value: "val"}}, nil, [][]content.FeedID{{4, 12}, {5, 8, 12, 23}}, nil},
	}

	type data struct {
		TagFeeds []tagsFeedIDs `json:"tagFeeds"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tagRepo := mock_repo.NewMockTag(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			if tt.hasUser {
				u := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, u))

				tagRepo.EXPECT().ForUser(userMatcher{u}).DoAndReturn(func(u content.User) ([]content.Tag, error) {
					return tt.listTags, tt.listErr
				})

				for i, tag := range tt.listTags {
					i := i
					tagRepo.EXPECT().FeedIDs(tag, userMatcher{u}).DoAndReturn(func(tag content.Tag, u content.User) ([]content.FeedID, error) {
						return tt.feedIDs[i], tt.feedIDsErr
					})
				}

				if tt.listErr == nil && tt.feedIDsErr == nil {
					code = http.StatusOK
				} else {
					code = http.StatusInternalServerError
				}
			}

			getTagsFeedIDs(tagRepo, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("getTagsFeedIDs() code = %v, want %v", w.Code, code)
				return
			}

			got := data{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (w.Code == http.StatusOK) {
				t.Errorf("getTagsFeedIDs() body = '%s', error = %v", w.Body, err)
				return
			}

			want := data{}
			if code == http.StatusOK {
				want.TagFeeds = make([]tagsFeedIDs, len(tt.feedIDs))
				for i := range tt.feedIDs {
					want.TagFeeds[i] = tagsFeedIDs{Tag: tt.listTags[i], IDs: tt.feedIDs[i]}
				}
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("getTagsFeedIDs() got = %v, want = %v", got, want)
			}
		})
	}
}

func Test_getTagFeedIDs(t *testing.T) {
	tests := []struct {
		name    string
		hasUser bool
		hasTag  bool
		ids     []content.FeedID
		idsErr  error
	}{
		{"no user", false, false, nil, nil},
		{"no tag", true, false, nil, nil},
		{"ids", true, true, []content.FeedID{4, 12, 18}, nil},
		{"list err", true, true, nil, errors.New("list err")},
	}
	type data struct {
		FeedIDs []content.FeedID `json:"feedIDs"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tagRepo := mock_repo.NewMockTag(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			if tt.hasUser {
				u := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, u))

				if tt.hasTag {
					tag := content.Tag{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), tagKey, tag))

					if tt.idsErr == nil {
						code = http.StatusOK
					} else {
						code = http.StatusInternalServerError
					}

					tagRepo.EXPECT().FeedIDs(tag, u).Return(tt.ids, tt.idsErr)
				}
			}

			getTagFeedIDs(tagRepo, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("getTagFeedIDs() code = %v, want %v", w.Code, code)
				return
			}

			got := data{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (w.Code == http.StatusOK) {
				t.Errorf("getTagFeedIDs() body = '%s', error = %v", w.Body, err)
				return
			}

			want := data{}
			if code == http.StatusOK {
				want.FeedIDs = tt.ids
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("getTagFeedIDs() got = %v, want = %v", got, want)
			}
		})
	}
}

func Test_getFeedTags(t *testing.T) {
	tests := []struct {
		name    string
		hasUser bool
		hasFeed bool
		tags    []content.Tag
		tagsErr error
	}{
		{"no user", false, false, nil, nil},
		{"no feed", true, false, nil, nil},
		{"tags", true, true, []content.Tag{{ID: 1}, {ID: 2, Value: "Foo"}, {ID: 3}}, nil},
		{"tags err", true, true, nil, errors.New("tags err")},
	}

	type data struct {
		Tags []string `json:"tags"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tagRepo := mock_repo.NewMockTag(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			if tt.hasUser {
				u := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, u))

				if tt.hasFeed {
					feed := content.Feed{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))

					if tt.tagsErr == nil {
						code = http.StatusOK
					} else {
						code = http.StatusInternalServerError
					}

					tagRepo.EXPECT().ForFeed(feed, u).Return(tt.tags, tt.tagsErr)
				}
			}

			getFeedTags(tagRepo, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("getFeedTags() code = %v, want %v", w.Code, code)
				return
			}

			got := data{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (w.Code == http.StatusOK) {
				t.Errorf("getFeedTags() body = '%s', error = %v", w.Body, err)
				return
			}

			want := data{}
			if code == http.StatusOK {
				want.Tags = make([]string, len(tt.tags))
				for i := range tt.tags {
					want.Tags[i] = tt.tags[i].String()
				}
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("getFeedTags() got = %v, want = %v", got, want)
			}
		})
	}
}

func Test_setFeedTags(t *testing.T) {
	tests := []struct {
		name    string
		hasUser bool
		hasFeed bool
		form    string
		tags    []*content.Tag
		setErr  error
	}{
		{"no user", false, false, "", nil, nil},
		{"no feed", true, false, "", nil, nil},
		{"tags", true, true, "tag=foo&tag=bar", []*content.Tag{{Value: "foo"}, {Value: "bar"}}, nil},
		{"set err", true, true, "tag=foo", []*content.Tag{{Value: "foo"}}, errors.New("set err")},
	}

	type data struct {
		Success bool `json:"success"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			feedRepo := mock_repo.NewMockFeed(ctrl)

			r := httptest.NewRequest("PUT", "/", strings.NewReader(tt.form))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.ParseForm()
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			if tt.hasUser {
				user := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.hasFeed {
					feed := content.Feed{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))

					if tt.setErr == nil {
						code = http.StatusOK
					} else {
						code = http.StatusInternalServerError
					}

					feedRepo.EXPECT().SetUserTags(feed, userMatcher{user}, gomock.Any()).DoAndReturn(func(f content.Feed, u content.User, tags []*content.Tag) error {
						if !reflect.DeepEqual(tt.tags, tags) {
							t.Errorf("setFeedTags() tags = %v, want %v", tags, tt.tags)
						}

						return tt.setErr
					})
				}
			}

			setFeedTags(feedRepo, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("setFeedTags() code = %v, want %v", w.Code, code)
				return
			}

			got := data{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (w.Code == http.StatusOK) {
				t.Errorf("setFeedTags() body = '%s', error = %v", w.Body, err)
				return
			}

			want := data{Success: code == http.StatusOK}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("setFeedTags() got = %v, want = %v", got, want)
			}
		})
	}
}

func Test_tagContext(t *testing.T) {
	tests := []struct {
		name    string
		hasUser bool
		hasID   bool
		getErr  error
	}{
		{"no user", false, false, nil},
		{"no id", true, false, nil},
		{"get err", true, true, errors.New("get err")},
		{"tag value", true, true, nil},
		{"unknown tag", true, true, content.ErrNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tagRepo := mock_repo.NewMockTag(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			code := http.StatusBadRequest
			if tt.hasUser {
				user := content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.hasID {
					r = addChiParam(r, "tagID", "1")
					tag := content.Tag{}
					if tt.getErr == nil {
						code = http.StatusNoContent
						tag.ID = 1
					} else if content.IsNoContent(tt.getErr) {
						code = http.StatusNotFound
					} else {
						code = http.StatusInternalServerError
					}

					tagRepo.EXPECT().Get(content.TagID(1), userMatcher{user}).Return(tag, tt.getErr)
				} else {
					r = addChiParam(r, "tagID", "foo")
				}
			}

			tagContext(tagRepo, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tag, ok := r.Context().Value(tagKey).(content.Tag); ok {
					if tag.ID != content.TagID(1) {
						t.Errorf("tagContext() tag id value = %v, want = %v", tag.ID, 1)
						return
					}
				} else {
					t.Errorf("tagContext() tag ctx value = %v", r.Context().Value(tagKey))
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("tagContext() code = %v, want %v", w.Code, code)
				return
			}
		})
	}
}
