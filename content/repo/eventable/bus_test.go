package eventable

import (
	"context"
	"testing"
	"time"

	"github.com/urandom/readeef/content"
)

type event1data struct {
	data int
}

func (e event1data) UserLogin() content.Login {
	return "user1"
}

type event2data struct {
	data string
}

func (e event2data) UserLogin() content.Login {
	return "user2"
}

func Test_bus_Dispatch(t *testing.T) {
	type args struct {
		name string
		data UserData
	}
	tests := []struct {
		name      string
		events    []args
		listeners int
	}{
		{"single event", []args{args{"event1", event1data{42}}}, 1},
		{"multiple event", []args{
			args{"event1", event1data{42}},
			args{"event2", event2data{"event2"}},
		}, 1},
		{"single event, multi listeners", []args{args{"event1", event1data{42}}}, 3},
		{"multiple event, multi listeners", []args{
			args{"event1", event1data{42}},
			args{"event2", event2data{"event2"}},
		}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			b := newBus(ctx)

			done := make(chan struct{})

			for i := 0; i < tt.listeners; i++ {
				go func() {
					l := b.Listener()
					i := 0

					for {
						select {
						case e := <-l:
							if e.Name != tt.events[i].name || e.Data != tt.events[i].data {
								t.Errorf("Test_bus_Dispatch(), expected %#v, got %#v", e, tt.events[i])
								return
							}
							i++
						case <-done:
							if i != len(tt.events) {
								t.Errorf("Received events = %d do not match expected = %d", i, len(tt.events))
							}
							return
						case <-ctx.Done():
							return
						}
					}
				}()
			}

			time.Sleep(500 * time.Millisecond)

			for _, e := range tt.events {
				b.Dispatch(e.name, e.data)
			}

			time.Sleep(500 * time.Millisecond)

			close(done)

			time.Sleep(500 * time.Millisecond)
		})
	}
}
