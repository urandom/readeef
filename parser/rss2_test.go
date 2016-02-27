package parser

import (
	"strings"
	"testing"
	"time"
)

var rss2Xml = `

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

func TestRss2Parse(t *testing.T) {
	f, err := ParseRss2([]byte(""))
	if err == nil {
		t.Fatalf("Expected an error, got none, feed: '%v'\n", f)
	}

	f, err = ParseRss2([]byte(rss2Xml))
	if err != nil {
		t.Fatal(err)
	}

	if f.Title != "Liftoff News" {
		t.Fatalf("Unexpected feed title: '%s'\n", f.Title)
	}

	if strings.TrimSpace(f.Description) != "Liftoff to Space Exploration." {
		t.Fatalf("Unexpected feed description: '%s'\n", strings.TrimSpace(f.Description))
	}

	if f.SiteLink != "http://liftoff.msfc.nasa.gov/" {
		t.Fatalf("Unexpected feed link: '%s'\n", f.SiteLink)
	}

	if f.HubLink != "" {
		t.Fatalf("Unexpected feed hub link: '%s'\n", f.HubLink)
	}

	if f.Image != (Image{}) {
		t.Fatalf("Unexpected feed image: '%v'\n", f.Image)
	}

	if 2 != len(f.Articles) {
		t.Fatalf("Unexpected number of feed articles: '%d'\n", len(f.Articles))
	}

	a := f.Articles[0]
	expectedStr := "http://liftoff.msfc.nasa.gov/2003/06/03.html#item573"
	if a.Guid != expectedStr {
		t.Fatalf("Expected %s as id, got '%s'\n", expectedStr, a.Guid)
	}

	if a.Title != "Star City" {
		t.Fatalf("Unexpected article title: '%v'\n", a.Title)
	}

	if !strings.Contains(a.Description, `How do Americans get ready to work with Russians aboard the International Space Station?`) {
		t.Fatalf("Unexpected article description: '%v'\n", a.Description)
	}

	if a.Link != "http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp" {
		t.Fatalf("Unexpected article link: '%v'\n", a.Link)
	}

	d, _ := parseDate("Tue, 03 Jun 2003 09:39:21 GMT")

	if !a.Date.Equal(d) {
		t.Fatalf("Unexpected article date: '%s'\n", a.Date)
	}

	expectedStr = "http://liftoff.msfc.nasa.gov/2003/05/30.html#item572"
	if f.Articles[1].Guid != expectedStr {
		t.Fatalf("Expected %s as id, got '%s'\n", expectedStr, f.Articles[1].Guid)
	}

	expectedDuration := 30 * time.Minute
	if f.TTL != expectedDuration {
		t.Fatalf("Expected ttl '%d', got: %d\n", expectedDuration, f.TTL)
	}

	expectedInt := 3
	if len(f.SkipHours) != expectedInt {
		t.Fatalf("Expected '%d' number of skipHours, got '%d'\n", expectedInt, len(f.SkipHours))
	}

	for _, expectedInt = range []int{3, 15, 23} {
		if !f.SkipHours[expectedInt] {
			t.Fatalf("Expected '%d' skipHour\n", expectedInt)
		}
	}

	expectedInt = 2
	if len(f.SkipDays) != expectedInt {
		t.Fatalf("Expected '%d' number of skipDays, got '%d'\n", expectedInt, len(f.SkipDays))
	}

	for _, expectedStr = range []string{"Monday", "Saturday"} {
		if !f.SkipDays[expectedStr] {
			t.Fatalf("Expected '%s' skipDay\n", expectedStr)
		}
	}
}
