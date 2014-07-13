package parser

import "encoding/xml"

type rss2Feed struct {
	XMLName xml.Name    `xml:"rss"`
	Channel rss2Channel `xml:"channel"`
}

type rss2Channel struct {
	XMLName     xml.Name  `xml:"channel"`
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Image       rssImage  `xml:"image"`
	Items       []rssItem `xml:"item"`
}

func ParseRss2(b []byte) (Feed, error) {
	var f Feed
	var rss rss2Feed

	if err := xml.Unmarshal(b, &rss); err != nil {
		return f, err
	}

	f = Feed{
		Title:       rss.Channel.Title,
		Description: rss.Channel.Description,
		Link:        rss.Channel.Link,
		Image: Image{
			rss.Channel.Image.Title, rss.Channel.Image.Url,
			rss.Channel.Image.Width, rss.Channel.Image.Height},
	}

	for _, i := range rss.Channel.Items {
		article := Article{Id: i.Id, Title: i.Title, Description: i.Description, Link: i.Link}

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
