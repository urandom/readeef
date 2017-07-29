package api

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/webfw/util"
)

func hubbubRegistration(
	hubbub *readeef.Hubbub,
	repo content.Repo,
	addFeed chan<- content.Feed,
	removeFeed chan<- content.Feed,
	log readeef.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		feedId, err := strconv.ParseInt(path.Base(r.URL.Path), 10, 64)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f := repo.FeedById(data.FeedId(feedId))
		s := f.Subscription()

		if s.HasErr() {
			http.Error(w, s.Err().Error(), http.StatusInternalServerError)
			return
		}

		log.Infoln("Receiving hubbub event " + params.Get("hub.mode") + " for " + f.String())

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
			log.Printf("Unable to subscribe to '%s': %s\n", params.Get("hub.topic"), params.Get("hub.reason"))
		default:
			w.Write([]byte{})

			buf := util.BufferPool.GetBuffer()
			defer util.BufferPool.Put(buf)

			if _, err := buf.ReadFrom(r.Body); err != nil {
				log.Print(err)
				return
			}

			newArticles := false

			if pf, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
				f.Refresh(pf)
				f.Update()

				if f.HasErr() {
					log.Print(f.Err())
					return
				}

				newArticles = len(f.NewArticles()) > 0
			} else {
				log.Print(err)
				return
			}

			if newArticles {
				if err := hubbub.ProcessFeedUpdate(f); err != nil {
					+log.Printf("Error processing feed update for '%s': %v\n", f, err)

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
			log.Print(fmt.Errorf("Error updating subscription %s: %v\n", s, s.Err()))
			return
		}

		if data.SubscriptionFailure {
			removeFeed <- f
		} else {
			addFeed <- f
		}
	}
}
