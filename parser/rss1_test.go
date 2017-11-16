package parser

import (
	"reflect"
	"testing"
	"time"
)

func TestParseRss1(t *testing.T) {
	tests := []struct {
		name    string
		b       []byte
		want    Feed
		wantErr bool
	}{
		{"single", []byte(singleRss1XML), singleRss1Feed, false},
		{"single date", []byte(singleDateRss1XML), singleRss1Feed, false},
		{"single no date", []byte(singleNoDateRss1XML), singleNoDateRss1Feed, false},
		{"multi last no date", []byte(multiLastNoDateRss1XML), multiLastNoDateRss1Feed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRss1(tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRss1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRss1() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	singleRss1Feed = Feed{
		Title:    "XML.com",
		SiteLink: "http://xml.com/pub",
		Description: `
      XML.com features a rich mix of information and services 
      for the XML community.
    `,
		Image: Image{
			Title: "XML.com",
			Url:   "http://xml.com/universal/images/xml_tiny.gif",
		},
		TTL:       30 * time.Minute,
		SkipHours: map[int]bool{3: true, 15: true, 23: true},
		SkipDays:  map[string]bool{"Monday": true, "Saturday": true},
		Articles: []Article{
			{
				Title: "Title 1",
				Link:  "http://xml.com/link1.html",
				Guid:  "http://xml.com/id1.html",
				Description: `
	Descr 1
	`,
				Date: time.Date(2003, time.June, 3, 9, 39, 21, 0, gmt),
			},
		},
	}

	singleNoDateRss1Feed = Feed{
		Title:    "XML.com",
		SiteLink: "http://xml.com/pub",
		Description: `
      XML.com features a rich mix of information and services 
      for the XML community.
    `,
		Image: Image{
			Title: "XML.com",
			Url:   "http://xml.com/universal/images/xml_tiny.gif",
		},
		TTL:       30 * time.Minute,
		SkipHours: map[int]bool{3: true, 15: true, 23: true},
		SkipDays:  map[string]bool{"Monday": true, "Saturday": true},
		Articles: []Article{
			{
				Title: "Title 1",
				Link:  "http://xml.com/link1.html",
				Guid:  "http://xml.com/id1.html",
				Description: `
	Descr 1
	`,
				Date: time.Unix(0, 0),
			},
		},
	}

	multiLastNoDateRss1Feed = Feed{
		Title:    "XML.com",
		SiteLink: "http://xml.com/pub",
		Description: `
      XML.com features a rich mix of information and services 
      for the XML community.
    `,
		Image: Image{
			Title: "XML.com",
			Url:   "http://xml.com/universal/images/xml_tiny.gif",
		},
		TTL:       30 * time.Minute,
		SkipHours: map[int]bool{3: true, 15: true, 23: true},
		SkipDays:  map[string]bool{"Monday": true, "Saturday": true},
		Articles: []Article{
			{
				Title: "Title 1",
				Link:  "http://xml.com/link1.html",
				Guid:  "http://xml.com/id1.html",
				Description: `
	Descr 1
	`,
				Date: time.Date(2003, time.June, 3, 9, 39, 21, 0, gmt),
			},
			{
				Title: "Title 2",
				Link:  "http://xml.com/link2.html",
				Guid:  "http://xml.com/id2.html",
				Description: `
	Descr 2
	`,
				Date: time.Date(2003, time.June, 3, 9, 39, 22, 0, gmt),
			},
		},
	}
)

const (
	singleRss1XML = `
<?xml version="1.0"?>
<rdf:RDF 
  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns="http://purl.org/rss/1.0/"
>

  <channel rdf:about="http://www.xml.com/xml/news.rss">
    <title>XML.com</title>
    <link>http://xml.com/pub</link>
    <description>
      XML.com features a rich mix of information and services 
      for the XML community.
    </description>
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

    <image rdf:resource="http://xml.com/universal/images/xml_tiny.gif" />

    <items>
      <rdf:Seq>
        <rdf:li resource="http://xml.com/pub/2000/08/09/xslt/xslt.html" />
        <rdf:li resource="http://xml.com/pub/2000/08/09/rdfdb/index.html" />
      </rdf:Seq>
    </items>

    <textinput rdf:resource="http://search.xml.com" />

  </channel>
  
  <image rdf:about="http://xml.com/universal/images/xml_tiny.gif">
    <title>XML.com</title>
    <link>http://www.xml.com</link>
    <url>http://xml.com/universal/images/xml_tiny.gif</url>
  </image>
  
  <item rdf:about="http://xml.com/id1.html">
    <title>Title 1</title>
    <link>http://xml.com/link1.html</link>
    <description>
	Descr 1
	</description>
    <pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
  </item>
  
  <textinput rdf:about="http://search.xml.com">
    <title>Search XML.com</title>
    <description>Search XML.com's XML collection</description>
    <name>s</name>
    <link>http://search.xml.com</link>
  </textinput>

</rdf:RDF>
`
	singleDateRss1XML = `
<?xml version="1.0"?>
<rdf:RDF 
  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns="http://purl.org/rss/1.0/"
>

  <channel rdf:about="http://www.xml.com/xml/news.rss">
    <title>XML.com</title>
    <link>http://xml.com/pub</link>
    <description>
      XML.com features a rich mix of information and services 
      for the XML community.
    </description>
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

	<image rdf:about="http://xml.com/universal/images/xml_tiny.gif">
	  <title>XML.com</title>
	  <link>http://www.xml.com</link>
	  <url>http://xml.com/universal/images/xml_tiny.gif</url>
	</image>
  
    <items>
      <rdf:Seq>
        <rdf:li resource="http://xml.com/pub/2000/08/09/xslt/xslt.html" />
        <rdf:li resource="http://xml.com/pub/2000/08/09/rdfdb/index.html" />
      </rdf:Seq>
    </items>

    <textinput rdf:resource="http://search.xml.com" />

  </channel>
  
  <item rdf:about="http://xml.com/id1.html">
    <title>Title 1</title>
    <link>http://xml.com/link1.html</link>
    <description>
	Descr 1
	</description>
    <date>Tue, 03 Jun 2003 09:39:21 GMT</date>
  </item>
</rdf:RDF>
`

	singleNoDateRss1XML = `
<?xml version="1.0"?>
<rdf:RDF 
  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns="http://purl.org/rss/1.0/"
>

  <channel rdf:about="http://www.xml.com/xml/news.rss">
    <title>XML.com</title>
    <link>http://xml.com/pub</link>
    <description>
      XML.com features a rich mix of information and services 
      for the XML community.
    </description>
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

	<image rdf:about="http://xml.com/universal/images/xml_tiny.gif">
	  <title>XML.com</title>
	  <link>http://www.xml.com</link>
	  <url>http://xml.com/universal/images/xml_tiny.gif</url>
	</image>
  
    <items>
      <rdf:Seq>
        <rdf:li resource="http://xml.com/pub/2000/08/09/xslt/xslt.html" />
        <rdf:li resource="http://xml.com/pub/2000/08/09/rdfdb/index.html" />
      </rdf:Seq>
    </items>

    <textinput rdf:resource="http://search.xml.com" />

  </channel>
  
  <item rdf:about="http://xml.com/id1.html">
    <title>Title 1</title>
    <link>http://xml.com/link1.html</link>
    <description>
	Descr 1
	</description>
  </item>
  
  <textinput rdf:about="http://search.xml.com">
    <title>Search XML.com</title>
    <description>Search XML.com's XML collection</description>
    <name>s</name>
    <link>http://search.xml.com</link>
  </textinput>

</rdf:RDF>
`

	multiLastNoDateRss1XML = `
<?xml version="1.0"?>
<rdf:RDF 
  xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
  xmlns="http://purl.org/rss/1.0/"
>

  <channel rdf:about="http://www.xml.com/xml/news.rss">
    <title>XML.com</title>
    <link>http://xml.com/pub</link>
    <description>
      XML.com features a rich mix of information and services 
      for the XML community.
    </description>
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

	<image rdf:about="http://xml.com/universal/images/xml_tiny.gif">
	  <title>XML.com</title>
	  <link>http://www.xml.com</link>
	  <url>http://xml.com/universal/images/xml_tiny.gif</url>
	</image>
  
    <items>
      <rdf:Seq>
        <rdf:li resource="http://xml.com/pub/2000/08/09/xslt/xslt.html" />
        <rdf:li resource="http://xml.com/pub/2000/08/09/rdfdb/index.html" />
      </rdf:Seq>
    </items>

    <textinput rdf:resource="http://search.xml.com" />

  </channel>
  
  <item rdf:about="http://xml.com/id1.html">
    <title>Title 1</title>
    <link>http://xml.com/link1.html</link>
    <description>
	Descr 1
	</description>
    <date>Tue, 03 Jun 2003 09:39:21 GMT</date>
  </item>

  <item rdf:about="http://xml.com/id2.html">
    <title>Title 2</title>
    <link>http://xml.com/link2.html</link>
    <description>
	Descr 2
	</description>
  </item>

</rdf:RDF>
`
)
