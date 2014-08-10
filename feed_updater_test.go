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
	cleanDB(t)

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

	conf, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}

	conf.Updater.Converted.Interval = 100 * time.Millisecond
	fu := NewFeedUpdater(db, conf, log.New(os.Stderr, "", 0), updateFeed)

	fu.Start()

	f := Feed{Feed: parser.Feed{Link: ts.URL + "/link"}}
	f, err = db.UpdateFeed(f)
	if err != nil {
		t.Fatal(err)
	}

	fu.AddFeed(f)

	<-done // First update request

	f2 := <-updateFeed // Feed gets updated

	f, err = db.GetFeed(f.Id)
	if err != nil {
		t.Fatal(err)
	}

	if f.SubscribeError != "" {
		t.Fatal(f.SubscribeError)
	}

	expectedStr := "Example Feed"
	if f.Title != expectedStr {
		t.Fatalf("Expected feed title to be '%s', got '%s'\n", expectedStr, f.Title)
	}

	if f.Link != f2.Link || f.Title != f2.Title || f.Description != f2.Description {
		t.Fatal("")
	}

	<-done // Second update request

	fu.RemoveFeed(f)

	time.Sleep(150 * time.Millisecond)
	select {
	case <-done:
		t.Fatalf("Did not expect another request\n")
	default:
	}
}
