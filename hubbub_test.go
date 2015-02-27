package readeef

/*

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

func TestHubbub(t *testing.T) {
	cleanDB(t)

	var ts *httptest.Server
	var callbackURL string

	done := make(chan bool)

	webfwConfig, err := webfw.ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := webfw.NewDispatcher("/", webfwConfig)
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/callback/v0/") {
			r.RequestURI = strings.Replace(r.RequestURI, "/callback/v0", "", -1)
		}
		defer func() { done <- true }()
		dispatcher.ServeHTTP(w, r)
	}))
	defer ts.Close()

	var conf Config
	addFeed := make(chan Feed)
	removeFeed := make(chan Feed)

	go func() {
		for {
			select {
			case <-addFeed:
			case <-removeFeed:
			}
		}
	}()

	h := NewHubbub(db, conf, webfw.NewStandardLogger(os.Stderr, "", 0), "/callback/", addFeed, removeFeed, &UpdateFeedReceiverManager{})
	f := Feed{Link: ts.URL + "/link"}
	f, _, err = db.UpdateFeed(f)
	if err != nil {
		t.Fatal(err)
	}

	err = h.Subscribe(f)
	if err != ErrNotConfigured {
		t.Fatalf("Exepcted a ErrNotConfigured error, got: '%v'\n", err)
	}

	callbackURL = fmt.Sprintf("%s/callback/v0/hubbub/%d", ts.URL, f.Id)

	conf.Hubbub.CallbackURL = ts.URL
	conf.Hubbub.RelativePath = "/hubbub"
	conf.Timeout.Converted.Connect = time.Second
	conf.Timeout.Converted.ReadWrite = time.Second
	h = NewHubbub(db, conf, webfw.NewStandardLogger(os.Stderr, "", 0), "/callback/", addFeed, removeFeed, &UpdateFeedReceiverManager{})

	dispatcher.Handle(hublink_con{webfw.NewBasePatternController("/hublink", webfw.MethodAll, ""), callbackURL, ts.URL + "/link", done})

	hc := NewHubbubController(h)
	dispatcher.Handle(hc)
	dispatcher.Initialize()

	err = h.Subscribe(f)
	if err != ErrNoFeedHubLink {
		t.Fatalf("Exepcted a ErrNoFeedHubLink error, got: '%v'\n", err)
	}

	f.HubLink = ts.URL + "/hublink"

	f, _, err = db.UpdateFeed(f)
	if err != nil {
		t.Fatalf("Got an error during feed db update: '%v'\n", err)
	}

	err = h.Subscribe(f)
	if err != nil {
		t.Fatalf("Got an error during subscription: '%v'\n", err)
	}

	<-done // hublink request
	<-done // callback request
	if s, err := db.GetHubbubSubscription(f.Id); err != nil {
		t.Fatal(err)
	} else {
		if s.SubscriptionFailure {
			t.Fatal("Expected successfull subscription\n")
		}
	}

	<-done // subscription challenge

	go func() {
		buf := bytes.NewBufferString(strings.Replace(atomXml, "{{ .FeedLink }}", f.Link, -1))
		resp, err := http.Post(callbackURL, "text/xml", buf)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}()

	<-done // subscription feed update

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

	subs, err := db.GetHubbubSubscriptions()
	if err != nil {
		t.Fatal(err)
	}

	expectedInt := 1
	if len(subs) != expectedInt {
		t.Fatalf("Expected %d subscriptions, got %d\n", expectedInt, len(subs))
	}
}

type hublink_con struct {
	webfw.BasePatternController
	callbackURL string
	feedLink    string
	done        chan bool
}

func (con hublink_con) Handler(c context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			panic(err)
		}

		expectedStr := con.callbackURL
		if r.Form.Get("hub.callback") != expectedStr {
			panic(fmt.Sprintf("Expected hub.callback '%s', got '%s'\n", r.Form.Get("hub.callback"), expectedStr))
		}

		expectedStr = con.feedLink
		if r.Form.Get("hub.topic") != expectedStr {
			panic(fmt.Sprintf("Expected hub.topic '%s', got '%s'\n", r.Form.Get("hub.topic"), expectedStr))
		}

		expectedStr = "subscribe"
		if r.Form.Get("hub.mode") != expectedStr {
			panic(fmt.Sprintf("Expected hub.mode '%s', got '%s'\n", r.Form.Get("hub.mode"), expectedStr))
		}
		w.WriteHeader(http.StatusAccepted)

		go func() {
			res, err := http.Get(con.callbackURL + "?hub.mode=subscribe&hub.challenge=secret")
			defer res.Body.Close()
			defer func() { con.done <- true }()
			if err != nil {
				panic(err)
			}

			challenge, err := ioutil.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			challengeStr := string(challenge[:])
			expectedStr = "secret"
			if challengeStr != expectedStr {
				panic(fmt.Sprintf("Expected challenge '%s', got '%s'\n", challengeStr, expectedStr))
			}
		}()
	})
}

var atomXml = `` +
	`<feed xmlns="http://www.w3.org/2005/Atom" updated="2003-12-13T18:30:02Z">` +
	`<title>Example Feed</title>` +
	`<id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>` +
	`<link href="{{ .FeedLink }}"></link>` +
	`<author><name>John Doe</name><uri></uri><email></email></author>` +
	`<entry>` +
	`<title>Atom-Powered Robots Run Amok</title>` +
	`<id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>` +
	`<link href="http://example.org/2003/12/13/atom03"></link>` +
	`<updated>2003-12-13T18:30:02Z</updated>` +
	`<author><name></name><uri></uri><email></email></author>` +
	`<summary>Some text.</summary>` +
	`</entry>` +
	`</feed>`
*/
