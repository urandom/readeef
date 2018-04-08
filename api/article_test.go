package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
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
		opts        content.QueryOptions
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
		{name: "articles err", url: "/", repoType: favoriteRepoType, articlesErr: errors.New("err"), code: 500, opts: content.QueryOptions{Limit: 50, FavoriteOnly: true, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular user", url: "/", repoType: popularRepoType, subType: userRepoType, articles: []content.Article{{ID: 1}}, code: 200, opts: content.QueryOptions{Limit: 50, IncludeScores: true, HighScoredFirst: true, BeforeDate: time.Now(), AfterDate: time.Now().AddDate(0, 0, -5), SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular tag", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: tagRepoType, articles: []content.Article{{ID: 1}}, code: 200, opts: content.QueryOptions{Limit: 25, Offset: 0, IncludeScores: true, HighScoredFirst: true, BeforeDate: time.Now(), AfterDate: time.Now().AddDate(0, 0, -5), FeedIDs: []content.FeedID{1, 2, 3, 4}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular no tag", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: tagRepoType, noTag: true, code: 400},
		{name: "popular tag err", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: tagRepoType, articles: nil, code: 500, feedIDsErr: errors.New("err")},
		{name: "popular feed", url: "/?limit=25", repoType: popularRepoType, subType: feedRepoType, articles: []content.Article{{ID: 1}}, code: 200, opts: content.QueryOptions{Limit: 25, IncludeScores: true, HighScoredFirst: true, BeforeDate: time.Now(), AfterDate: time.Now().AddDate(0, 0, -5), FeedIDs: []content.FeedID{1}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular no feed", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: feedRepoType, noFeed: true, code: 400},
		{name: "popular unknown", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: 0, code: 400},
		{name: "tag", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadOnly: true, FeedIDs: []content.FeedID{1, 2, 3, 4}, SortField: content.SortByDate, SortOrder: content.AscendingOrder}, articles: []content.Article{{ID: 1}, {ID: 2, Link: "http://example.com"}}},
		{name: "tag err", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 500, feedIDsErr: errors.New("err")},
		{name: "no tag", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 400, noTag: true},
		{name: "feed", url: "/?limit=25&unreadFirst", repoType: feedRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadFirst: true, FeedIDs: []content.FeedID{1}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}, articles: []content.Article{{ID: 1}, {ID: 2, Link: "http://example.com"}}},
		{name: "no feed", url: "/?limit=25&unreadFirst", repoType: feedRepoType, code: 400, noFeed: true},
		{name: "user", url: "/?limit=25&beforeTime=100000&afterTime=500", repoType: userRepoType, code: 200, opts: content.QueryOptions{Limit: 25, AfterDate: time.Unix(500, 0), BeforeDate: time.Unix(100000, 0), SortField: content.SortByDate, SortOrder: content.DescendingOrder}, articles: []content.Article{{ID: 1}, {ID: 2, Link: "http://example.com"}}},
	}
	type data struct {
		Articles []content.Article `json:"articles"`
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
			r.ParseForm()
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
				}
				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.badQuery {
					break
				}

				if tt.repoType == feedRepoType || tt.subType == feedRepoType {
					if tt.noFeed {
						break
					}
					feed = content.Feed{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))
				}

				if tt.repoType == tagRepoType || tt.subType == tagRepoType {
					if tt.noTag {
						break
					}
					tag = content.Tag{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), tagKey, tag))

					ids := []content.FeedID{1, 2, 3, 4}
					tagRepo.EXPECT().FeedIDs(tag, userMatcher{user}).Return(ids, tt.feedIDsErr)

					if tt.feedIDsErr != nil {
						break
					}
				}

				if tt.repoType == 0 || tt.repoType == popularRepoType && tt.subType == 0 {
					break
				}

				articleRepo.EXPECT().ForUser(userMatcher{user}, gomock.Any()).DoAndReturn(func(user content.User, opts ...content.QueryOpt) ([]content.Article, error) {
					o := content.QueryOptions{}
					o.Apply(opts)

					if o.BeforeDate.Sub(tt.opts.BeforeDate) > time.Second || o.AfterDate.Sub(tt.opts.AfterDate) > time.Second {
						t.Errorf("getArticles() date options = %#v, want %#v", o, tt.opts)
						return tt.articles, tt.articlesErr
					}

					tt.opts.BeforeDate, tt.opts.AfterDate, o.BeforeDate, o.AfterDate = time.Time{}, time.Time{}, time.Time{}, time.Time{}

					if !reflect.DeepEqual(o, tt.opts) {
						t.Errorf("getArticles() options = %#v, want %#v", o, tt.opts)
					}

					return tt.articles, tt.articlesErr
				})

				if tt.articlesErr != nil {
					break
				}

				proc.EXPECT().ProcessArticles(tt.articles).Return(tt.articles)
			}

			getArticles(service, tt.repoType, tt.subType, []processor.Article{proc}, 50, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("getArticles() code = %v, want %v", w.Code, tt.code)
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (tt.code == http.StatusOK) {
				t.Errorf("getArticles() body = %s", w.Body)
				return
			}

			if !reflect.DeepEqual(got.Articles, tt.articles) {
				t.Errorf("getArticles() got = %v, want %v", got.Articles, tt.articles)
				return
			}
		})
	}
}

