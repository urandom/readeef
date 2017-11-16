package parser

import (
	"encoding/xml"
	"strings"
)

type Opml struct {
	Feeds []OpmlFeed
}

type OpmlFeed struct {
	Title string
	URL   string
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
	URL      string        `xml:"url,attr,omitempty"`
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

func processOutline(opml *Opml, outlines []OpmlOutline, text string) {
	for _, outline := range outlines {
		if len(outline.Outline) == 0 {
			feed := OpmlFeed{Title: outline.Text}
			if outline.URL == "" {
				feed.URL = outline.XmlUrl
			} else {
				feed.URL = outline.URL
			}

			tagSet := map[string]struct{}{}
			if text == "" {
				if outline.Category != "" {
					for _, tag := range strings.Split(outline.Category, ",") {
						tagSet[strings.TrimSpace(tag)] = struct{}{}
					}
				}
			} else {
				for _, tag := range strings.Split(text, ",") {
					tagSet[strings.TrimSpace(tag)] = struct{}{}
				}
			}

			feed.Tags = make([]string, 0, len(tagSet))
			for tag := range tagSet {
				feed.Tags = append(feed.Tags, tag)
			}
			opml.Feeds = append(opml.Feeds, feed)
		} else {
			processOutline(opml, outline.Outline, outline.Text)
		}
	}
}
