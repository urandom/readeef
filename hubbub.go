package readeef

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type Hubbub struct {
	config     Config
	db         DB
	pattern    string
	addFeed    chan<- Feed
	removeFeed chan<- Feed
	updateFeed chan<- Feed
	client     *http.Client
	logger     *log.Logger
}

type SubscriptionError struct {
	error
	Subscription *HubbubSubscription
}

type HubbubSubscription struct {
	Link                string
	FeedId              int64     `db:"feed_id"`
	LeaseDuration       int64     `db:"lease_duration"`
	VerificationTime    time.Time `db:"verification_time"`
	SubscriptionFailure bool      `db:"subscription_failure"`

	hubbub *Hubbub
}

type HubbubController struct {
	webfw.BaseController
	hubbub *Hubbub
}

var (
	ErrNotConfigured = errors.New("Hubbub callback URL is not set")
	ErrNoFeedHubLink = errors.New("Feed does not contain a hub link")
	ErrSubscribed    = errors.New("Feed already subscribed")
)

func NewHubbub(db DB, c Config, l *log.Logger, pattern string, addFeed chan<- Feed, removeFeed chan<- Feed, updateFeed chan<- Feed) *Hubbub {
	return &Hubbub{db: db, config: c, logger: l, pattern: pattern,
		addFeed: addFeed, removeFeed: removeFeed, updateFeed: updateFeed,
		client: NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite)}
}

func (h *Hubbub) SetClient(c *http.Client) {
	h.client = c
}

func (h *Hubbub) Subscribe(f Feed) error {
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

	if _, err := h.db.GetHubbubSubscription(f.Id); err == nil {
		Debug.Println("Already subscribed to " + f.HubLink)
		return ErrSubscribed
	}

	s := &HubbubSubscription{Link: f.HubLink, FeedId: f.Id, hubbub: h, SubscriptionFailure: true}

	if err := h.db.UpdateHubbubSubscription(s); err != nil {
		return err
	}

	go func() {
		h.subscribe(s, f)
	}()

	return nil
}

func (h *Hubbub) InitSubscriptions() error {
	if err := h.db.FailHubbubSubscriptions(); err != nil {
		return err
	}

	subscriptions, err := h.db.GetHubbubSubscriptions()
	if err != nil {
		return err
	}

	Debug.Printf("Initializing %d hubbub subscriptions", len(subscriptions))

	go func() {
		for _, s := range subscriptions {
			f, err := h.db.GetFeed(s.FeedId)
			if err != nil {
				continue
			}

			s.hubbub = h

			h.subscribe(s, f)
		}
	}()

	return nil
}

func (h *Hubbub) subscribe(s *HubbubSubscription, f Feed) {
	err := s.Subscribe()
	if err != nil {
		f.SubscribeError = err.Error()
		h.logger.Printf("Error subscribing to hub feed '%s': %s\n", f.Link, err)

		if _, _, err := h.db.UpdateFeed(f); err != nil {
			h.logger.Printf("Error updating feed database record for '%s': %s\n", f.Link, err)
		}

		h.removeFeed <- f
	}
}

func (s *HubbubSubscription) Subscribe() error {
	return s.subscription(true)
}

func (s *HubbubSubscription) Unsubscribe() error {
	return s.subscription(false)
}

func (s *HubbubSubscription) Validate() error {
	if u, err := url.Parse(s.Link); err != nil || !u.IsAbs() {
		return ValidationError{errors.New("Invalid subscription link")}
	}

	if s.FeedId == 0 {
		return ValidationError{errors.New("Invalid feed id")}
	}

	return nil
}

func (s *HubbubSubscription) subscription(subscribe bool) error {
	u := callbackURL(s.hubbub.config, s.hubbub.pattern, s.FeedId)
	feed, err := s.hubbub.db.GetFeed(s.FeedId)
	if err != nil {
		return SubscriptionError{error: err, Subscription: s}
	}

	body := url.Values{}
	body.Set("hub.callback", u)
	if subscribe {
		Debug.Println("Subscribing to hubbub for " + feed.Link + " with url " + u)
		body.Set("hub.mode", "subscribe")
	} else {
		Debug.Println("Unsubscribing to hubbub for " + feed.Link + " with url " + u)
		body.Set("hub.mode", "unsubscribe")
	}
	body.Set("hub.topic", feed.Link)

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
		return SubscriptionError{error: errors.New("Expected response status 202, got " + resp.Status), Subscription: s}
	}

	return nil
}

func NewHubbubController(h *Hubbub) HubbubController {
	return HubbubController{
		webfw.NewBaseController(h.config.Hubbub.RelativePath+"/:feed-id", webfw.MethodGet|webfw.MethodPost, "hubbub-callback"),
		h}
}

func (con HubbubController) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		params := r.URL.Query()
		pathParams := webfw.GetParams(c, r)
		feedId, err := strconv.ParseInt(pathParams["feed-id"], 10, 64)

		if err != nil {
			webfw.GetLogger(c).Print(err)
			return
		}

		s, err := con.hubbub.db.GetHubbubSubscription(feedId)

		if err != nil {
			webfw.GetLogger(c).Print(err)
			return
		}

		f, err := con.hubbub.db.GetFeed(s.FeedId)
		if err != nil {
			webfw.GetLogger(c).Print(err)
			return
		}

		Debug.Println("Receiving hubbub event " + params.Get("hub.mode") + " for " + f.Link)

		switch params.Get("hub.mode") {
		case "subscribe":
			if lease, err := strconv.Atoi(params.Get("hub.lease_seconds")); err == nil {
				s.LeaseDuration = int64(lease) * int64(time.Second)
			}
			s.VerificationTime = time.Now()

			w.Write([]byte(params.Get("hub.challenge")))
		case "unsubscribe":
			w.Write([]byte(params.Get("hub.challenge")))
		case "denied":
			w.Write([]byte{})
			webfw.GetLogger(c).Printf("Unable to subscribe to '%s': %s\n", params.Get("hub.topic"), params.Get("hub.reason"))
		default:
			w.Write([]byte{})

			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			if _, err := buf.ReadFrom(r.Body); err != nil {
				webfw.GetLogger(c).Print(err)
				return
			}

			newArticles := false

			if pf, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
				f = f.UpdateFromParsed(pf)

				_, newArticles, err = con.hubbub.db.UpdateFeed(f)
				if err != nil {
					webfw.GetLogger(c).Print(err)
					return
				}
			} else {
				webfw.GetLogger(c).Print(err)
				return
			}

			if newArticles {
				con.hubbub.updateFeed <- f
			}

			return
		}

		switch params.Get("hub.mode") {
		case "subscribe":
			s.SubscriptionFailure = false
		case "unsubscribe", "denied":
			s.SubscriptionFailure = true
		}

		if err := con.hubbub.db.UpdateHubbubSubscription(s); err != nil {
			webfw.GetLogger(c).Print(err)
			return
		}

		if s.SubscriptionFailure {
			con.hubbub.removeFeed <- f
		} else {
			con.hubbub.addFeed <- f
		}
	}
}

func callbackURL(c Config, pattern string, feedId int64) string {
	return fmt.Sprintf("%s%sv%d%s/%d", c.Hubbub.CallbackURL, pattern, c.API.Version, c.Hubbub.RelativePath, feedId)
}
