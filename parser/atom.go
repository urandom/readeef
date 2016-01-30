package parser

import (
	"bytes"
	"encoding/xml"
	"io"
	"time"
)

type atomFeed struct {
	XMLName     xml.Name   `xml:"feed"`
	Title       string     `xml:"title"`
	Description string     `xml:"description"`
	Link        atomLink   `xml:"link"`
	Image       rssImage   `xml:"image"`
	Items       []atomItem `xml:"entry"`
}

type atomItem struct {
	XMLName     xml.Name   `xml:"entry"`
	Id          string     `xml:"id"`
	Title       string     `xml:"title"`
	Description rssContent `xml:"summary"`
	Content     rssContent `xml:"content"`
	Link        atomLink   `xml:"link"`
	Date        string     `xml:"updated"`
	PubDate     string     `xml:"published"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
}

func ParseAtom(b []byte) (Feed, error) {
	var f Feed
	var rss atomFeed

	decoder := xml.NewDecoder(bytes.NewReader(b))
	decoder.DefaultSpace = "parserfeed"

	if err := decoder.Decode(&rss); err != nil {
		return f, err
	}

	f = Feed{
		Title:       rss.Title,
		Description: rss.Description,
		SiteLink:    rss.Link.Href,
		Image: Image{
			rss.Image.Title, rss.Image.Url,
			rss.Image.Width, rss.Image.Height},
	}

	var lastValidDate time.Time
	for _, i := range rss.Items {
		article := Article{Title: i.Title, Link: i.Link.Href, Guid: i.Id}
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
