package readeef

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/urandom/webfw/util"
)
import (
	"readeef/parser"
	"regexp"
)

var (
	commentPattern = regexp.MustCompile("<!--.*?-->")
	linkPattern    = regexp.MustCompile(`<link ([^>]+)>`)
)

func getLink(link string) (string, error) {
	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	buf.ReadFrom(resp.Body)

	if _, err := parser.ParseFeed(buf.Bytes(), parser.ParseRss2, parser.ParseAtom, parser.ParseRss1); err == nil {
		return link, nil
	} else {
		html := commentPattern.ReplaceAllString(buf.String(), "")
		links := linkPattern.FindAllStringSubmatch(html, -1)

		for _, l := range links {
			attrs := l[1]
			if strings.Contains(attrs, `"application/rss+xml"`) || strings.Contains(attrs, `'application/rss+xml'`) {
				index := strings.Index(attrs, "href=")
				attr := attrs[index+6:]
				index = strings.IndexByte(attr, attrs[index+5])
				href := attr[:index]

				if u, err := url.Parse(href); err != nil {
					return "", err
				} else {
					if !u.IsAbs() {
						l, _ := url.Parse(link)

						if href[0] == '/' {
							href = l.Scheme + "://" + l.Host + href
						} else {
							href = l.Scheme + "://" + l.Host + l.Path[:strings.LastIndex(l.Path, "/")] + "/" + href
						}
					}

					return getLink(href)
				}
			}

		}
	}

	return "", errors.New("No rss link found")
}
