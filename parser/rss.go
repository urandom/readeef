package parser

import "encoding/xml"

type rssImage struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	Url     string   `xml:"url"`
	Width   int      `xml:"width"`
	Height  int      `xml:"height"`
}

type rssItem struct {
	XMLName     xml.Name `xml:"item"`
	Id          string   `xml:"guid"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Content     string   `xml:"content:encoded"`
	PubDate     string   `xml:"pubDate"`
	Date        string   `xml:"date"`
	TTL         int      `xml:"ttl"`
	SkipHours   []int    `xml:"skipHours>hour"`
	SkipDays    []string `xml:"skipDays>day"`
}
