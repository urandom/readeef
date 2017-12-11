package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_getArticle(t *testing.T) {
	tests := []struct {
		name      string
		noArticle bool
		code      int
	}{
		{name: "no article", noArticle: true, code: http.StatusBadRequest},
		{name: "article", code: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			w := NewCloseNotifier()

			if !tt.noArticle {
				r = r.WithContext(context.WithValue(r.Context(), articleKey, content.Article{ID: 1}))
			}

			getArticle(w, r)

			if tt.code != w.Code {
				t.Errorf("getArticle() code = %v, want %v", w.Code, tt.code)
				return
			}
		})
	}
}

func Test_getArticles(t *testing.T) {
	tests := []struct {
		name        string
		noUser      bool
		url         string
		badQuery    bool
		repoType    articleRepoType
		subType     articleRepoType
		noFeed      bool
		noTag       bool
		feedIDsErr  error
		queryOpts   []interface{}
		articles    []content.Article
		articlesErr error
		code        int
	}{
		{name: "no user", url: "/", noUser: true, code: 400},
		{name: "invalid limit opt", url: "/?limit=no", badQuery: true, code: 400},
		{name: "invalid offset opt", url: "/?offset=no", badQuery: true, code: 400},
		{name: "invalid before id opt", url: "/?beforeID=no", badQuery: true, code: 400},
		{name: "invalid after id opt", url: "/?afterID=no", badQuery: true, code: 400},
		{name: "invalid before time opt", url: "/?beforeTime=no", badQuery: true, code: 400},
		{name: "invalid after time opt", url: "/?afterTime=no", badQuery: true, code: 400},
		{name: "invalid ids", url: "/?id=4&id=no", badQuery: true, code: 400},
		{name: "invalid repo type", url: "/?id=4&id=1&limit=10&beforeID=3", code: 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := mock_repo.NewMockService(ctrl)
			articleRepo := mock_repo.NewMockArticle(ctrl)
			tagRepo := mock_repo.NewMockTag(ctrl)
			proc := NewMockArticleProcessor(ctrl)

			r := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			service.EXPECT().ArticleRepo().Return(articleRepo)
			service.EXPECT().TagRepo().Return(tagRepo)

			switch {
			default:
				var user content.User
				var feed content.Feed
				var tag content.Tag
				if tt.noUser {
					break
				} else {
					user = content.User{Login: "test"}
					r = r.WithContext(context.WithValue(r.Context(), userKey, user))
				}

				if tt.badQuery {
					break
				}

				if tt.repoType == feedRepoType || tt.subType == feedRepoType {
					if !tt.noFeed {
						feed = content.Feed{ID: 1}
						r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))
					}
				}

				if tt.repoType == tagRepoType || tt.subType == tagRepoType {
					if !tt.noTag {
						tag = content.Tag{ID: 1}
						r = r.WithContext(context.WithValue(r.Context(), tagKey, tag))

						ids := []content.FeedID{1, 2, 3, 4}
						tagRepo.EXPECT().FeedIDs(tag, userMatcher{user}).Return(ids, tt.feedIDsErr)

						if tt.feedIDsErr != nil {
							break
						}
					}
				}

				if tt.repoType == 0 || tt.repoType == popularRepoType && tt.subType == 0 {
					break
				}

				articleRepo.EXPECT().ForUser(userMatcher{user}, tt.queryOpts...).Return(tt.articles, tt.articlesErr)

				if tt.articlesErr != nil {
					break
				}

				proc.EXPECT().ProcessArticles(tt.articles).Return(tt.articles)
			}

			getArticles(service, tt.repoType, tt.subType, []processor.Article{proc}, 50, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("getArticles() code = %v, want %v", w.Code, tt.code)
			}
		})
	}
}
