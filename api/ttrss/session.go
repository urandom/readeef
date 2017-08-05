package ttrss

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/urandom/readeef/content"
)

type session struct {
	login     content.Login
	lastVisit time.Time
}

type sessMap map[string]session

type sessionManager struct {
	ops chan func(sessMap)
}

func newSession(ctx context.Context) sessionManager {
	sm := sessionManager{
		ops: make(chan func(sessMap)),
	}

	go sm.loop(ctx)

	return sm
}

func (sm sessionManager) add(id string, sess session) {
	sm.ops <- func(m sessMap) {
		m[id] = sess
	}
}

func (sm sessionManager) remove(id string) {
	sm.ops <- func(m sessMap) {
		delete(m, id)
	}
}

func (sm sessionManager) get(id string) session {
	sess := make(chan session)

	sm.ops <- func(m sessMap) {
		if s, ok := m[id]; ok {
			sess <- s
		} else {
			sess <- session{}
		}
	}

	return <-sess
}

func (sm sessionManager) set(id string, sess session) {
	sm.ops <- func(m sessMap) {
		m[id] = sess
	}
}

func (sm sessionManager) update(sess session) string {
	idRes := make(chan string)

	sm.ops <- func(m sessMap) {
		for id, s := range m {
			if sess.login == s.login {
				idRes <- id
				return
			}
		}

		id := uuid()
		m[id] = sess

		idRes <- id
	}

	return <-idRes
}

func (sm sessionManager) loop(ctx context.Context) {
	sessMap := make(sessMap)
	ticker := time.NewTicker(time.Hour)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case op := <-sm.ops:
			op(sessMap)
		case <-ticker.C:
			fiveDaysAgo := time.Now().AddDate(0, 0, -5)
			for id, sess := range sessMap {
				if sess.lastVisit.Before(fiveDaysAgo) {
					delete(sessMap, id)
				}
			}
		}
	}
}

func uuid() string {
	var u [16]byte

	rand.Read(u[:])

	return strings.Replace(fmt.Sprintf("%x-%x-%x-%x-%x",
		u[:4], u[4:6], u[6:8], u[8:10], u[10:]), "-", "", -1)
}
