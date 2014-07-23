package readeef

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/urandom/webfw/util"
)

type Hubbub struct {
	config Config
	db     DB
	client *http.Client
}

type SubscriptionError struct {
	error
	Subscription HubbubSubscription
}

type HubbubSubscription struct {
	Link                string
	FeedLink            string        `db:"feed_link"`
	LeaseDuration       time.Duration `db:"lease_duration"`
	VerificationTime    time.Time     `db:"verification_time"`
	SubscriptionFailure bool          `db:"subscription_failure"`

	hubbub Hubbub
}

var (
	ErrNotConfigured = errors.New("Hubbub callback URL is not set")
	ErrNoFeedHubLink = errors.New("Feed does not contain a hub link")
)

func NewHubbub(db DB, c Config) Hubbub {
	return Hubbub{db: db, config: c, client: &http.Client{}}
}

func (h Hubbub) SetClient(c http.Client) {
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

	s := HubbubSubscription{Link: f.HubLink, FeedLink: f.Link, hubbub: h, SubscriptionFailure: true}

	if err := h.db.UpdateHubbubSubscription(s); err != nil {
		return err
	}

	return nil
}

func (s HubbubSubscription) Subscription(subscribe bool) error {
	u := callbackURL(s.hubbub.config, s.FeedLink)

	body := url.Values{}
	body.Set("hub.callback", u)
	if subscribe {
		body.Set("hub.mode", "subscribe")
	} else {
		body.Set("hub.mode", "unsubscribe")
	}
	body.Set("hub.topic", s.FeedLink)

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.WriteString(body.Encode())
	req, _ := http.NewRequest("POST", s.Link, buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("From", s.hubbub.config.Hubbub.From)

	resp, err := s.hubbub.client.Do(req)

	if err != nil {
		return SubscriptionError{error: err, Subscription: s}
	} else if resp.StatusCode != 202 {
		return SubscriptionError{error: errors.New(resp.Status), Subscription: s}
	}

	s.SubscriptionFailure = false
	s.hubbub.db.UpdateHubbubSubscription(s)

	return nil
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

func callbackURL(c Config, link string) string {
	return fmt.Sprintf("%s/v%s/%s/%s", c.Hubbub.CallbackURL, c.API.Version, c.Hubbub.RelativePath, url.QueryEscape(link))
}
