package api

import (
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/pool"
)

func hubbubRegistration(
	hubbub *readeef.Hubbub,
	repo repo.Feed,
	feedManager *readeef.FeedManager,
	log log.Log,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		feedID, err := strconv.ParseInt(path.Base(r.URL.Path), 10, 64)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f := repo.Get(content.FeedID(feedID), content.User{})
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
			data.SubscriptionFailure = false

			w.Write([]byte(params.Get("hub.challenge")))
		case "unsubscribe":
			data.SubscriptionFailure = true

			w.Write([]byte(params.Get("hub.challenge")))
		case "denied":
			w.Write([]byte{})
			log.Printf("Unable to subscribe to '%s': %s\n", params.Get("hub.topic"), params.Get("hub.reason"))
		default:
			w.Write([]byte{})

			buf := pool.Buffer.Get()
			defer pool.Buffer.Put(buf)

			if _, err := buf.ReadFrom(r.Body); err != nil {
				log.Print(err)
				return
			}

			if pf, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
				f.Refresh(pf)
				f.Update()

				if f.HasErr() {
					log.Print(f.Err())
					return
				}
			} else {
				log.Print(err)
				return
			}

			hubbub.ProcessFeedUpdate(f)

			return
		}

		s.Data(data)
		s.Update()
		if s.HasErr() {
			log.Printf("Error updating subscription %s: %+v\n", s, s.Err())
			return
		}
	}
}
