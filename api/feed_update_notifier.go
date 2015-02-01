package api

import (
	"io"
	"net/http"
	"sync"

	"github.com/urandom/readeef"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
	"golang.org/x/net/websocket"
)

type FeedUpdateNotifier struct {
	webfw.BasePatternController
	updateFeed chan readeef.Feed
}

func NewFeedUpdateNotifier() FeedUpdateNotifier {
	return FeedUpdateNotifier{
		BasePatternController: webfw.NewBasePatternController("/v:version/feed-update-notifier", webfw.MethodGet, ""),
		updateFeed:            make(chan readeef.Feed),
	}
}

type feedReceiver struct {
	FeedIds []int64
}

type message struct {
	Success bool
	Message string
}

func (con FeedUpdateNotifier) UpdateFeedChannel() chan<- readeef.Feed {
	return con.updateFeed
}

func (con FeedUpdateNotifier) Handler(c context.Context) http.Handler {
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

	return websocket.Handler(func(ws *websocket.Conn) {
		var err error
		var currentFeeds map[int64]bool

		receiver := make(chan readeef.Feed)

		mutex.Lock()
		receivers[receiver] = true
		mutex.Unlock()
		defer func() {
			mutex.Lock()
			close(receiver)
			delete(receivers, receiver)
			mutex.Unlock()
		}()

		done := make(chan bool)
		defer close(done)

		go func() {
			for {
				select {
				case f := <-receiver:
					readeef.Debug.Println("Received notification for feed update of" + f.Link)

					if currentFeeds[f.Id] {
						resp := map[string]interface{}{"Feed": feed{
							Id: f.Id, Title: f.Title, Description: f.Description,
							Link: f.Link, Image: f.Image,
						}}

						err = websocket.JSON.Send(ws, resp)
						if err != nil {
							webfw.GetLogger(c).Print(err)

							websocket.JSON.Send(ws, message{Success: false, Message: err.Error()})
						}
					}
				case <-done:
					return
				}
			}
		}()

		for {
			var msg feedReceiver
			err = websocket.JSON.Receive(ws, &msg)
			if err != nil {
				// WebSocket is closed
				if err == io.EOF {
					break
				} else {
					websocket.JSON.Send(ws, message{Success: false, Message: err.Error()})
				}
			}

			currentFeeds = map[int64]bool{}

			for _, id := range msg.FeedIds {
				currentFeeds[id] = true
			}
		}

		readeef.Debug.Println("Closing web socket")
	})
}

func (con FeedUpdateNotifier) AuthRequired(c context.Context, r *http.Request) bool {
	return true
}
