package readeef

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type Hubbub struct {
	*UpdateFeedReceiverManager
	config     Config
	repo       content.Repo
	pattern    string
	addFeed    chan<- content.Feed
	removeFeed chan<- content.Feed
	client     *http.Client
	logger     webfw.Logger
}

type SubscriptionError struct {
	error
	Subscription content.Subscription
}

type HubbubController struct {
	webfw.BasePatternController
	hubbub *Hubbub
}

var (
	ErrNotConfigured = errors.New("Hubbub callback URL is not set")
	ErrNoFeedHubLink = errors.New("Feed does not contain a hub link")
	ErrSubscribed    = errors.New("Feed already subscribed")
)

func NewHubbub(repo content.Repo, c Config, l webfw.Logger, pattern string, addFeed chan<- content.Feed, removeFeed chan<- content.Feed, um *UpdateFeedReceiverManager) *Hubbub {
	return &Hubbub{
		UpdateFeedReceiverManager: um,
		repo: repo, config: c, logger: l, pattern: pattern,
		addFeed: addFeed, removeFeed: removeFeed,
		client: NewTimeoutClient(c.Timeout.Converted.Connect, c.Timeout.Converted.ReadWrite)}
}

func (h *Hubbub) SetClient(c *http.Client) {
	h.client = c
}

func (h *Hubbub) Subscribe(f content.Feed) error {
	if u, err := url.Parse(h.config.Hubbub.CallbackURL); err != nil {
		return ErrNotConfigured
	} else {
		if !u.IsAbs() {
			return ErrNotConfigured
		}
	}

	finfo := f.Info()
	if u, err := url.Parse(finfo.HubLink); err != nil {
		return ErrNoFeedHubLink
	} else {
		if !u.IsAbs() {
			return ErrNoFeedHubLink
		}
	}

	s := f.Subscription()
	if s.HasErr() {
		return s.Err()
	}

	info := s.Info()
	if info.FeedId == finfo.Id {
		h.logger.Infoln("Already subscribed to " + finfo.HubLink)
		return ErrSubscribed
	}

	info.Link = finfo.HubLink
	info.FeedId = finfo.Id
	info.SubscriptionFailure = true

	s.Info(info)
	s.Update()

	if s.HasErr() {
		return s.Err()
	}

	go func() {
		h.subscribe(s, f, true)
	}()

	return nil
}

func (h *Hubbub) InitSubscriptions() error {
	h.repo.FailSubscriptions()
	subscriptions := h.repo.AllSubscriptions()

	h.logger.Infof("Initializing %d hubbub subscriptions", len(subscriptions))

	go func() {
		for _, s := range subscriptions {
			f := h.repo.FeedById(s.Info().FeedId)
			if f.Err() != nil {
				continue
			}

			h.subscribe(s, f, true)
		}
	}()

	return h.repo.Err()
}

func (h *Hubbub) subscribe(s content.Subscription, f content.Feed, subscribe bool) {
	var err error

	finfo := f.Info()
	u := callbackURL(h.config, h.pattern, finfo.Id)

	body := url.Values{}
	body.Set("hub.callback", u)
	if subscribe {
		h.logger.Infoln("Subscribing to hubbub for " + f.String() + " with url " + u)
		body.Set("hub.mode", "subscribe")
	} else {
		h.logger.Infoln("Unsubscribing to hubbub for " + f.String() + " with url " + u)
		body.Set("hub.mode", "unsubscribe")
	}
	body.Set("hub.topic", finfo.Link)

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.WriteString(body.Encode())
	req, _ := http.NewRequest("POST", s.Info().Link, buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("From", h.config.Hubbub.From)

	resp, err := h.client.Do(req)

	if err != nil {
		err = SubscriptionError{error: err, Subscription: s}
	} else if resp.StatusCode != 202 {
		err = SubscriptionError{error: errors.New("Expected response status 202, got " + resp.Status), Subscription: s}
	}

	if err != nil {
		finfo.SubscribeError = err.Error()
		h.logger.Printf("Error subscribing to hub feed '%s': %s\n", f, err)

		f.Info(finfo)
		f.Update()
		if f.HasErr() {
			h.logger.Printf("Error updating feed database record for '%s': %s\n", f, f.Err())
		}

		h.removeFeed <- f
	}
}

func NewHubbubController(h *Hubbub) HubbubController {
	return HubbubController{
		webfw.NewBasePatternController(h.config.Hubbub.RelativePath+"/:feed-id", webfw.MethodGet|webfw.MethodPost, "hubbub-callback"),
		h}
}

func (con HubbubController) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		pathParams := webfw.GetParams(c, r)
		logger := webfw.GetLogger(c)
		feedId, err := strconv.ParseInt(pathParams["feed-id"], 10, 64)

		if err != nil {
			webfw.GetLogger(c).Print(err)
			return
		}

		repo := con.hubbub.repo
		f := repo.FeedById(info.FeedId(feedId))
		s := f.Subscription()

		err = s.Err()

		if err != nil {
			webfw.GetLogger(c).Print(err)
			return
		}

		logger.Infoln("Receiving hubbub event " + params.Get("hub.mode") + " for " + f.String())

		info := s.Info()
		switch params.Get("hub.mode") {
		case "subscribe":
			if lease, err := strconv.Atoi(params.Get("hub.lease_seconds")); err == nil {
				info.LeaseDuration = int64(lease) * int64(time.Second)
			}
			info.VerificationTime = time.Now()

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
				f.Refresh(pf)
				f.Update()

				if f.HasErr() {
					webfw.GetLogger(c).Print(f.Err())
					return
				}

				newArticles = len(f.NewArticles()) > 0
			} else {
				webfw.GetLogger(c).Print(err)
				return
			}

			if newArticles {
				con.hubbub.NotifyReceivers(f)
			}

			return
		}

		switch params.Get("hub.mode") {
		case "subscribe":
			info.SubscriptionFailure = false
		case "unsubscribe", "denied":
			info.SubscriptionFailure = true
		}

		s.Info(info)
		s.Update()
		if s.HasErr() {
			webfw.GetLogger(c).Print(s.Err())
			return
		}

		if info.SubscriptionFailure {
			con.hubbub.removeFeed <- f
		} else {
			con.hubbub.addFeed <- f
		}
	})
}

func callbackURL(c Config, pattern string, feedId info.FeedId) string {
	return fmt.Sprintf("%s%sv%d%s/%d", c.Hubbub.CallbackURL, pattern, c.API.Version, c.Hubbub.RelativePath, feedId)
}
