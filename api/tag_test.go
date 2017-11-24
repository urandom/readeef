package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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

			var exp data
			if tt.hasUser && !tt.listErr {
				exp.Tags = tags
			}

			if !reflect.DeepEqual(got, exp) {
				t.Errorf("listTags() got = %v, want = %v", got, exp)
			}
		})
	}
}