func Test_articleSearch(t *testing.T) {
	tests := []struct {
		name        string
		noUser      bool
		url         string
		noQuery     bool
		badQuery    bool
		repoType    articleRepoType
		noFeed      bool
		noTag       bool
		feedIDsErr  error
		opts        content.QueryOptions
		articles    []content.Article
		articlesErr error
		code        int
	}{
		{name: "no query", url: "/", noQuery: true, code: 400},
		{name: "no user", url: "/?query=test", noUser: true, code: 400},
		{name: "invalid limit opt", url: "/?query=test&limit=no", badQuery: true, code: 400},
		{name: "invalid offset opt", url: "/?query=test&offset=no", badQuery: true, code: 400},
		{name: "invalid before id opt", url: "/?query=test&beforeID=no", badQuery: true, code: 400},
		{name: "invalid after id opt", url: "/?query=test&afterID=no", badQuery: true, code: 400},
		{name: "invalid before time opt", url: "/?query=test&beforeTime=no", badQuery: true, code: 400},
		{name: "invalid after time opt", url: "/?query=test&afterTime=no", badQuery: true, code: 400},
		{name: "invalid ids", url: "/?query=test&id=4&id=no", badQuery: true, code: 400},
		{name: "invalid repo type", url: "/?query=test&id=4&id=1&limit=10&beforeID=3", code: 400},
		{name: "articles err", url: "/?query=test", repoType: userRepoType, articlesErr: errors.New("err"), code: 500, opts: content.QueryOptions{Limit: 50, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "tag", url: "/?query=test&limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadOnly: true, FeedIDs: []content.FeedID{1, 2, 3, 4}, SortField: content.SortByDate, SortOrder: content.AscendingOrder}, articles: []content.Article{{ID: 1}, {ID: 2, Link: "http://example.com"}}},
		{name: "tag err", url: "/?query=test&limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 500, feedIDsErr: errors.New("err")},
		{name: "no tag", url: "/?query=test&limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 400, noTag: true},
		{name: "feed", url: "/?query=test&limit=25&unreadFirst", repoType: feedRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadFirst: true, FeedIDs: []content.FeedID{1}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}, articles: []content.Article{{ID: 1}, {ID: 2, Link: "http://example.com"}}},
		{name: "no feed", url: "/?query=test&limit=25&unreadFirst", repoType: feedRepoType, code: 400, noFeed: true},
		{name: "user", url: "/?query=test&limit=25&beforeTime=100000&afterTime=500", repoType: userRepoType, code: 200, opts: content.QueryOptions{Limit: 25, AfterDate: time.Unix(500, 0), BeforeDate: time.Unix(100000, 0), SortField: content.SortByDate, SortOrder: content.DescendingOrder}, articles: []content.Article{{ID: 1}, {ID: 2, Link: "http://example.com"}}},
	}
	type data struct {
		Articles []content.Article `json:"articles"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := mock_repo.NewMockService(ctrl)
			tagRepo := mock_repo.NewMockTag(ctrl)
			searchProvider := NewMocksearcher(ctrl)
			proc := NewMockArticleProcessor(ctrl)

			r := httptest.NewRequest("GET", tt.url, nil)
			r.ParseForm()
			w := httptest.NewRecorder()

			service.EXPECT().TagRepo().Return(tagRepo)

			switch {
			default:
				if tt.noQuery {
					break
				}

				if tt.noUser {
					break
				}

				var user content.User
				var feed content.Feed
				var tag content.Tag

				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.badQuery {
					break
				}

				if tt.repoType == feedRepoType {
					if tt.noFeed {
						break
					}
					feed = content.Feed{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))
				}

				if tt.repoType == tagRepoType {
					if tt.noTag {
						break
					}
					tag = content.Tag{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), tagKey, tag))

					ids := []content.FeedID{1, 2, 3, 4}
					tagRepo.EXPECT().FeedIDs(tag, userMatcher{user}).Return(ids, tt.feedIDsErr)

					if tt.feedIDsErr != nil {
						break
					}
				}

				if tt.repoType == 0 {
					break
				}

				searchProvider.EXPECT().Search(r.Form.Get("query"), userMatcher{user}, gomock.Any()).DoAndReturn(func(query string, user content.User, opts ...content.QueryOpt) ([]content.Article, error) {
					o := content.QueryOptions{}
					o.Apply(opts)

					if o.BeforeDate.Sub(tt.opts.BeforeDate) > time.Second || o.AfterDate.Sub(tt.opts.AfterDate) > time.Second {
						t.Errorf("getArticles() date options = %#v, want %#v", o, tt.opts)
						return tt.articles, tt.articlesErr
					}

					tt.opts.BeforeDate, tt.opts.AfterDate, o.BeforeDate, o.AfterDate = time.Time{}, time.Time{}, time.Time{}, time.Time{}

					if !reflect.DeepEqual(o, tt.opts) {
						t.Errorf("getArticles() options = %#v, want %#v", o, tt.opts)
					}

					return tt.articles, tt.articlesErr
				})

				if tt.articlesErr != nil {
					break
				}

				proc.EXPECT().ProcessArticles(tt.articles).Return(tt.articles)
			}

			articleSearch(service, searchProvider, tt.repoType, []processor.Article{proc}, 50, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("articleSearch() code = %v, want %v", w.Code, tt.code)
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (tt.code == http.StatusOK) {
				t.Errorf("articleSearch() body = %s", w.Body)
				return
			}

			if !reflect.DeepEqual(got.Articles, tt.articles) {
				t.Errorf("articleSearch() got = %v, want %v", got.Articles, tt.articles)
				return
			}
		})
	}
}

func Test_getIDs(t *testing.T) {
	tests := []struct {
		name       string
		noUser     bool
		url        string
		badQuery   bool
		repoType   articleRepoType
		subType    articleRepoType
		noFeed     bool
		noTag      bool
		feedIDsErr error
		opts       content.QueryOptions
		ids        []content.ArticleID
		idsErr     error
		code       int
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
		{name: "articles err", url: "/", repoType: favoriteRepoType, idsErr: errors.New("err"), code: 500, opts: content.QueryOptions{Limit: 2500, FavoriteOnly: true, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular user", url: "/", repoType: popularRepoType, subType: userRepoType, ids: []content.ArticleID{1}, code: 200, opts: content.QueryOptions{Limit: 2500, IncludeScores: true, HighScoredFirst: true, BeforeDate: time.Now(), AfterDate: time.Now().AddDate(0, 0, -5), SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular tag", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: tagRepoType, ids: []content.ArticleID{1}, code: 200, opts: content.QueryOptions{Limit: 25, Offset: 0, IncludeScores: true, HighScoredFirst: true, BeforeDate: time.Now(), AfterDate: time.Now().AddDate(0, 0, -5), FeedIDs: []content.FeedID{1, 2, 3, 4}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular no tag", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: tagRepoType, noTag: true, code: 400},
		{name: "popular tag err", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: tagRepoType, code: 500, feedIDsErr: errors.New("err")},
		{name: "popular feed", url: "/?limit=25", repoType: popularRepoType, subType: feedRepoType, ids: []content.ArticleID{1}, code: 200, opts: content.QueryOptions{Limit: 25, IncludeScores: true, HighScoredFirst: true, BeforeDate: time.Now(), AfterDate: time.Now().AddDate(0, 0, -5), FeedIDs: []content.FeedID{1}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "popular no feed", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: feedRepoType, noFeed: true, code: 400},
		{name: "popular unknown", url: "/?limit=25&offset=10", repoType: popularRepoType, subType: 0, code: 400},
		{name: "tag", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadOnly: true, FeedIDs: []content.FeedID{1, 2, 3, 4}, SortField: content.SortByDate, SortOrder: content.AscendingOrder}, ids: []content.ArticleID{1, 2}},
		{name: "tag err", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 500, feedIDsErr: errors.New("err")},
		{name: "no tag", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 400, noTag: true},
		{name: "feed", url: "/?limit=25&unreadFirst", repoType: feedRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadFirst: true, FeedIDs: []content.FeedID{1}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}, ids: []content.ArticleID{1, 2}},
		{name: "no feed", url: "/?limit=25&unreadFirst", repoType: feedRepoType, code: 400, noFeed: true},
		{name: "user", url: "/?limit=25&beforeTime=100000&afterTime=500", repoType: userRepoType, code: 200, opts: content.QueryOptions{Limit: 25, AfterDate: time.Unix(500, 0), BeforeDate: time.Unix(100000, 0), SortField: content.SortByDate, SortOrder: content.DescendingOrder}, ids: []content.ArticleID{1, 2}},
	}
	type data struct {
		IDs []content.ArticleID `json:"ids"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := mock_repo.NewMockService(ctrl)
			articleRepo := mock_repo.NewMockArticle(ctrl)
			tagRepo := mock_repo.NewMockTag(ctrl)

			r := httptest.NewRequest("GET", tt.url, nil)
			r.ParseForm()
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
				}
				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.badQuery {
					break
				}

				if tt.repoType == feedRepoType || tt.subType == feedRepoType {
					if tt.noFeed {
						break
					}
					feed = content.Feed{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))
				}

				if tt.repoType == tagRepoType || tt.subType == tagRepoType {
					if tt.noTag {
						break
					}
					tag = content.Tag{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), tagKey, tag))

					ids := []content.FeedID{1, 2, 3, 4}
					tagRepo.EXPECT().FeedIDs(tag, userMatcher{user}).Return(ids, tt.feedIDsErr)

					if tt.feedIDsErr != nil {
						break
					}
				}

				if tt.repoType == 0 || tt.repoType == popularRepoType && tt.subType == 0 {
					break
				}

				articleRepo.EXPECT().IDs(userMatcher{user}, gomock.Any()).DoAndReturn(func(user content.User, opts ...content.QueryOpt) ([]content.ArticleID, error) {
					o := content.QueryOptions{}
					o.Apply(opts)

					if o.BeforeDate.Sub(tt.opts.BeforeDate) > time.Second || o.AfterDate.Sub(tt.opts.AfterDate) > time.Second {
						t.Errorf("getIDs() date options = %#v, want %#v", o, tt.opts)
						return tt.ids, tt.idsErr
					}

					tt.opts.BeforeDate, tt.opts.AfterDate, o.BeforeDate, o.AfterDate = time.Time{}, time.Time{}, time.Time{}, time.Time{}

					if !reflect.DeepEqual(o, tt.opts) {
						t.Errorf("getIDs() options = %#v, want %#v", o, tt.opts)
					}

					return tt.ids, tt.idsErr
				})

				if tt.idsErr != nil {
					break
				}
			}

			getIDs(service, tt.repoType, tt.subType, 50, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("getIDs() code = %v, want %v", w.Code, tt.code)
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (tt.code == http.StatusOK) {
				t.Errorf("getIDs() body = %s", w.Body)
				return
			}

			if !reflect.DeepEqual(got.IDs, tt.ids) {
				t.Errorf("getIDs() got = %v, want %v", got.IDs, tt.ids)
				return
			}
		})
	}
}

