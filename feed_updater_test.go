package readeef

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"readeef/parser"
	"strings"
	"testing"
	"time"
)

func TestFeedUpdater(t *testing.T) {
	var ts *httptest.Server

	done := make(chan bool)
	updateFeed := make(chan Feed)

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { done <- true }()

		if r.RequestURI == "/link" {
			w.WriteHeader(http.StatusOK)

			w.Write([]byte(strings.Replace(atomXml, "{{ .FeedLink }}", ts.URL+"/link", -1)))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	go func() {
		for {
			select {
			case <-updateFeed:
			}
		}
	}()

	conf, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}

	conf.Updater.Converted.Interval = 50 * time.Millisecond
	fu := NewFeedUpdater(db, conf, log.New(os.Stderr, "", 0), updateFeed)

	fu.Start()

	f := Feed{Feed: parser.Feed{Link: ts.URL + "/link"}}
	fu.AddFeed(f)

	<-done // First update request
	<-done // Second update request

	cleanDB(t)
}
