package readeef

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Nonce struct {
	store map[string]int64
	mutex sync.RWMutex
}

func NewNonce() *Nonce {
	return &Nonce{
		store: make(map[string]int64),
	}
}

func (n *Nonce) Generate() string {
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(time.Now().Unix(), 32))
	io.WriteString(h, strconv.FormatInt(rand.Int63(), 32))

	nonce := fmt.Sprintf("%x", h.Sum(nil))

	return nonce
}

func (n *Nonce) Set(nonce string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.store[nonce] = time.Now().Unix()
}

func (n *Nonce) Check(nonce string) bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	_, ok := n.store[nonce]

	return ok
}

func (n *Nonce) Remove(nonce string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	delete(n.store, nonce)
}

func (n *Nonce) Clean(age time.Duration) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	now := time.Now().Unix()
	for nonce, timestamp := range n.store {
		if now-timestamp > int64(age) {
			delete(n.store, nonce)
		}
	}
}
