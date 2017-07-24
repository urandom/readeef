package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func eventSocket(
	ctx context.Context,
	storage content.TokenStorage,
	feedManager *readeef.FeedManager,
) http.HandlerFunc {
	monitor := &feedMonitor{ops: make(chan func(connMap))}

	go monitor.loop(ctx)
	feedManager.AddFeedMonitor(monitor)

	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusNotAcceptable)
			return
		}

		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feeds := user.AllFeeds()
		if user.HasErr() {
			http.Error(w, "Error getting user feeds: "+user.Err().Error(), http.StatusInternalServerError)

			return
		}

		feedSet := feedSet{}
		for i := range feeds {
			feedSet[feeds[i].Data().Id] = struct{}{}
		}
		feeds = nil

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		defer monitor.removeConn(r.RemoteAddr)

		done := monitor.addConn(w, flusher, connectionValidator(storage, r), r.RemoteAddr, feedSet)

		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		}
	}
}

func connectionValidator(storage content.TokenStorage, r *http.Request) func() bool {
	return func() bool {
		if token := auth.Token(r); token != "" {
			if exists, err := storage.Exists(token); err == nil && exists {
				return false
			}
		}
		if claims, ok := auth.Claims(r).(*jwt.StandardClaims); ok {
			return claims.Valid() == nil
		}

		return false
	}
}

type feedMonitor struct {
	ops chan func(connMap)
	log readeef.Logger
}

type connMap map[string]connData
type feedSet map[data.FeedId]struct{}

type connData struct {
	writer    io.Writer
	flusher   http.Flusher
	validator func() bool
	feedSet   feedSet
	done      chan struct{}
}

type event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (e event) WriteJSON(w io.Writer, flusher http.Flusher, data interface{}, log readeef.Logger) error {
	if b, err := json.Marshal(data); err == nil {
		e.Data = json.RawMessage(b)

		if b, err = json.Marshal(e); err == nil {
			_, err := w.Write(b)
			if err != nil {
				return err
			}

			flusher.Flush()
		} else {
			log.Printf("Error converting event to json: %+v", err)
		}
	} else {
		log.Printf("Error converting %#v to json: %+v", data, err)
	}

	return nil
}

func (fm *feedMonitor) FeedUpdated(feed content.Feed) error {
	fm.ops <- func(conns connMap) {
		for _, d := range conns {
			if _, ok := d.feedSet[feed.Data().Id]; ok {
				if !d.validator() {
					close(d.done)
					continue
				}

				err := event{Type: "feed-update"}.WriteJSON(d.writer, d.flusher, args{"feed": feed}, fm.log)
				if err != nil {
					close(d.done)
				}
			}
		}
	}

	return nil
}

func (fm *feedMonitor) FeedDeleted(feed content.Feed) error {
	return nil
}

func (fm *feedMonitor) addConn(w io.Writer, flusher http.Flusher, validator func() bool, addr string, feedSet feedSet) chan struct{} {
	done := make(chan struct{})

	fm.ops <- func(conns connMap) {
		conns[addr] = connData{w, flusher, validator, feedSet, done}
	}

	return done
}

func (fm *feedMonitor) removeConn(addr string) {
	fm.ops <- func(conns connMap) {
		delete(conns, addr)
	}
}

func (fm *feedMonitor) loop(ctx context.Context) {
	conns := make(connMap)

	for {
		select {
		case <-ctx.Done():
			return
		case op := <-fm.ops:
			op(conns)
		}
	}
}
