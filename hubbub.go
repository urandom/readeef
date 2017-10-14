package readeef

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/pool"
)

type Hubbub struct {
	service       repo.Service
	config        config.Config
	endpoint      string
	subscribe     chan content.Subscription
	unsubscribe   chan content.Subscription
	client        *http.Client
	log           log.Log
	subscriptions []content.Subscription
	feedManager   *FeedManager
}

type SubscriptionError struct {
	error
	Subscription content.Subscription
}

var (
	ErrNotConfigured = errors.New("Hubbub callback URL is not set")
	ErrNoFeedHubLink = errors.New("Feed does not contain a hub link")
	ErrSubscribed    = errors.New("Feed already subscribed")
	ErrNotSubscribed = errors.New("Feed is not subscribed")
)

func NewHubbub(
	service repo.Service,
	c config.Config,
	l log.Log,
	endpoint string,
	feedManager *FeedManager,
) *Hubbub {

	return &Hubbub{
		service: service,
		config:  c, log: l, endpoint: endpoint,
		subscribe: make(chan content.Subscription), unsubscribe: make(chan content.Subscription),
		client:      NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite),
		feedManager: feedManager,
	}
}

func (h *Hubbub) Subscribe(f content.Feed) error {
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

	repo := h.service.SubscriptionRepo()
	s, err := repo.Get(f)
	if err != nil && !content.IsNoContent(err) {
		return errors.WithMessage(err, "getting feed subscription during subscribe")
	}

	if s.FeedID == f.ID {
		h.log.Infoln("Already subscribed to " + f.HubLink)
		return ErrSubscribed
	}

	s.Link = f.HubLink
	s.FeedID = f.ID
	s.SubscriptionFailure = true

	if err = repo.Update(s); err != nil {
		return errors.WithMessage(err, "updating subscription during subscribe")
	}

	go func() {
		h.subscription(s, f, true)
	}()

	return nil
}

func (h *Hubbub) Unsubscribe(f content.Feed) error {
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

	s, err := h.service.SubscriptionRepo().Get(f)
	if err != nil {
		return errors.WithMessage(err, "getting feed subscription during unsubscribe")
	}

	if s.FeedID != f.ID {
		h.log.Infoln("Not subscribed to " + f.HubLink)
		return ErrNotSubscribed
	}

	go func() {
		h.subscription(s, f, false)
	}()
	return nil
}

func (h *Hubbub) InitSubscriptions() error {
	subscriptions, err := h.service.SubscriptionRepo().All()
	if err != nil {
		return errors.WithMessage(err, "getting subscriptions")
	}

	h.log.Infof("Initializing %d hubbub subscriptions", len(subscriptions))

	feedRepo := h.service.FeedRepo()
	go func() {
		for _, s := range subscriptions {
			f, err := feedRepo.Get(s.FeedID, content.User{})
			if err != nil {
				h.log.Printf("Error getting subscription feed: %+v\n", err)
				continue
			}

			h.subscription(s, f, true)
		}
	}()

	go func() {
		after := time.After(h.config.FeedManager.Converted.UpdateInterval)
		subscriptions := []content.Subscription{}

		for {
			select {
			case s := <-h.subscribe:
				subscriptions = append(subscriptions, s)
			case s := <-h.unsubscribe:
				filtered := []content.Subscription{}
				for i := range subscriptions {
					if subscriptions[i] != s {
						filtered = append(filtered, subscriptions[i])
					}
				}

				subscriptions = filtered
			case <-after:
				for _, s := range subscriptions {
					if s.VerificationTime.Add(time.Duration(s.LeaseDuration)).Before(time.Now().Add(-30 * time.Minute)) {
						f, err := feedRepo.Get(s.FeedID, content.User{})
						if err != nil {
							h.log.Printf("Error getting subscription feed: %+v\n", err)
							continue
						}

						h.log.Infof("Renewing subscription to %s\n", s)
						h.subscription(s, f, true)
					}
				}
			}
		}
	}()

	return nil
}

func (h *Hubbub) subscription(s content.Subscription, f content.Feed, subscribe bool) {
	var err error

	u := callbackURL(h.config, h.endpoint, f.ID)

	body := url.Values{}
	body.Set("hub.callback", u)
	if subscribe {
		h.log.Infoln("Subscribing to hubbub for " + f.String() + " with url " + u)
		body.Set("hub.mode", "subscribe")
	} else {
		h.log.Infoln("Unsubscribing to hubbub for " + f.String() + " with url " + u)
		body.Set("hub.mode", "unsubscribe")
	}
	body.Set("hub.topic", f.Link)

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	buf.WriteString(body.Encode())
	req, _ := http.NewRequest("POST", s.Link, buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("From", h.config.Hubbub.From)

	resp, err := h.client.Do(req)

	if err != nil {
		err = SubscriptionError{error: err, Subscription: s}
	} else if resp.StatusCode != 202 {
		err = SubscriptionError{error: errors.New("Expected response status 202, got " + resp.Status), Subscription: s}
	}

	if err == nil {
		if subscribe {
			h.subscribe <- s
		} else {
			h.unsubscribe <- s
		}
	} else {
		f.SubscribeError = fmt.Sprintf("%s: %s", time.Now().Format(time.UnixDate), err.Error())
		h.log.Printf("Error subscribing to hub feed '%s': %s\n", f, err)

		if _, err = h.service.FeedRepo().Update(&f); err != nil {
			h.log.Printf("Error updating feed database record for %s: %+v", f, err)
		}
	}
}

func callbackURL(c config.Config, endpoint string, feedID content.FeedID) string {
	return fmt.Sprintf("%s%s/%d", c.Hubbub.CallbackURL, endpoint, feedID)
}
