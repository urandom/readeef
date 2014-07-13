package parser

import "encoding/xml"

type pubsubFeed struct {
	Link    []pubsubLink  `xml:"link"`
	Channel pubsubChannel `xml:"channel"`
}

type pubsubChannel struct {
	Link []pubsubLink `xml:"http://www.w3.org/2005/Atom link"`
}

type pubsubLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
}

func getHubLink(b []byte) string {
	var f pubsubFeed

	if err := xml.Unmarshal(b, &f); err == nil {
		links := append(f.Link, f.Channel.Link...)
		for _, link := range links {
			if link.Rel == "hub" {
				return link.Href
			}
		}
	}

	return ""
}
