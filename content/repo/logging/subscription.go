package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type subscriptionRepo struct {
	repo.Subscription

	log log.Log
}

func (r subscriptionRepo) Get(feed content.Feed) (content.Subscription, error) {
	start := time.Now()

	subscription, err := r.Subscription.Get(feed)

	r.log.Infof("repo.Subscription.Get took %s", time.Now().Sub(start))

	return subscription, err
}

func (r subscriptionRepo) All() ([]content.Subscription, error) {
	start := time.Now()

	subscriptions, err := r.Subscription.All()

	r.log.Infof("repo.Subscription.All took %s", time.Now().Sub(start))

	return subscriptions, err
}

func (r subscriptionRepo) Update(subscription content.Subscription) error {
	start := time.Now()

	err := r.Subscription.Update(subscription)

	r.log.Infof("repo.Subscription.Update took %s", time.Now().Sub(start))

	return err
}
