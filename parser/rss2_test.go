package parser

import (
	"reflect"
	"testing"
	"time"
)

func TestParseRss2(t *testing.T) {
	tests := []struct {
		name    string
		b       []byte
		want    Feed
		wantErr bool
	}{
		{"single", []byte(singleRss2XML), singleRss2Feed, false},
		{"single date", []byte(singleDateRss2XML), singleRss2Feed, false},
		{"single no date", []byte(singleNoDateRss2XML), singleNoDateRss2Feed, false},
		{"multi last no date", []byte(multiLastNoDateRss2XML), multiLastNoDateRss2Feed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRss2(tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRss2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRss2() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	gmt, _ = time.LoadLocation("GMT")

	singleRss2Feed = Feed{
		Title:       "Liftoff News",
		SiteLink:    "http://liftoff.msfc.nasa.gov/",
		Description: "Liftoff to Space Exploration.",
		TTL:         30 * time.Minute,
		SkipHours:   map[int]bool{3: true, 15: true, 23: true},
		SkipDays:    map[string]bool{"Monday": true, "Saturday": true},
		Articles: []Article{
			{
				Title:       "Star City",
				Link:        "http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp",
				Guid:        "http://liftoff.msfc.nasa.gov/2003/06/03.html#item573",
				Description: `How do Americans get ready to work with Russians aboard the International Space Station? They take a crash course in culture, language and protocol at Russia's <a href="http://howe.iki.rssi.ru/GCTC/gctc_e.htm">Star City</a>.`,
				Date:        time.Date(2003, time.June, 3, 9, 39, 21, 0, gmt),
			},
		},
	}

	singleNoDateRss2Feed = Feed{
		Title:       "Liftoff News",
		SiteLink:    "http://liftoff.msfc.nasa.gov/",
		Description: "Liftoff to Space Exploration.",
		TTL:         30 * time.Minute,
		SkipHours:   map[int]bool{3: true, 15: true, 23: true},
		SkipDays:    map[string]bool{"Monday": true, "Saturday": true},
		Articles: []Article{
			{
				Title:       "Star City",
				Link:        "http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp",
				Guid:        "http://liftoff.msfc.nasa.gov/2003/06/03.html#item573",
				Description: `How do Americans get ready to work with Russians aboard the International Space Station? They take a crash course in culture, language and protocol at Russia's <a href="http://howe.iki.rssi.ru/GCTC/gctc_e.htm">Star City</a>.`,
				Date:        time.Unix(0, 0),
			},
		},
	}

	multiLastNoDateRss2Feed = Feed{
		Title:       "Liftoff News",
		SiteLink:    "http://liftoff.msfc.nasa.gov/",
		Description: "Liftoff to Space Exploration.",
		TTL:         30 * time.Minute,
		SkipHours:   map[int]bool{3: true, 15: true, 23: true},
		SkipDays:    map[string]bool{"Monday": true, "Saturday": true},
		Articles: []Article{
			{
				Title:       "Star City",
				Link:        "http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp",
				Guid:        "http://liftoff.msfc.nasa.gov/2003/06/03.html#item573",
				Description: `Descr 1`,
				Date:        time.Date(2003, time.June, 3, 9, 39, 21, 0, gmt),
			},
			{
				Title:       "",
				Link:        "",
				Guid:        "http://liftoff.msfc.nasa.gov/2003/05/30.html#item572",
				Description: `Descr 2`,
				Date:        time.Date(2003, time.June, 3, 9, 39, 22, 0, gmt),
			},
		},
	}
)

const (
	singleRss2XML = `

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
   </channel>
</rss>
`

	singleDateRss2XML = `

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
         <date>Tue, 03 Jun 2003 09:39:21 GMT</date>
         <guid>http://liftoff.msfc.nasa.gov/2003/06/03.html#item573</guid>
      </item>
   </channel>
</rss>
`
	singleNoDateRss2XML = `

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
         <guid>http://liftoff.msfc.nasa.gov/2003/06/03.html#item573</guid>
      </item>
   </channel>
</rss>
`
	multiLastNoDateRss2XML = `

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
         <description>Descr 1</description>
         <pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
         <guid>http://liftoff.msfc.nasa.gov/2003/06/03.html#item573</guid>
      </item>
      <item>
         <description>Descr 2</description>
         <guid>http://liftoff.msfc.nasa.gov/2003/05/30.html#item572</guid>
      </item>
   </channel>
</rss>
`
)
