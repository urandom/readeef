package readeef

import (
	"net"
	"net/http"
	"time"
)

func TimeoutDialer(ct time.Duration, rwt time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, ct)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwt))
		return conn, nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}
