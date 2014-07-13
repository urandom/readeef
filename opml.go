package readeef

import "encoding/xml"

type Opml struct {
	Feeds []OpmlFeed
}

type OpmlFeed struct {
	Title string
	Url   string
}

type opmlXml struct {
	Version string   `xml:"version,attr"`
	Head    opmlHead `xml:"head"`
	Body    opmlBody `xml:"body"`
}

type opmlHead struct {
	Title string `xml:"title"`
}

type opmlBody struct {
	Outline []opmlOutline `xml:"outline"`
}

type opmlOutline struct {
	Type    string `xml:"type,attr,omitempty"`
	Text    string `xml:"text,attr,omitempty"`
	Title   string `xml:"title,attr,omitempty"`
	XmlUrl  string `xml:"xmlUrl,attr,omitempty"`
	HtmlUrl string `xml:"htmlUrl,attr,omitempty"`
	Url     string `xml:"url,attr,omitempty"`
}

func ParseOpml(content []byte) Opml {
	var o opmlXml
	if err := xml.Unmarshal(content, &o); err != nil {
		logger.Fatal(err)
	}

	opml := Opml{}
	for _, outline := range o.Body.Outline {
		feed := OpmlFeed{Title: outline.Text}
		if outline.Url == "" {
			feed.Url = outline.XmlUrl
		} else {
			feed.Url = outline.Url
		}
		opml.Feeds = append(opml.Feeds, feed)
	}

	return opml
}
