package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
	"github.com/urandom/readeef/api/token"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/log"
)

type event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

func eventSocket(
	ctx context.Context,
	service eventable.Service,
	storage token.Storage,
	log log.Log,
) http.HandlerFunc {
	repo := service.FeedRepo()
	monitor := &feedMonitor{ops: make(chan func(connMap), 10), service: service, log: log}

	go monitor.loop(ctx)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Debugln("Event connection initializing")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusNotAcceptable)
			return
		}

		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		feeds, err := repo.ForUser(user)
		if err != nil {
			fatal(w, log, "Error getting user feeds: %+v", err)
			return
		}

		feedSet := feedSet{}
		for i := range feeds {
			feedSet[feeds[i].ID] = struct{}{}
		}
		feeds = nil

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		defer monitor.removeConn(r.RemoteAddr)

		done := monitor.addConn(w, flusher, connectionValidator(storage, r), r.RemoteAddr, feedSet, user)

		log.Debugln("Initializing event stream")
		err = event{Type: "connection-established"}.Write(w, flusher, log)
		if err != nil {
			log.Printf("Error sending initial data: %+v", err)
			return
		}

		for {
			select {
			case <-time.After(10 * time.Second):
				log.Debugln("Sending ping")
				monitor.ping(r.RemoteAddr)
			case <-done:
				log.Debugln("Connection done")
				return
			case <-w.(http.CloseNotifier).CloseNotify():
				log.Debugln("Connection closed")
				return
			case <-ctx.Done():
				log.Debugln("Context cancelled")
				return
			}
		}
	}
}

func connectionValidator(storage token.Storage, r *http.Request) func() bool {
	return func() bool {
		if token := auth.Token(r); token != "" {
			// Check that the claim isn't blacklisted
			if exists, err := storage.Exists(token); err == nil && exists {
				return false
			}
		}

		if claims := auth.Claims(r); claims != nil {
			// Check that the claim hasn't expired
			return claims.Valid() == nil
		}

		return false
	}
}

type feedMonitor struct {
	ops     chan func(connMap)
	service eventable.Service
	log     log.Log
}

type connMap map[string]connData
type feedSet map[content.FeedID]struct{}

type connData struct {
	writer    io.Writer
	flusher   http.Flusher
	validator func() bool
	feedSet   feedSet
	login     content.Login
	done      chan struct{}
}

func (e event) Write(w io.Writer, flusher http.Flusher, log log.Log) error {
	if e.Type == "" {
		// Comment event to keep the connection alive
		if _, err := w.Write([]byte(": ping\n\n")); err != nil {
			return errors.Wrap(err, "sending ping")
		}
		flusher.Flush()
		log.Debug("Wrote ping message")

		return nil
	}

	data := []byte("event: " + e.Type + "\n")
	if e.Data != nil {
		b, err := json.Marshal(e.Data)
		if err != nil {
			log.Printf("Error converting data %#v to json: %+v", e.Data, err)
			return nil
		}

		data = append(data, []byte("data: ")...)
		data = append(data, b...)
		data = append(data, '\n')
	}

	data = append(data, '\n')

	if _, err := w.Write(data); err != nil {
		return errors.Wrapf(err, "sending event %s", string(data))
	}
	flusher.Flush()
	log.Debugf("Wrote and flushed SSE: %s", string(data))

	return nil
}

func (fm *feedMonitor) processEvent(ev eventable.Event) {
	fm.ops <- func(conns connMap) {
		for _, d := range conns {
			if ud, ok := ev.Data.(eventable.UserData); ok && d.login != ud.UserLogin() {
				continue
			}

			if fd, ok := ev.Data.(eventable.FeedData); ok {
				if _, ok := d.feedSet[fd.FeedID()]; !ok {
					continue
				}
			}

			if !d.validator() {
				close(d.done)
				continue
			}

			event := event{Type: ev.Name, Data: ev.Data}

			err := event.Write(d.writer, d.flusher, fm.log)
			if err != nil {
				fm.log.Printf("Error sending article state event: %+v", err)
				close(d.done)
			}
		}
	}
}

func (fm *feedMonitor) ping(addr string) {
	fm.ops <- func(conns connMap) {
		if d, ok := conns[addr]; ok {
			if !d.validator() {
				close(d.done)
				return
			}

			err := event{}.Write(d.writer, d.flusher, fm.log)
			if err != nil {
				fm.log.Printf("Error sending ping event: %+v", err)
				close(d.done)
			}
		}
	}
}

func (fm *feedMonitor) addConn(w io.Writer, flusher http.Flusher, validator func() bool, addr string, feedSet feedSet, user content.User) chan struct{} {
	done := make(chan struct{})

	fm.ops <- func(conns connMap) {
		conns[addr] = connData{w, flusher, validator, feedSet, user.Login, done}
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
	done := make(chan struct{})

	listener := fm.service.Listener()

	closed := false
	for {
		select {
		case <-ctx.Done():
			if !closed {
				closed = true
				// Give some time for the ops channel to drain the connection
				// removal actions before exiting the loop.
				time.AfterFunc(100*time.Millisecond, func() {
					close(done)
				})
			}
		case op := <-fm.ops:
			op(conns)
		case event := <-listener:
			fm.log.Debugf("Got service event %s", event.Name)
			go fm.processEvent(event)
		case <-done:
			return
		}
	}
}
