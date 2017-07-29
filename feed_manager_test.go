package readeef

/*

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/urandom/readeef"
)

func TestFeedManager(t *testing.T) {
	cleanDB(t)

	var ts *httptest.Server

	done := make(chan bool)
	rec := updateReceiver{updateFeed: make(chan Feed)}

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
	fm := NewFeedManager(db, conf, readeef.NewStandardLogger(os.Stderr, "", 0), &UpdateFeedReceiverManager{})
	fm.AddUpdateReceiver(rec)

	fm.Start()

	f := Feed{Link: ts.URL + "/link"}
	f, _, err = db.UpdateFeed(f)
	if err != nil {
		t.Fatal(err)
	}

	fm.AddFeed(f)

	<-done // First update request

	f2 := <-rec.updateFeed // Feed gets updated

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

	fm.RemoveFeed(f)

	time.Sleep(150 * time.Millisecond)
	select {
	case <-done:
		t.Fatalf("Did not expect another request\n")
	default:
	}
}

func TestFeedManagerDetection(t *testing.T) {
	var ts *httptest.Server

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/link" {
			w.WriteHeader(http.StatusOK)

			w.Write([]byte(strings.Replace(atomXml, "{{ .FeedLink }}", ts.URL+"/link", -1)))
		} else if r.RequestURI == "/html" {
			w.Write([]byte(`
<html>
	<head>
		<link type="text/css" href="/foo.css">
		<link rel="alternative" type="application/rss+xml" href="/link"/>
	</head>
	<body><main></main></body>
</html>
			`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	_, err := discoverParserFeeds(ts.URL)
	if err == nil {
		t.Fatalf("Expected an ErrNoFeed error, got nothing\n")
	} else if err != ErrNoFeed {
		t.Fatalf("Expected an ErrNoFeed error, got %v\n", err)
	}

	pf, err := discoverParserFeeds(ts.URL + "/link")
	if err != nil {
		t.Fatal(err)
	}

	expectedStr := ts.URL + "/link"
	if pf[0].Link != expectedStr {
		t.Fatalf("Expected '%s' for a url, got '%s'\n", expectedStr, pf[0].Link)
	}

	pf, err = discoverParserFeeds(ts.URL + "/html")
	if err != nil {
		t.Fatal(err)
	}

	expectedStr = ts.URL + "/link"
	if pf[0].Link != expectedStr {
		t.Fatalf("Expected '%s' for a url, got '%s'\n", expectedStr, pf[0].Link)
	}
}

type updateReceiver struct {
	updateFeed chan Feed
}

func (r updateReceiver) UpdateFeedChannel() chan<- Feed {
	return r.updateFeed
}
*/
