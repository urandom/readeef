package base

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/tests"
)

func TestDomain(t *testing.T) {
	d := Domain{}

	tests.CheckBool(t, true, d.URL() == nil)

	urlString := "https://sugr.org"
	tests.CheckBool(t, true, d.URL(urlString) != nil)
	tests.CheckBool(t, false, d.HasErr())
	tests.CheckBool(t, true, d.CheckHTTPSSupport(), "The schema contains https")

	tests.CheckString(t, urlString, d.URL().String())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "1")
	}))
	defer ts.Close()

	tests.CheckBool(t, true, d.URL(ts.URL+"/example") != nil)
	tests.CheckBool(t, false, d.HasErr())
	tests.CheckBool(t, false, d.CheckHTTPSSupport(), "Test server is not TLS")

	resp := make(chan bool)
	tlsts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.URL.Path == "" || r.URL.Path == "/") && r.Method == "HEAD" {
			resp <- true
		} else {
			resp <- false
		}
	}))
	defer tlsts.Close()

	tlsts.TLS = new(tls.Config)
	tlsts.TLS.InsecureSkipVerify = true
	tlsts.StartTLS()

	timeoutClient = &http.Client{
		Transport: &http.Transport{
			Dial:            readeef.TimeoutDialer(1*time.Second, 1*time.Second),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		tests.CheckBool(t, true, <-resp, "The Server should've been queries using a HEAD method with no Path")
	}()

	u, err := url.Parse(tlsts.URL)
	tests.CheckBool(t, true, err == nil)
	u.Scheme = "http"
	u.Path = "example"
	tests.CheckBool(t, true, d.URL(u.String()) != nil)
	tests.CheckBool(t, false, d.HasErr())
	tests.CheckBool(t, true, d.CheckHTTPSSupport(), "Test server is TLS")

	wg.Wait()
}
