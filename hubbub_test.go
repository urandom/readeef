package readeef

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"readeef/parser"
	"strings"
	"testing"
	"time"

	"github.com/urandom/webfw"
	"github.com/urandom/webfw/context"
)

func TestHubbub(t *testing.T) {
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

	callbackURL = fmt.Sprintf("%s/callback/v0/hubbub/%s", ts.URL, url.QueryEscape(strings.Replace(ts.URL+"/link", "/", "|", -1)))

	db := NewDB("sqlite3", conn)
	if err := db.Connect(); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var conf Config
	addFeed := make(chan Feed)
	removeFeed := make(chan Feed)
	updateFeed := make(chan Feed)

	go func() {
		for {
			select {
			case <-addFeed:
			case <-removeFeed:
			case <-updateFeed:
			}
		}
	}()

	h := NewHubbub(db, conf, log.New(os.Stderr, "", 0), addFeed, removeFeed, updateFeed)
	f := Feed{Feed: parser.Feed{Link: ts.URL + "/link"}}

	err = h.Subscribe(f)
	if err != ErrNotConfigured {
		t.Fatalf("Exepcted a ErrNotConfigured error, got: '%v'\n", err)
	}

	conf.Hubbub.CallbackURL = ts.URL + "/callback"
	conf.Hubbub.RelativePath = "/hubbub"
	conf.Timeout.Converted.Connect = time.Second
	conf.Timeout.Converted.ReadWrite = time.Second
	h = NewHubbub(db, conf, log.New(os.Stderr, "", 0), addFeed, removeFeed, updateFeed)

	dispatcher.Handle(hublink_con{webfw.NewBaseController("/hublink", webfw.MethodAll, ""), callbackURL, ts.URL + "/link", t, done})

	hc := NewHubbubController(h)
	dispatcher.Handle(hc)
	dispatcher.Initialize()

	err = h.Subscribe(f)
	if err != ErrNoFeedHubLink {
		t.Fatalf("Exepcted a ErrNoFeedHubLink error, got: '%v'\n", err)
	}

	f = Feed{Feed: parser.Feed{Link: ts.URL + "/link"}, HubLink: ts.URL + "/hublink"}

	err = db.UpdateFeed(f)
	if err != nil {
		t.Fatalf("Got an error during feed db update: '%v'\n", err)
	}

	err = h.Subscribe(f)
	if err != nil {
		t.Fatalf("Got an error during subscription: '%v'\n", err)
	}

	<-done // hublink request
	<-done // callback request
	if s, err := db.GetHubbubSubscriptionByFeed(f.Link); err != nil {
		t.Fatal(err)
	} else {
		if s.SubscriptionFailure {
			t.Fatal("Expected successfull subscription\n")
		}
	}

	<-done // subscription challenge

	go func() {
		buf := bytes.NewBufferString(strings.Replace(atomXml, "{{ .FeedLink }}", f.Link, -1))
		_, err = http.Post(callbackURL, "text/xml", buf)
		if err != nil {
			t.Fatal(err)
		}
	}()

	<-done // subscription feed update

	f, err = db.GetFeed(f.Link)
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
}

type hublink_con struct {
	webfw.BaseController
	callbackURL string
	feedLink    string
	t           *testing.T
	done        chan bool
}

func (con hublink_con) Handler(c context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			con.t.Fatal(err)
		}

		expectedStr := con.callbackURL
		if r.Form.Get("hub.callback") != expectedStr {
			con.t.Fatalf("Expected hub.callback '%s', got '%s'\n", r.Form.Get("hub.callback"), expectedStr)
		}

		expectedStr = con.feedLink
		if r.Form.Get("hub.topic") != expectedStr {
			con.t.Fatalf("Expected hub.topic '%s', got '%s'\n", r.Form.Get("hub.topic"), expectedStr)
		}

		expectedStr = "subscribe"
		if r.Form.Get("hub.mode") != expectedStr {
			con.t.Fatalf("Expected hub.mode '%s', got '%s'\n", r.Form.Get("hub.mode"), expectedStr)
		}
		w.WriteHeader(http.StatusAccepted)

		go func() {
			res, err := http.Get(con.callbackURL + "?hub.mode=subscribe&hub.challenge=secret")
			defer func() { con.done <- true }()
			if err != nil {
				con.t.Fatal(err)
			}

			challenge, err := ioutil.ReadAll(res.Body)
			if err != nil {
				con.t.Fatal(err)
			}

			challengeStr := string(challenge[:])
			expectedStr = "secret"
			if challengeStr != expectedStr {
				con.t.Fatalf("Expected challenge '%s', got '%s'\n", challengeStr, expectedStr)
			}
		}()
	}
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
