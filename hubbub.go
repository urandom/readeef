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

	s := HubbubSubscription{Link: f.HubLink, FeedLink: f.Link}

	return h.db.UpdateHubbubSubscription(s)
}

func (s HubbubSubscription) Validate() error {
	if u, err := url.Parse(s.Link); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Invalid subscription link")}
	}

	if u, err := url.Parse(s.FeedLink); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Invalid feed link")}
	}

	return nil
}
