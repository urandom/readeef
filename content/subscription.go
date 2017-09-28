package content

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

type Subscription struct {
	FeedID              FeedID    `db:"feed_id"`
	Link                string    `db:"link"`
	LeaseDuration       int64     `db:"lease_duration"`
	VerificationTime    time.Time `db:"verification_time"`
	SubscriptionFailure bool      `db:"subscription_failure"`
}

func (s Subscription) Validate() error {
	if s.FeedID == 0 {
		return NewValidationError(errors.New("Invalid feed id"))
	}

	if s.Link == "" {
		return NewValidationError(errors.New("No subscription link"))
	}

	if u, err := url.Parse(s.Link); err != nil || !u.IsAbs() {
		return NewValidationError(errors.New("Invalid subscription link"))
	}

	return nil
}

func (s Subscription) String() string {
	return fmt.Sprintf("%s: %d", s.Link, s.FeedID)
}
