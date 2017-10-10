package api

import "context"

type busCall func(*busPayload)

type busPayload struct {
	listeners []chan event
}

type bus struct {
	ops chan busCall
}

func NewBus(ctx context.Context) bus {
	b := bus{
		ops: make(chan busCall),
	}

	go b.loop(ctx)

	return b
}

func (b bus) Dispatch(name string, data interface{}) {
	b.ops <- func(p *busPayload) {
		event := event{name, data}
		for i := range p.listeners {
			p.listeners[i] <- event
		}
	}
}

func (b bus) Listener() chan event {
	ret := make(chan event)

	b.ops <- func(p *busPayload) {
		p.listeners = append(p.listeners, ret)
	}

	return ret
}

func (b bus) loop(ctx context.Context) {
	payload := busPayload{
		[]chan event{},
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
