package content

import (
	"errors"
	"net/url"
	"time"
)

type Subscription struct {
	Link                string
	FeedID              FeedID    `db:"feed_id"`
	LeaseDuration       int64     `db:"lease_duration"`
	VerificationTime    time.Time `db:"verification_time"`
	SubscriptionFailure bool      `db:"subscription_failure"`
}

func (s Subscription) Validate() error {
	if s.Link == "" {
		return NewValidationError(errors.New("No subscription link"))
	}

	if u, err := url.Parse(s.Link); err != nil || !u.IsAbs() {
		return NewValidationError(errors.New("Invalid subscription link"))
	}

	if s.FeedID == 0 {
		return NewValidationError(errors.New("Invalid feed id"))
	}

	return nil
}
