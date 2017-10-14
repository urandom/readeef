package eventable

import (
	"context"

	"github.com/urandom/readeef/content"
)

type Event struct {
	Name string
	Data interface{}
}

type UserData interface {
	UserLogin() content.Login
}

type FeedData interface {
	FeedID() content.FeedID
}

type Stream chan Event

type busCall func(*busPayload)

type busPayload struct {
	listeners []Stream
}

type bus struct {
	ops chan busCall
}

func newBus(ctx context.Context) bus {
	b := bus{
		ops: make(chan busCall),
	}

	go b.loop(ctx)

	return b
}

func (b bus) Dispatch(name string, data interface{}) {
	b.ops <- func(p *busPayload) {
		event := Event{name, data}
		for i := range p.listeners {
			p.listeners[i] <- event
		}
	}
}

func (b bus) Listener() Stream {
	ret := make(chan Event, 10)

	b.ops <- func(p *busPayload) {
		p.listeners = append(p.listeners, ret)
	}

	return ret
}

func (b bus) loop(ctx context.Context) {
	payload := busPayload{
		[]Stream{},
	}

	for {
		select {
		case op := <-b.ops:
			op(&payload)
		case <-ctx.Done():
			return
		}
	}
}
