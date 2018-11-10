package parser

import (
	"testing"
	"time"

	"github.com/go-test/deep"
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
		{"html escapes in xml", []byte(htmlEscapesInXML), htmlEscapesInXMLFeed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRss2(tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRss2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("ParseRss2() = %v", diff)
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

	htmlEscapesInXMLFeed = Feed{
		Title:       "xkcd.com",
		SiteLink:    "https://xkcd.com/",
		Description: "xkcd.com: A webcomic of romance and math humor.",
		SkipHours:   map[int]bool{},
		SkipDays:    map[string]bool{},
		Articles: []Article{
			{
				Title:       "Election Night",
				Link:        "https://xkcd.com/2068/",
				Guid:        "https://xkcd.com/2068/",
				Description: `<img src="https://imgs.xkcd.com/comics/election_night.png" title="&quot;Even the blind—those who are anxious to hear, but are not able to see—will be taken care of. Immense megaphones have been constructed and will be in use at The Tribune office and in the Coliseum. The one at the Coliseum will be operated by a gentleman who draws $60 a week from Barnum & Bailey's circus for the use of his voice.&quot;" alt="&quot;Even the blind—those who are anxious to hear, but are not able to see—will be taken care of. Immense megaphones have been constructed and will be in use at The Tribune office and in the Coliseum. The one at the Coliseum will be operated by a gentleman who draws $60 a week from Barnum & Bailey's circus for the use of his voice.&quot;" />`,
				Date:        time.Date(2018, time.November, 5, 5, 0, 0, 0, gmt),
			},
			{
				Title:       "Challengers",
				Link:        "https://xkcd.com/2067/",
				Guid:        "https://xkcd.com/2067/",
				Description: `<img src="https://imgs.xkcd.com/comics/challengers.png" title="Use your mouse or fingers to pan + zoom. To edit the map, submit your ballot on November 6th." alt="Use your mouse or fingers to pan + zoom. To edit the map, submit your ballot on November 6th." />`,
				Date:        time.Date(2018, time.November, 2, 4, 0, 0, 0, gmt),
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
	htmlEscapesInXML = `
<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
    <channel>
        <title>xkcd.com</title>
        <link>https://xkcd.com/</link>
        <description>xkcd.com: A webcomic of romance and math humor.</description>
        <language>en</language>
        <item>
            <title>Election Night</title>
            <link>https://xkcd.com/2068/</link>
            <description>&lt;img src="https://imgs.xkcd.com/comics/election_night.png" title="&amp;quot;Even the blind&#x2014;those who are anxious to hear, but are not able to see&#x2014;will be taken care of. Immense megaphones have been constructed and will be in use at The Tribune office and in the Coliseum. The one at the Coliseum will be operated by a gentleman who draws $60 a week from Barnum &amp; Bailey's circus for the use of his voice.&amp;quot;" alt="&amp;quot;Even the blind&#x2014;those who are anxious to hear, but are not able to see&#x2014;will be taken care of. Immense megaphones have been constructed and will be in use at The Tribune office and in the Coliseum. The one at the Coliseum will be operated by a gentleman who draws $60 a week from Barnum &amp; Bailey's circus for the use of his voice.&amp;quot;" /&gt;</description>
            <pubDate>Mon, 05 Nov 2018 05:00:00 -0000</pubDate>
            <guid>https://xkcd.com/2068/</guid>
        </item>
        <item>
            <title>Challengers</title>
            <link>https://xkcd.com/2067/</link>
            <description>&lt;img src="https://imgs.xkcd.com/comics/challengers.png" title="Use your mouse or fingers to pan + zoom. To edit the map, submit your ballot on November 6th." alt="Use your mouse or fingers to pan + zoom. To edit the map, submit your ballot on November 6th." /&gt;</description>
            <pubDate>Fri, 02 Nov 2018 04:00:00 -0000</pubDate>
            <guid>https://xkcd.com/2067/</guid>
        </item>
    </channel>
</rss>


`
)
