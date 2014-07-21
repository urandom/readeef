package readeef

import (
	"errors"
	"net/url"
	"time"
)

type Hubbub struct {
	config Config
	db     DB
}

type HubbubSubscription struct {
	Id               int64
	Link             string
	FeedLink         string        `db:"feed_link"`
	LeaseDuration    time.Duration `db:"lease_duration"`
	VerificationTime time.Time     `db:"verification_time"`
}

var (
	ErrNotConfigured = errors.New("Hubbub callback URL is not set")
	ErrNoFeedHubLink = errors.New("Feed does not contain a hub link")
)

func NewHubbub(db DB, c Config) Hubbub {
	return Hubbub{db: db, config: c}
}

func (h Hubbub) Subscribe(f Feed) error {
	if u, err := url.Parse(h.config.Hubbub.CallbackURL); err != nil {
		return ErrNotConfigured
	} else {
		if !u.IsAbs() {
			return ErrNotConfigured
		}
	}

	if u, err := url.Parse(f.HubLink); err != nil {
		return ErrNoFeedHubLink
	} else {
		if !u.IsAbs() {
			return ErrNoFeedHubLink
		}
	}

	return nil
}

func (s HubbubSubscription) Validate() error {
	if s.Link == "" {
		return ValidationError{errors.New("Invalid subscription link")}
	}

	if s.FeedLink == "" {
		return ValidationError{errors.New("Invalid feed link")}
	}

	return nil
}
