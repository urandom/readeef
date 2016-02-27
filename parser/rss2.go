package parser

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"time"
)

type rss2Feed struct {
	XMLName xml.Name    `xml:"rss"`
	Channel rss2Channel `xml:"channel"`
}

type rss2Channel struct {
	XMLName     xml.Name  `xml:"channel"`
	Title       string    `xml:"title"`
	Link        string    `xml:"parserfeed link"`
	Description string    `xml:"description"`
	Image       rssImage  `xml:"image"`
	Items       []RssItem `xml:"item"`
	TTL         int       `xml:"ttl"`
	SkipHours   []int     `xml:"skipHours>hour"`
	SkipDays    []string  `xml:"skipDays>day"`
}

func ParseRss2(b []byte) (Feed, error) {
	var f Feed
	var rss rss2Feed

	decoder := xml.NewDecoder(bytes.NewReader(b))
	decoder.DefaultSpace = "parserfeed"

	if err := decoder.Decode(&rss); err != nil {
		return f, err
	}

	f = Feed{
		Title:       rss.Channel.Title,
		Description: rss.Channel.Description,
		SiteLink:    rss.Channel.Link,
		Image: Image{
			rss.Channel.Image.Title, rss.Channel.Image.Url,
			rss.Channel.Image.Width, rss.Channel.Image.Height},
	}

	if rss.Channel.TTL != 0 {
		f.TTL = time.Duration(rss.Channel.TTL) * time.Minute
	}

	f.SkipHours = make(map[int]bool, len(rss.Channel.SkipHours))
	for _, v := range rss.Channel.SkipHours {
		f.SkipHours[v] = true
	}

	f.SkipDays = make(map[string]bool, len(rss.Channel.SkipDays))
	for _, v := range rss.Channel.SkipDays {
		f.SkipDays[strings.Title(v)] = true
	}

	var lastValidDate time.Time
	for _, i := range rss.Channel.Items {
		article := Article{Title: i.Title, Link: i.Link, Guid: i.Id}
		article.Description = getLargerContent(i.Content, i.Description)

		var err error
		if i.PubDate != "" {
			article.Date, err = parseDate(i.PubDate)
		} else if i.Date != "" {
			article.Date, err = parseDate(i.Date)
		} else {
			err = io.EOF
		}

		if err == nil {
			lastValidDate = article.Date.Add(time.Second)
		} else if lastValidDate.IsZero() {
			article.Date = unknownTime
		} else {
			article.Date = lastValidDate
		}

		f.Articles = append(f.Articles, article)
	}
	f.HubLink = getHubLink(b)

	return f, nil
}
