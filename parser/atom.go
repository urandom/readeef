package parser

import "encoding/xml"

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
	Link        atomLink `xml:"link"`
	Date        string   `xml:"updated"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
}

func ParseAtom(b []byte) (feed, error) {
	var f feed
	var rss atomFeed

	if err := xml.Unmarshal(b, &rss); err != nil {
		return f, err
	}

	f = feed{
		title:       rss.Title,
		description: rss.Description,
		link:        rss.Link.Href,
		image: image{
			rss.Image.Title, rss.Image.Url,
			rss.Image.Width, rss.Image.Height},
	}

	for _, i := range rss.Items {
		article := article{id: i.Id, title: i.Title, description: i.Description, link: i.Link.Href}

		var err error
		if article.date, err = parseDate(i.Date); err != nil {
			return f, err
		}
		f.articles = append(f.articles, article)
	}
	f.hubLink = getHubLink(b)

	return f, nil
}
