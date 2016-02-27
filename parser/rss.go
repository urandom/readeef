package parser

import (
	"encoding/xml"
	"strings"
)

type rssImage struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	Url     string   `xml:"url"`
	Width   int      `xml:"width"`
	Height  int      `xml:"height"`
}

// RssItem is the base content for both rss1 and rss2 feeds. The only reason
// it's public is because of the refrect package
type RssItem struct {
	XMLName     xml.Name   `xml:"item"`
	Id          string     `xml:"guid"`
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description rssContent `xml:"description"`
	Content     rssContent `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	PubDate     string     `xml:"pubDate"`
	Date        string     `xml:"date"`
	TTL         int        `xml:"ttl"`
	SkipHours   []int      `xml:"skipHours>hour"`
	SkipDays    []string   `xml:"skipDays>day"`
}

type rssContent struct {
	XMLName  xml.Name
	InnerXML string `xml:",innerxml"`
	Chardata string `xml:",chardata"`
}

func (c rssContent) Content() string {
	if strings.TrimSpace(c.Chardata) == "" {
		return c.InnerXML
	} else {
		return c.Chardata
	}
}

func getLargerContent(first, second rssContent) string {
	c1, c2 := first.Content(), second.Content()

	if len(c1) < len(c2) {
		return c2
	} else {
		return c1
	}
}
