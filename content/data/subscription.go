package data

import "time"

type Subscription struct {
	Link                string
	FeedId              FeedId    `db:"feed_id"`
	LeaseDuration       int64     `db:"lease_duration"`
	VerificationTime    time.Time `db:"verification_time"`
	SubscriptionFailure bool      `db:"subscription_failure"`
}
