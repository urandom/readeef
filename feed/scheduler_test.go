package feed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/urandom/readeef/config"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

func TestScheduler_ScheduleFeed(t *testing.T) {
	type args struct {
		timeout time.Duration
		feed    content.Feed
		update  time.Duration
	}
	tests := []struct {
		name           string
		connectTimeout time.Duration
		args           args
		want           []int
	}{
		{"client timeout", time.Microsecond, args{time.Second, content.Feed{ID: 100, Link: "/feed"}, time.Second}, []int{-1}},
		{"timeout", time.Second, args{time.Nanosecond, content.Feed{ID: 100, Link: "/feed"}, time.Second}, []int{-1}},
		{"feed update", time.Second, args{2 * time.Second, content.Feed{ID: 100, Link: "/feed"}, time.Second}, []int{2, 1}},
		{"not-feed-content", time.Second, args{2 * time.Second, content.Feed{ID: 100, Link: "/not-feed"}, time.Second}, []int{-1}},
		{"404", time.Second, args{2 * time.Second, content.Feed{ID: 100, Link: "/404"}, time.Second}, []int{-1}},
		{"http error then update", time.Second, args{2 * time.Second, content.Feed{ID: 100, Link: "/error-update"}, time.Second}, []int{-1, 2}},
		{"same content", time.Second, args{2 * time.Second, content.Feed{ID: 100, Link: "/same-content"}, 100 * time.Millisecond}, []int{2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/feed":
					if iter == 0 {
						w.Write([]byte(rss2Xml))
					} else {
						w.Write([]byte(rss2Xmlv2))
					}
				case "/not-feed":
					w.Write([]byte("Hello world"))
				case "/404":
					w.WriteHeader(http.StatusNotFound)
				case "/error-update":
					if iter == 0 {
						w.WriteHeader(http.StatusServiceUnavailable)
					} else {
						w.Write([]byte(rss2Xml))
					}
				case "/same-content":
					if iter == 0 || iter == 1 {
						w.Write([]byte(rss2Xml))
					} else {
						w.Write([]byte(rss2Xmlv2))
					}
				}
				iter++
			}))
			defer ts.Close()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := config.Log{}
			cfg.Converted.Writer = os.Stderr
			s := Scheduler{
				ops:    make(chan feedOp),
				client: &http.Client{Timeout: tt.connectTimeout},
				log:    log.WithStd(cfg),
			}

			go s.Start(ctx)

			schedCtx := ctx

			if tt.args.timeout > 0 {
				var timeoutCancel context.CancelFunc
				schedCtx, timeoutCancel = context.WithTimeout(schedCtx, tt.args.timeout)
				defer timeoutCancel()
			}

			tt.args.feed.Link = ts.URL + tt.args.feed.Link

			up := s.ScheduleFeed(schedCtx, tt.args.feed, tt.args.update)

			if len(tt.want) == 0 {
				return
			}

			i := 0
			for {
				select {
				case data, ok := <-up:
					if tt.want[i] == -1 {
						if !ok {
							return
						}
						if !data.IsErr() {
							t.Errorf("Scheduler.ScheduleFeed() expected an error on update")
							return
						}
					} else {
						if data.IsErr() {
							t.Errorf("Scheduler.ScheduleFeed() unexpected error = %v", data.Error())
							return
						}

						if tt.want[i] != len(data.Feed.Articles) {
							t.Errorf("Scheduler.ScheduleFeed() len(data.Feed.Articles) = %d, want %d", len(data.Feed.Articles), tt.want[i])
							return
						}
					}
				case <-time.After(tt.args.update + 250*time.Millisecond):
					t.Errorf("Scheduler.ScheduleFeed() timeout waiting for data")
					return
				}
				i++

				if i == len(tt.want) {
					return
				}
			}
		})
	}
}

const (
	rss2Xml = `

<?xml version="1.0"?>
<rss version="2.0">
   <channel>
      <title>Liftoff News</title>
      <link>http://liftoff.msfc.nasa.gov/</link>
      <description>Liftoff to Space Exploration.</description>
      <language>en-us</language>
      <pubDate>Tue, 10 Jun 2003 04:00:00 GMT</pubDate>
      <lastBuildDate>Tue, 10 Jun 2003 09:41:01 GMT</lastBuildDate>
      <docs>http://blogs.law.harvard.edu/tech/rss</docs>
      <generator>Weblog Editor 2.0</generator>
      <managingEditor>editor@example.com</managingEditor>
      <webMaster>webmaster@example.com</webMaster>
	  <ttl>30</ttl>
	  <skipHours>
		  <hour>3</hour>
		  <hour>15</hour>
		  <hour>23</hour>
	  </skipHours>
	  <skipDays>
		  <day>Monday</day>
		  <day>Saturday</day>
	  </skipDays>
      <item>
         <title>Star City</title>
         <link>http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp</link>
         <description>How do Americans get ready to work with Russians aboard the International Space Station? They take a crash course in culture, language and protocol at Russia's &lt;a href="http://howe.iki.rssi.ru/GCTC/gctc_e.htm"&gt;Star City&lt;/a&gt;.</description>
         <pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
         <guid>http://liftoff.msfc.nasa.gov/2003/06/03.html#item573</guid>
      </item>
      <item>
         <description>Sky watchers in Europe, Asia, and parts of Alaska and Canada will experience a &lt;a href="http://science.nasa.gov/headlines/y2003/30may_solareclipse.htm"&gt;partial eclipse of the Sun&lt;/a&gt; on Saturday, May 31st.</description>
         <pubDate>Fri, 30 May 2003 11:06:42 GMT</pubDate>
         <guid>http://liftoff.msfc.nasa.gov/2003/05/30.html#item572</guid>
      </item>
   </channel>
</rss>
`
	rss2Xmlv2 = `

<?xml version="1.0"?>
<rss version="2.0">
   <channel>
      <title>Liftoff News</title>
      <link>http://liftoff.msfc.nasa.gov/</link>
      <description>Liftoff to Space Exploration.</description>
      <language>en-us</language>
      <pubDate>Tue, 10 Jun 2003 05:00:00 GMT</pubDate>
      <lastBuildDate>Tue, 10 Jun 2003 09:41:01 GMT</lastBuildDate>
      <docs>http://blogs.law.harvard.edu/tech/rss</docs>
      <generator>Weblog Editor 2.0</generator>
      <managingEditor>editor@example.com</managingEditor>
      <webMaster>webmaster@example.com</webMaster>
	  <ttl>30</ttl>
	  <skipHours>
		  <hour>3</hour>
		  <hour>15</hour>
		  <hour>23</hour>
	  </skipHours>
	  <skipDays>
		  <day>Monday</day>
		  <day>Saturday</day>
	  </skipDays>
      <item>
         <title>Sky watchers</title>
         <description>Sky watchers in Europe</description>
         <pubDate>Fri, 30 May 2003 15:06:42 GMT</pubDate>
         <guid>http://liftoff.msfc.nasa.gov/2003/05/30.html#item624</guid>
      </item>
   </channel>
</rss>
`
)
