package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	time "time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_eventSocket(t *testing.T) {
	tests := []struct {
		name             string
		ctx              func() context.Context
		noUser           bool
		feeds            []content.Feed
		feedsErr         error
		data             string
		noToken          bool
		blacklistToken   bool
		expiredClaims    bool
		closedConnection bool
		worker           func(context.Context, eventable.Service, *mock_repo.MockArticle, *mock_repo.MockFeed)
	}{
		{name: "no user", ctx: lazyContext(0), noUser: true},
		{name: "feed error", ctx: lazyContext(0), feedsErr: errors.New("err")},
		{
			name:    "initial data",
			ctx:     lazyContext(time.Millisecond),
			feeds:   []content.Feed{{ID: 1, Link: "http://example.com"}},
			data:    "event: connection-established\n\n",
			noToken: true,
		},
		{
			name:    "no claim",
			ctx:     lazyContext(10 * time.Millisecond),
			feeds:   []content.Feed{{ID: 1, Link: "http://example.com"}},
			data:    "event: connection-established\n\n",
			noToken: true,
			worker: func(ctx context.Context, s eventable.Service, a *mock_repo.MockArticle, f *mock_repo.MockFeed) {
				time.Sleep(time.Millisecond)

				f.EXPECT().Update(gomock.Any()).Return([]content.Article{{ID: 1}, {ID: 2}}, nil)
				s.FeedRepo().Update(&content.Feed{ID: 1})
			},
		},
		{
			name:           "logged out",
			ctx:            lazyContext(10 * time.Millisecond),
			feeds:          []content.Feed{{ID: 1, Link: "http://example.com"}},
			data:           "event: connection-established\n\n",
			blacklistToken: true,
			worker: func(ctx context.Context, s eventable.Service, a *mock_repo.MockArticle, f *mock_repo.MockFeed) {
				time.Sleep(time.Millisecond)

				f.EXPECT().Update(gomock.Any()).Return([]content.Article{{ID: 1}, {ID: 2}}, nil)
				s.FeedRepo().Update(&content.Feed{ID: 1})
			},
		},
		{
			name:          "expired token",
			ctx:           lazyContext(10 * time.Millisecond),
			feeds:         []content.Feed{{ID: 1, Link: "http://example.com"}},
			data:          "event: connection-established\n\n",
			expiredClaims: true,
			worker: func(ctx context.Context, s eventable.Service, a *mock_repo.MockArticle, f *mock_repo.MockFeed) {
				time.Sleep(time.Millisecond)

				f.EXPECT().Update(gomock.Any()).Return([]content.Article{{ID: 1}, {ID: 2}}, nil)
				s.FeedRepo().Update(&content.Feed{ID: 1})
			},
		},
		{
			name:             "closed connection",
			ctx:              lazyContext(10 * time.Millisecond),
			feeds:            []content.Feed{{ID: 1, Link: "http://example.com"}},
			data:             "event: connection-established\n\n",
			closedConnection: true,
		},
		{
			name:  "one event",
			ctx:   lazyContext(10 * time.Millisecond),
			feeds: []content.Feed{{ID: 1, Link: "http://example.com"}},
			data: `event: connection-established

event: feed-update
data: {"articleIDs":[1,2],"feedID":1}

`,
			worker: func(ctx context.Context, s eventable.Service, a *mock_repo.MockArticle, f *mock_repo.MockFeed) {
				time.Sleep(time.Millisecond)

				f.EXPECT().Update(gomock.Any()).Return([]content.Article{{ID: 1}, {ID: 2}}, nil)
				s.FeedRepo().Update(&content.Feed{ID: 1})
			},
		},
		{
			name:             "one event and close",
			ctx:              lazyContext(10 * time.Millisecond),
			feeds:            []content.Feed{{ID: 1, Link: "http://example.com"}},
			closedConnection: true,
			data: `event: connection-established

event: feed-update
data: {"articleIDs":[1,2],"feedID":1}

`,
			worker: func(ctx context.Context, s eventable.Service, a *mock_repo.MockArticle, f *mock_repo.MockFeed) {
				f.EXPECT().Update(gomock.Any()).Return([]content.Article{{ID: 1}, {ID: 2}}, nil)
				s.FeedRepo().Update(&content.Feed{ID: 1})

				time.Sleep(2 * time.Millisecond)

				f.EXPECT().Update(gomock.Any()).Return([]content.Article{{ID: 3}, {ID: 4}}, nil)
				s.FeedRepo().Update(&content.Feed{ID: 1})

			},
		},
		{
			name:  "one event and ping",
			ctx:   lazyContext(10*time.Second + 10*time.Millisecond),
			feeds: []content.Feed{{ID: 1, Link: "http://example.com"}},
			data: `event: connection-established

event: feed-update
data: {"articleIDs":[1,2],"feedID":1}

: ping

`,
			worker: func(ctx context.Context, s eventable.Service, a *mock_repo.MockArticle, f *mock_repo.MockFeed) {
				time.Sleep(2 * time.Second)

				f.EXPECT().Update(gomock.Any()).Return([]content.Article{{ID: 1}, {ID: 2}}, nil)
				s.FeedRepo().Update(&content.Feed{ID: 1})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := mock_repo.NewMockService(ctrl)
			feedRepo := mock_repo.NewMockFeed(ctrl)
			articleRepo := mock_repo.NewMockArticle(ctrl)
			storage := NewMockStorage(ctrl)

			r := httptest.NewRequest("GET", "/", nil)
			r.ParseForm()
			w := NewCloseNotifier()

			service.EXPECT().FeedRepo().Return(feedRepo)
			service.EXPECT().ArticleRepo().Return(articleRepo)

			ctx := tt.ctx()
			ev := eventable.NewService(ctx, service, logger)

			code := http.StatusOK
			var flush bool
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

				feedRepo.EXPECT().ForUser(userMatcher{user}).Return(tt.feeds, tt.feedsErr)

				if tt.feedsErr != nil {
					code = http.StatusInternalServerError
					break
				}

				flush = true

				if tt.worker != nil {
					go tt.worker(ctx, ev, articleRepo, feedRepo)
				}

				if !tt.noToken && tt.worker != nil {
					r = r.WithContext(context.WithValue(r.Context(), auth.TokenKey, "token"))
					storage.EXPECT().Exists("token").MinTimes(1).Return(tt.blacklistToken, nil)

					r = r.WithContext(context.WithValue(r.Context(), auth.ClaimsKey, claims{!tt.expiredClaims}))
				}

				if tt.closedConnection {
					time.AfterFunc(time.Millisecond, func() { w.Notify(true) })
				}
			}

			eventSocket(ctx, ev, storage, logger).ServeHTTP(w, r)

			if code != w.Code {
				t.Errorf("eventSocket() code = %v, want %v", w.Code, code)
				return
			}

			ct := "text/plain; charset=utf-8"
			if code == http.StatusOK {
				ct = "text/event-stream"
			}
			if w.Header().Get("Content-Type") != ct {
				t.Errorf("eventSocket() content-type = %v, want %v", w.Header().Get("Content-Type"), ct)
				return
			}

			if w.Flushed != flush {
				t.Errorf("eventSocket() flush = %v, want %v", w.Flushed, flush)
				return
			}

			if code == http.StatusOK {
				if w.Body.String() != tt.data {
					t.Errorf("eventSocket() data = %v, want %v", w.Body.String(), tt.data)
					return
				}
			}
		})
	}
}

func lazyContext(t time.Duration) func() context.Context {
	return func() context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), t)
		time.AfterFunc(t, cancel)
		return ctx
	}
}

var _ http.CloseNotifier = CloseNotify{}

type CloseNotify struct {
	*httptest.ResponseRecorder
	ch chan bool
}

func NewCloseNotifier() CloseNotify {
	return CloseNotify{httptest.NewRecorder(), make(chan bool)}
}

func (c CloseNotify) CloseNotify() <-chan bool {
	return c.ch
}

func (c CloseNotify) Notify(b bool) {
	c.ch <- b
}

type claims struct{ valid bool }

func (c claims) Valid() error {
	if c.valid {
		return nil
	}

	return errors.New("invalid")
}
