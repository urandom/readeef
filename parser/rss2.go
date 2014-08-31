package parser

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
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
	Items       []rssItem `xml:"item"`
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

	for _, i := range rss.Channel.Items {
		article := Article{Title: i.Title, Link: i.Link}

		if i.Id == "" {
			article.Id = i.Link
		} else {
			article.Id = i.Id
		}

		hash := sha1.New()
		hash.Write([]byte(article.Id))
		article.Id = hex.EncodeToString(hash.Sum(nil))

		if i.Content == "" || len(i.Content) < len(i.Description) {
			article.Description = i.Description
		} else {
			article.Description = i.Content
		}

		var err error
		if i.PubDate != "" {
			if article.Date, err = parseDate(i.PubDate); err != nil {
				return f, err
			}
		} else {
			if article.Date, err = parseDate(i.Date); err != nil {
				return f, err
			}
		}
		f.Articles = append(f.Articles, article)
	}
	f.HubLink = getHubLink(b)

	return f, nil
}