func Test_formatArticle(t *testing.T) {
	tests := []struct {
		name       string
		noUser     bool
		noArticle  bool
		extract    content.Extract
		extractErr error
		code       int
	}{
		{name: "no user", noUser: true, code: 400},
		{name: "no article", noArticle: true, code: 400},
		{name: "extract error", extractErr: errors.New("err"), code: 500},
		{name: "success", extract: content.Extract{Title: "New package updates", Content: "There have been a total of 40 package updates to date. Some of them are crucial. Others not so."}, code: 200},
	}
	type data struct {
		KeyPoints []string `json:"keyPoints"`
		Content   string   `json:"content"`
		TopImage  string   `json:"topImage"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			extractRepo := mock_repo.NewMockExtract(ctrl)
			generator := NewMockGenerator(ctrl)
			proc := NewMockArticleProcessor(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			switch {
			default:
				var user content.User
				var article content.Article

				if tt.noUser {
					break
				}

				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.noArticle {
					break
				}

				article = content.Article{ID: 4, Link: "http://example.com"}
				r = r.WithContext(context.WithValue(r.Context(), articleKey, article))

				extractRepo.EXPECT().Get(article).Return(tt.extract, tt.extractErr)

				if tt.extractErr != nil {
					break
				}

				modified := []content.Article{{Description: tt.extract.Content}}
				proc.EXPECT().ProcessArticles(modified).Return(modified)
			}

			formatArticle(extractRepo, generator, []processor.Article{proc}, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("formatArticle() code = %v, want %v", w.Code, tt.code)
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (tt.code == http.StatusOK) {
				t.Errorf("formatArticle() body = %s", w.Body)
				return
			}

			var keypoints []string
			if tt.extract.Content != "" {
				keypoints = []string{"There have been a total of 40 package updates to date.", "Some of them are crucial.", "Others not so."}
			}
			want := data{
				KeyPoints: keypoints,
				Content:   tt.extract.Content,
				TopImage:  tt.extract.TopImage,
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("formatArticle() got = %v, want %v", got, want)
				return
			}
		})
	}
}

func Test_articleStateChange(t *testing.T) {
	tests := []struct {
		name      string
		state     articleState
		value     bool
		current   bool
		noUser    bool
		noArticle bool
		stateErr  error
		code      int
	}{
		{name: "no user", noUser: true, code: 400},
		{name: "no article", noArticle: true, code: 400},
		{name: "no change read true", state: read, current: true, value: true, code: 200},
		{name: "no change read false", state: read, current: false, value: false, code: 200},
		{name: "no change favorite true", state: favorite, current: true, value: true, code: 200},
		{name: "no change favorite false", state: favorite, current: false, value: false, code: 200},
		{name: "read err", state: read, value: true, stateErr: errors.New("err"), code: 500},
		{name: "favorite err", state: favorite, value: true, stateErr: errors.New("err"), code: 500},
		{name: "change read true", state: read, value: true, code: 200},
		{name: "change read false", state: read, current: true, code: 200},
		{name: "change favorite true", state: favorite, value: true, code: 200},
		{name: "change favorite false", state: favorite, current: true, code: 200},
	}

	type data struct {
		Success  bool `json:"success"`
		Read     bool `json:"read"`
		Favorite bool `json:"favorite"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			articleRepo := mock_repo.NewMockArticle(ctrl)

			method := "POST"
			if !tt.value {
				method = "DELETE"
			}
			r := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()

			switch {
			default:
				var user content.User
				var article content.Article

				if tt.noUser {
					break
				}

				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.noArticle {
					break
				}

				article = content.Article{ID: 4, Link: "http://example.com"}
				if tt.current {
					if tt.state == read {
						article.Read = tt.current
					} else {
						article.Favorite = tt.current
					}
				}
				r = r.WithContext(context.WithValue(r.Context(), articleKey, article))

				if tt.current == tt.value {
					break
				}

				if tt.state == read {
					articleRepo.EXPECT().Read(tt.value, userMatcher{user}, gomock.Any()).Return(tt.stateErr)
				} else {
					articleRepo.EXPECT().Favor(tt.value, userMatcher{user}, gomock.Any()).Return(tt.stateErr)
				}

				if tt.stateErr != nil {
					break
				}
			}

			articleStateChange(articleRepo, tt.state, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("articleStateChange() code = %v, want %v", w.Code, tt.code)
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (tt.code == http.StatusOK) {
				t.Errorf("articleStateChange() body = %s", w.Body)
				return
			}
			var want data
			if tt.code == 200 {
				want = data{Success: true}
				if tt.state == read {
					want.Read = tt.value
				} else {
					want.Favorite = tt.value
				}
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("articleStateChange() got = %v, want %v", got, want)
				return
			}
		})
	}
}

