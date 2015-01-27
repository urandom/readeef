package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

type FeedUpdateNotificator struct {
	webfw.BasePatternController
	updateFeed <-chan readeef.Feed
}

func NewFeedUpdateNotificator(updateFeed <-chan readeef.Feed) FeedUpdateNotificator {
	return FeedUpdateNotificator{
		BasePatternController: webfw.NewBasePatternController("/v:version/feed-update-notifier", webfw.MethodGet, ""),
		updateFeed:            updateFeed,
	}
}

func (con FeedUpdateNotificator) Handler(c context.Context) http.HandlerFunc {
	var mutex sync.RWMutex

	receivers := make(map[chan readeef.Feed]bool)

	go func() {
		for {
			select {
			case feed := <-con.updateFeed:
				mutex.RLock()

				readeef.Debug.Printf("Feed %s updated. Notifying %d receivers.", feed.Link, len(receivers))
				for receiver, _ := range receivers {
					receiver <- feed
				}

				mutex.RUnlock()
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		receiver := make(chan readeef.Feed)

		mutex.Lock()
		receivers[receiver] = true
		mutex.Unlock()
		defer func() {
			mutex.Lock()
			delete(receivers, receiver)
			mutex.Unlock()
		}()

		f := <-receiver
		readeef.Debug.Println("Feed " + f.Link + " updated")

		resp := map[string]interface{}{"Feed": feed{
			Id: f.Id, Title: f.Title, Description: f.Description,
			Link: f.Link, Image: f.Image,
		}}

		var b []byte
		if err == nil {
			b, err = json.Marshal(resp)
		}
		if err != nil {
			webfw.GetLogger(c).Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)
	}
}

func (con FeedUpdateNotificator) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
