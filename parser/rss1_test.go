package parser

import (
	"strings"
	"testing"
	"time"
)

var rss1Xml = `
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
  
  <item rdf:about="http://xml.com/pub/2000/08/09/xslt/xslt.html">
    <title>Processing Inclusions with XSLT</title>
    <link>http://xml.com/pub/2000/08/09/xslt/xslt.html</link>
    <description>
     Processing document inclusions with general XML tools can be 
     problematic. This article proposes a way of preserving inclusion 
     information through SAX-based processing.
    </description>
  </item>
  
  <item rdf:about="http://xml.com/pub/2000/08/09/rdfdb/index.html">
    <title>Putting RDF to Work</title>
    <link>http://xml.com/pub/2000/08/09/rdfdb/index.html</link>
    <description>
     Tool and API support for the Resource Description Framework 
     is slowly coming of age. Edd Dumbill takes a look at RDFDB, 
     one of the most exciting new RDF toolkits.
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

func TestRss1Parse(t *testing.T) {
	f, err := ParseRss1([]byte(""))
	if err == nil {
		t.Fatalf("Expected an error, got none, feed: '%v'\n", f)
	}

	f, err = ParseRss1([]byte(rss1Xml))
	if err != nil {
		t.Fatal(err)
	}

	if f.Title != "XML.com" {
		t.Fatalf("Unexpected feed title: '%s'\n", f.Title)
	}

	descr := `XML.com features a rich mix of information and services 
      for the XML community.`
	if strings.TrimSpace(f.Description) != descr {
		t.Fatalf("Unexpected feed description: '%s'\n", strings.TrimSpace(f.Description))
	}

	if f.SiteLink != "http://xml.com/pub" {
		t.Fatalf("Unexpected feed link: '%s'\n", f.SiteLink)
	}

	if f.HubLink != "" {
		t.Fatalf("Unexpected feed hub link: '%s'\n", f.HubLink)
	}

	if f.Image != (Image{"XML.com", "http://xml.com/universal/images/xml_tiny.gif", 0, 0}) {
		t.Fatalf("Unexpected feed image: '%v'\n", f.Image)
	}

	if 2 != len(f.Articles) {
		t.Fatalf("Unexpected number of feed articles: '%d'\n", len(f.Articles))
	}

	a := f.Articles[0]
	expectedStr := "http://xml.com/pub/2000/08/09/xslt/xslt.html"
	if a.Guid != expectedStr {
		t.Fatalf("Expected %s as id, got '%s'\n", expectedStr, a.Guid)
	}

	if a.Title != "Processing Inclusions with XSLT" {
		t.Fatalf("Unexpected article title: '%v'\n", a.Title)
	}

	if !strings.Contains(a.Description, `Processing document inclusions with general XML tools can be `) {
		t.Fatalf("Unexpected article description: '%v'\n", a.Description)
	}

	if a.Link != "http://xml.com/pub/2000/08/09/xslt/xslt.html" {
		t.Fatalf("Unexpected article link: '%v'\n", a.Link)
	}

	if !a.Date.Equal(unknownTime) {
		t.Fatalf("Unexpected article date: '%s'\n", a.Date)
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

	for _, expectedStr := range []string{"Monday", "Saturday"} {
		if !f.SkipDays[expectedStr] {
			t.Fatalf("Expected '%s' skipDay\n", expectedStr)
		}
	}
}
