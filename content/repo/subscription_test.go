package repo_test

import (
	"sort"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
)

var (
	subscription1    content.Subscription
	subscription2    content.Subscription
	subscriptionSync sync.Once
)

func Test_subscriptionRepo_Get(t *testing.T) {
	skipTest(t)
	setupSubscription()

	tests := []struct {
		name      string
		feed      content.Feed
		want      content.Subscription
		wantErr   bool
		noContent bool
	}{
		{"valid1", feed1, subscription1, false, false},
		{"valid2", feed2, subscription2, false, false},
		{"no content", content.Feed{ID: 4000, Link: "http://sugr.org"}, content.Subscription{}, true, true},
		{"invalid", content.Feed{}, content.Subscription{}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.SubscriptionRepo()

			got, err := r.Get(tt.feed)
			if (err != nil) != tt.wantErr {
				t.Errorf("subscriptionRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.noContent && errors.Cause(err) != content.ErrNoContent {
				t.Errorf("subscriptionRepo.Get() error = %v, wanted no content", err)
				return
			}

			if !subscriptionsEqual(got, tt.want) {
				t.Errorf("subscriptionRepo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_subscriptionRepo_All(t *testing.T) {
	skipTest(t)
	setupSubscription()

	tests := []struct {
		name    string
		want    []content.Subscription
		wantErr bool
	}{
		{"all", []content.Subscription{subscription1, subscription2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.SubscriptionRepo()
			got, err := r.All()
			if (err != nil) != tt.wantErr {
				t.Errorf("subscriptionRepo.All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Slice(got, func(i, j int) bool {
				return got[i].FeedID < got[j].FeedID
			})

			if len(got) != len(tt.want) {
				t.Errorf("subscriptionRepo.All() = %v, want %v", got, tt.want)
			}

			for i := range got {
				if !subscriptionsEqual(got[i], tt.want[i]) {
					t.Errorf("subscriptionRepo.All() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_subscriptionRepo_Update(t *testing.T) {
	skipTest(t)
	setupSubscription()

	tests := []struct {
		name         string
		feed         content.Feed
		subscription content.Subscription
		wantErr      bool
	}{
		{"valid", feed1, content.Subscription{FeedID: feed1.ID, Link: "http://sugr.org"}, false},
		{"invalid 1", content.Feed{}, content.Subscription{Link: "http://sugr.org"}, true},
		{"invalid 2", content.Feed{}, content.Subscription{FeedID: feed1.ID}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := service.SubscriptionRepo()
			if err := r.Update(tt.subscription); (err != nil) != tt.wantErr {
				t.Errorf("subscriptionRepo.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got, err := r.Get(tt.feed)
			if err != nil {
				t.Errorf("subscriptionRepo.Update() post fetch error = %v", err)
				return
			}

			if !subscriptionsEqual(got, tt.subscription) {
				t.Errorf("subscriptionRepo.Update() = %#v, want %#v", got, tt.subscription)
			}
		})
	}
}

func setupSubscription() {
	setupFeed()

	subscriptionSync.Do(func() {
		subscription1 = content.Subscription{Link: "http://sugr.org", FeedID: feed1.ID}
		subscription2 = content.Subscription{Link: "http://sugr.org", FeedID: feed2.ID}

		if err := service.SubscriptionRepo().Update(subscription1); err != nil {
			panic(err)
		}

		if err := service.SubscriptionRepo().Update(subscription2); err != nil {
			panic(err)
		}
	})
}

func subscriptionsEqual(a, b content.Subscription) bool {
	return a.FeedID == b.FeedID &&
		a.Link == b.Link &&
		a.LeaseDuration == b.LeaseDuration &&
		a.SubscriptionFailure == b.SubscriptionFailure &&
		a.VerificationTime.Equal(b.VerificationTime)
}
