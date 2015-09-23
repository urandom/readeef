package api

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"github.com/urandom/webfw/util"
)

type HubbubController struct {
	webfw.BasePatternController

	hubbub     *readeef.Hubbub
	addFeed    chan<- content.Feed
	removeFeed chan<- content.Feed
}

func NewHubbubController(h *readeef.Hubbub, relativePath string,
	addFeed chan<- content.Feed, removeFeed chan<- content.Feed) HubbubController {

	return HubbubController{
		BasePatternController: webfw.NewBasePatternController(
			"/v:version"+relativePath+"/:feed-id",
			webfw.MethodGet|webfw.MethodPost, "hubbub-callback",
		),
		hubbub: h, addFeed: addFeed, removeFeed: removeFeed}
}

func (con HubbubController) Handler(c context.Context) http.Handler {
	logger := webfw.GetLogger(c)
	repo := readeef.GetRepo(c)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		pathParams := webfw.GetParams(c, r)
		feedId, err := strconv.ParseInt(pathParams["feed-id"], 10, 64)

		if err != nil {
			logger.Print(err)
			return
		}

		f := repo.FeedById(data.FeedId(feedId))
		s := f.Subscription()

		err = s.Err()

		if err != nil {
			logger.Print(err)
			return
		}

		logger.Infoln("Receiving hubbub event " + params.Get("hub.mode") + " for " + f.String())

		data := s.Data()
		switch params.Get("hub.mode") {
		case "subscribe":
			if lease, err := strconv.Atoi(params.Get("hub.lease_seconds")); err == nil {
				data.LeaseDuration = int64(lease) * int64(time.Second)
			}
			data.VerificationTime = time.Now()

			w.Write([]byte(params.Get("hub.challenge")))
		case "unsubscribe":
			// Nothing to do here, the subscription will be removed along with the feed by the manager
			w.Write([]byte(params.Get("hub.challenge")))
		case "denied":
			w.Write([]byte{})
			logger.Printf("Unable to subscribe to '%s': %s\n", params.Get("hub.topic"), params.Get("hub.reason"))
		default:
			w.Write([]byte{})

			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			if _, err := buf.ReadFrom(r.Body); err != nil {
				logger.Print(err)
				return
			}

			newArticles := false

			if pf, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
				f.Refresh(pf)
				f.Update()

				if f.HasErr() {
					logger.Print(f.Err())
					return
				}

				newArticles = len(f.NewArticles()) > 0
			} else {
				logger.Print(err)
				return
			}

			if newArticles {
				for _, m := range con.hubbub.FeedMonitors() {
					if err := m.FeedUpdated(f); err != nil {
						logger.Printf("Error invoking monitor '%s' on updated feed '%s': %v\n",
							reflect.TypeOf(m), f, err)
					}
				}
			}

			return
		}

		switch params.Get("hub.mode") {
		case "subscribe":
			data.SubscriptionFailure = false
		case "unsubscribe", "denied":
			data.SubscriptionFailure = true
		}

		s.Data(data)
		s.Update()
		if s.HasErr() {
			logger.Print(fmt.Errorf("Error updating subscription %s: %v\n", s, s.Err()))
			return
		}

		if data.SubscriptionFailure {
			con.removeFeed <- f
		} else {
			con.addFeed <- f
		}
	})
}

func (con HubbubController) AuthRequired(c context.Context, r *http.Request) bool {
	return false
}