func Test_articlesStateChange(t *testing.T) {
	tests := []struct {
		name       string
		value      bool
		url        string
		repoType   articleRepoType
		state      articleState
		noUser     bool
		badQuery   bool
		noFeed     bool
		noTag      bool
		feedIDsErr error
		opts       content.QueryOptions
		stateErr   error
		code       int
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
		{name: "state err", url: "/", repoType: favoriteRepoType, stateErr: errors.New("err"), code: 500, opts: content.QueryOptions{FavoriteOnly: true, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "tag", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadOnly: true, FeedIDs: []content.FeedID{1, 2, 3, 4}, SortField: content.SortByDate, SortOrder: content.AscendingOrder}},
		{name: "tag err", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 500, feedIDsErr: errors.New("err")},
		{name: "no tag", url: "/?limit=25&unreadOnly&olderFirst", repoType: tagRepoType, code: 400, noTag: true},
		{name: "feed", url: "/?limit=25&unreadFirst", repoType: feedRepoType, code: 200, opts: content.QueryOptions{Limit: 25, UnreadFirst: true, FeedIDs: []content.FeedID{1}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "no feed", url: "/?limit=25&unreadFirst", repoType: feedRepoType, code: 400, noFeed: true},
		{name: "user", url: "/?limit=25&beforeTime=100000&afterTime=500", repoType: userRepoType, code: 200, opts: content.QueryOptions{Limit: 25, AfterDate: time.Unix(500, 0), BeforeDate: time.Unix(100000, 0), SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
		{name: "user favorite", url: "/?id=5&id=13&id=14", repoType: userRepoType, state: favorite, code: 200, opts: content.QueryOptions{IDs: []content.ArticleID{5, 13, 14}, SortField: content.SortByDate, SortOrder: content.DescendingOrder}},
	}
	type data struct {
		Success bool `json:"success"`
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := mock_repo.NewMockService(ctrl)
			articleRepo := mock_repo.NewMockArticle(ctrl)
			tagRepo := mock_repo.NewMockTag(ctrl)

			service.EXPECT().ArticleRepo().Return(articleRepo)
			service.EXPECT().TagRepo().Return(tagRepo)

			method := "POST"
			if !tt.value {
				method = "DELETE"
			}
			r := httptest.NewRequest(method, tt.url, nil)
			r.ParseForm()
			w := httptest.NewRecorder()

			switch {
			default:
				var user content.User
				var feed content.Feed
				var tag content.Tag

				if tt.noUser {
					break
				}

				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				if tt.badQuery {
					break
				}

				if tt.repoType == feedRepoType {
					if tt.noFeed {
						break
					}
					feed = content.Feed{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), feedKey, feed))
				}

				if tt.repoType == tagRepoType {
					if tt.noTag {
						break
					}
					tag = content.Tag{ID: 1}
					r = r.WithContext(context.WithValue(r.Context(), tagKey, tag))

					ids := []content.FeedID{1, 2, 3, 4}
					tagRepo.EXPECT().FeedIDs(tag, userMatcher{user}).Return(ids, tt.feedIDsErr)

					if tt.feedIDsErr != nil {
						break
					}
				}

				if tt.repoType == 0 {
					break
				}

				do := func(value bool, user content.User, opts ...content.QueryOpt) error {
					o := content.QueryOptions{}
					o.Apply(opts)

					if o.BeforeDate.Sub(tt.opts.BeforeDate) > time.Second || o.AfterDate.Sub(tt.opts.AfterDate) > time.Second {
						t.Errorf("articlesReadStateChange() date options = %#v, want %#v", o, tt.opts)
						return tt.stateErr
					}

					tt.opts.BeforeDate, tt.opts.AfterDate, o.BeforeDate, o.AfterDate = time.Time{}, time.Time{}, time.Time{}, time.Time{}

					if !reflect.DeepEqual(o, tt.opts) {
						t.Errorf("articlesReadStateChange() options = %#v, want %#v", o, tt.opts)
					}

					return tt.stateErr
				}

				if tt.state == read {
					articleRepo.EXPECT().Read(tt.value, userMatcher{user}, gomock.Any()).DoAndReturn(do)
				} else if tt.state == favorite {
					articleRepo.EXPECT().Favor(tt.value, userMatcher{user}, gomock.Any()).DoAndReturn(do)
				}
			}

			articlesStateChange(service, tt.repoType, tt.state, logger).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("articleReadStateChange() code = %v, want %v", w.Code, tt.code)
			}

			var got data
			if err := json.Unmarshal(w.Body.Bytes(), &got); (err != nil) && (tt.code == http.StatusOK) {
				t.Errorf("articleReadStateChange() body = %s", w.Body)
				return
			}
			want := data{Success: tt.code == 200}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("articleReadStateChange() got = %v, want %v", got, want)
				return
			}
		})
	}
}

