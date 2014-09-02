package parser

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
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
	XMLName     xml.Name `xml:"entry"`
	Id          string   `xml:"id"`
	Title       string   `xml:"title"`
	Description string   `xml:"summary"`
	Content     string   `xml:"content"`
	Link        atomLink `xml:"link"`
	Date        string   `xml:"updated"`
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

	for _, i := range rss.Items {
		article := Article{Id: i.Id, Title: i.Title, Link: i.Link.Href}

		hash := sha1.New()
		hash.Write([]byte(article.Id))
		article.Id = hex.EncodeToString(hash.Sum(nil))

		if len(i.Content) < len(i.Description) {
			article.Description = i.Description
		} else {
			article.Description = i.Content
		}

		var err error
		if article.Date, err = parseDate(i.Date); err != nil {
			return f, err
		}
		f.Articles = append(f.Articles, article)
	}
	f.HubLink = getHubLink(b)

	return f, nil
}
