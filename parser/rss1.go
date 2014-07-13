package parser

import "encoding/xml"

type rss1Feed struct {
	XMLName xml.Name    `xml:"RDF"`
	Channel rss1Channel `xml:"channel"`
	Items   []rssItem   `xml:"item"`
}

type rss1Channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Image       rssImage `xml:"image"`
}

func ParseRss1(b []byte) (feed, error) {
	var f feed
	var rss rss1Feed

	if err := xml.Unmarshal(b, &rss); err != nil {
		return f, err
	}

	f = feed{
		title:       rss.Channel.Title,
		description: rss.Channel.Description,
		link:        rss.Channel.Link,
		image: image{
			rss.Channel.Image.Title, rss.Channel.Image.Url,
			rss.Channel.Image.Width, rss.Channel.Image.Height},
	}

	for _, i := range rss.Items {
		article := article{id: i.Id, title: i.Title, description: i.Description, link: i.Link}

		var err error
		if i.PubDate != "" {
			if article.date, err = parseDate(i.PubDate); err != nil {
				return f, err
			}
		} else {
			if article.date, err = parseDate(i.Date); err != nil {
				return f, err
			}
		}
		f.articles = append(f.articles, article)
	}

	f.hubLink = getHubLink(b)

	return f, nil
}