func Test_articleContext(t *testing.T) {
	tests := []struct {
		name        string
		isGet       bool
		noUser      bool
		sArticleID  string
		invalidID   bool
		articleID   content.ArticleID
		article     content.Article
		articlesErr error
		code        int
	}{
		{name: "no user", noUser: true, code: 400},
		{name: "no article id", invalidID: true, code: 400},
		{name: "invalid id", sArticleID: "non-numeric", invalidID: true, code: 400},
		{name: "article err", sArticleID: "4", articleID: 4, articlesErr: errors.New("err"), code: 500},
		{name: "no such article", sArticleID: "4", articleID: 4, code: 404},
		{name: "success get", sArticleID: "4", articleID: 4, isGet: true, article: content.Article{ID: 4}, code: 204},
		{name: "success post", sArticleID: "4", articleID: 4, article: content.Article{ID: 4}, code: 204},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			articleRepo := mock_repo.NewMockArticle(ctrl)
			proc := NewMockArticleProcessor(ctrl)

			method := "GET"
			if !tt.isGet {
				method = "POST"
			}
			r := httptest.NewRequest(method, "/", nil)
			r.ParseForm()
			w := httptest.NewRecorder()

			switch {
			default:
				var user content.User
				if tt.noUser {
					break
				}
				user = content.User{Login: "test"}
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))

				r = addChiParam(r, "articleID", tt.sArticleID)
				if tt.invalidID {
					break
				}

				articleRepo.EXPECT().ForUser(userMatcher{user}, gomock.Any()).DoAndReturn(func(user content.User, opts ...content.QueryOpt) ([]content.Article, error) {
					o := content.QueryOptions{}
					o.Apply(opts)

					want := content.QueryOptions{IDs: []content.ArticleID{tt.articleID}}

					if !reflect.DeepEqual(o, want) {
						t.Errorf("getArticles() options = %#v, want %#v", o, want)
					}

					if tt.articlesErr == nil && tt.article.ID != 0 {
						return []content.Article{tt.article}, nil
					}

					return nil, tt.articlesErr
				})

				if tt.articlesErr != nil {
					break
				}

				if tt.isGet {
					proc.EXPECT().ProcessArticles([]content.Article{tt.article}).Return([]content.Article{tt.article})
				}
			}

			articleContext(articleRepo, []processor.Article{proc}, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if article, stop := articleFromRequest(w, r); !stop {
					if !reflect.DeepEqual(article, tt.article) {
						t.Errorf("articleContext() article = %v, want %v", article, tt.article)
						return
					}
				}

				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(w, r)

			if tt.code != w.Code {
				t.Errorf("articleContext() code = %v, want %v", w.Code, tt.code)
			}
		})
	}
}
