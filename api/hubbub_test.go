package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/mock_repo"
)

func Test_hubbubRegistration(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		form          string
		hasURLErr     bool
		feedID        content.FeedID
		feed          content.Feed
		feedErr       error
		sub           content.Subscription
		subErr        error
		updateSub     content.Subscription
		updateSubErr  error
		hasFeedXML    bool
		hasFeedXMLErr bool
		updateFeedErr error
		response      []byte
	}{
		{name: "invalid url", url: "/whatever", hasURLErr: true},
		{name: "get feed err", url: "/feed/12", feedID: 12, feedErr: errors.New("get feed err")},
		{name: "get sub err", url: "/feed/12", feedID: 12, feed: content.Feed{ID: 12}, subErr: errors.New("get sub err")},
		{name: "subscription", url: "/feed/12", form: "hub.mode=subscribe&hub.challenge=test", feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12, SubscriptionFailure: true}, updateSub: content.Subscription{FeedID: 12, VerificationTime: time.Now()}, response: []byte("test")},
		{name: "subscription with lease", url: "/feed/12", form: "hub.mode=subscribe&hub.lease_seconds=300&hub.challenge=test", feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12, SubscriptionFailure: true}, updateSub: content.Subscription{FeedID: 12, VerificationTime: time.Now(), LeaseDuration: int64(300 * time.Second)}, response: []byte("test")},
		{name: "unsubscribe", url: "/feed/12", form: "hub.mode=unsubscribe&hub.challenge=test", feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12}, updateSub: content.Subscription{FeedID: 12, SubscriptionFailure: true}, response: []byte("test")},
		{name: "denied", url: "/feed/12", form: "hub.mode=denied", feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12}, updateSub: content.Subscription{FeedID: 12}},
		{name: "update error", url: "/feed/12", form: "hub.mode=denied", feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12}, updateSub: content.Subscription{FeedID: 12}, updateSubErr: errors.New("err")},
		{name: "feed update", url: "/feed/12", form: singleAtomXML, feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12}, hasFeedXML: true},
		{name: "unknown feed format", url: "/feed/12", form: "not-a-feed-xml-format", feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12}, hasFeedXML: true, hasFeedXMLErr: true},
		{name: "feed update err", url: "/feed/12", form: singleAtomXML, feedID: 12, feed: content.Feed{ID: 12}, sub: content.Subscription{FeedID: 12}, hasFeedXML: true, updateFeedErr: errors.New("update feed err")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			service := mock_repo.NewMockService(ctrl)
			feedRepo := mock_repo.NewMockFeed(ctrl)
			subRepo := mock_repo.NewMockSubscription(ctrl)

			service.EXPECT().FeedRepo().Return(feedRepo)
			service.EXPECT().SubscriptionRepo().Return(subRepo)

			r := httptest.NewRequest("POST", tt.url, strings.NewReader(tt.form))
			if !tt.hasFeedXML {
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			r.ParseForm()
			w := httptest.NewRecorder()

			code := http.StatusOK

			switch {
			default:
				if tt.hasURLErr {
					code = http.StatusBadRequest
					break
				}

				feedRepo.EXPECT().Get(tt.feedID, userMatcher{content.User{}}).Return(tt.feed, tt.feedErr)
				if tt.feedErr != nil {
					code = http.StatusInternalServerError
					break
				}

				subRepo.EXPECT().Get(tt.feed).Return(tt.sub, tt.subErr)

				if tt.subErr != nil {
					code = http.StatusInternalServerError
					break
				}

				if tt.hasFeedXML {
					if !tt.hasFeedXMLErr {
						feedRepo.EXPECT().Update(gomock.Any()).Return(nil, tt.updateFeedErr)
					}
					break
				} else {
					subRepo.EXPECT().Update(subscriptionMatcher{tt.updateSub}).Return(tt.updateSubErr)
				}

			}

			hubbubRegistration(service, logger).ServeHTTP(w, r)

			if w.Code != code {
				t.Errorf("hubbubRegistration() code = %v, want %v", w.Code, code)
				return
			}

			if code == http.StatusOK {
				if !reflect.DeepEqual(tt.response, w.Body.Bytes()) {
					t.Errorf("hubbubRegistration() response = %s, want %s", w.Body, string(tt.response))
					return
				}
			}
		})
	}
}

type subscriptionMatcher struct{ subscription content.Subscription }

func (u subscriptionMatcher) Matches(x interface{}) bool {
	if subscription, ok := x.(content.Subscription); ok {
		return u.subscription.FeedID == subscription.FeedID &&
			u.subscription.Link == subscription.Link &&
			u.subscription.LeaseDuration == subscription.LeaseDuration &&
			u.subscription.SubscriptionFailure == subscription.SubscriptionFailure &&
			u.subscription.VerificationTime.Sub(subscription.VerificationTime) < time.Second
	}
	return false
}

func (u subscriptionMatcher) String() string {
	return "Matches by certain subscription fields"
}

const (
	singleAtomXML = `
<feed xmlns="http://www.w3.org/2005/Atom" updated="2003-12-13T18:30:02Z">
	<title>Example Feed</title>
	<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>
	<link href="http://example.org/"></link>
	<author>
		<name>John Doe</name>
		<uri></uri>
		<email></email>
	</author>
	<entry>
		<title>Atom-Powered Robots Run Amok</title>
		<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
		<link href="http://example.org/2003/12/13/atom03"></link>
		<updated>2003-12-13T18:30:02Z</updated>
		<author><name></name><uri></uri><email></email></author>
		<summary>Some text.</summary>
	</entry>
</feed>
`
)
