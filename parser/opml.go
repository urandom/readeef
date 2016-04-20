package parser

import "encoding/xml"

type Opml struct {
	Feeds []OpmlFeed
}

type OpmlFeed struct {
	Title string
	Url   string
	Tags  []string
}

type OpmlXml struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    OpmlHead `xml:"head"`
	Body    OpmlBody `xml:"body"`
}

type OpmlHead struct {
	Title string `xml:"title"`
}

type OpmlBody struct {
	Outline []OpmlOutline `xml:"outline"`
}

type OpmlOutline struct {
	Type     string        `xml:"type,attr,omitempty"`
	Text     string        `xml:"text,attr,omitempty"`
	Title    string        `xml:"title,attr,omitempty"`
	XmlUrl   string        `xml:"xmlUrl,attr,omitempty"`
	HtmlUrl  string        `xml:"htmlUrl,attr,omitempty"`
	Url      string        `xml:"url,attr,omitempty"`
	Category string        `xml:"category,attr,omitempty"`
	Outline  []OpmlOutline `xml:"outline"`
}

func ParseOpml(content []byte) (Opml, error) {
	var o OpmlXml
	opml := Opml{}

	if err := xml.Unmarshal(content, &o); err != nil {
		return opml, err
	}

	processOutline(&opml, o.Body.Outline, "")

	return opml, nil
}

func processOutline(opml *Opml, outlines []OpmlOutline, tag string) {
	for _, outline := range outlines {
		if len(outline.Outline) == 0 {
			feed := OpmlFeed{Title: outline.Text}
			if outline.Url == "" {
				feed.Url = outline.XmlUrl
			} else {
				feed.Url = outline.Url
			}
			if tag != "" {
				feed.Tags = append(feed.Tags, tag)
			}
			opml.Feeds = append(opml.Feeds, feed)
		} else {
			processOutline(opml, outline.Outline, outline.Text)
		}
	}
}
