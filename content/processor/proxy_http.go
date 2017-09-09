package processor

import (
	"bytes"
	"net/url"
	"path"
	"strings"
	"text/template"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
	"github.com/urandom/readeef/pool"
)

type ProxyHTTP struct {
	urlTemplate *template.Template
	logger      log.Log
}

func NewProxyHTTP(urlTemplate string, log log.Log) (ProxyHTTP, error) {
	log.Infof("URL Template: %s\n", urlTemplate)
	t, err := template.New("proxy-http-url-template").Parse(urlTemplate)
	if err != nil {
		return ProxyHTTP{}, errors.Wrap(err, "parsing template")
	}

	return ProxyHTTP{logger: log, urlTemplate: t}, nil
}

func (p ProxyHTTP) ProcessArticles(articles []content.Article) []content.Article {
	if len(articles) == 0 {
		return articles
	}

	p.logger.Infof("Proxying urls of feed '%d'", articles[0].FeedID)

	for i := range articles {
		articles[i].Description = p.processArticle(
			articles[i].Description, articles[i].Link,
		)
	}

	return articles
}

func (p ProxyHTTP) ProcessFeed(f parser.Feed) parser.Feed {
	for i := range f.Articles {
		f.Articles[i].Description = p.processArticle(
			f.Articles[i].Description, f.Articles[i].Link,
		)
	}

	return f
}

func (p ProxyHTTP) processArticle(description, link string) string {
	if d, err := goquery.NewDocumentFromReader(strings.NewReader(description)); err == nil {
		if proxyArticleLinks(d, p.urlTemplate, link) {
			if content, err := d.Html(); err == nil {
				// net/http tries to provide valid html, adding html, head and body tags
				content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]

				return content
			}
		}
	}

	return description
}

func proxyArticleLinks(d *goquery.Document, urlTemplate *template.Template, link string) bool {
	changed := false
	d.Find("[src]").Each(func(i int, s *goquery.Selection) {
		var val string
		var ok bool

		if val, ok = s.Attr("src"); !ok {
			return
		}

		if link, err := processUrl(val, urlTemplate, link); err == nil && link != "" {
			s.SetAttr("src", link)
		} else {
			return
		}

		changed = true
		return
	})

	d.Find("[srcset]").Each(func(i int, s *goquery.Selection) {
		var val string
		var ok bool

		if val, ok = s.Attr("srcset"); !ok {
			return
		}

		var res, buf bytes.Buffer

		expectUrl := true
		for _, r := range val {
			if unicode.IsSpace(r) {
				if buf.Len() != 0 {
					// From here on, only descriptors follow, until the end, or the comma
					expectUrl = false
					if link, err := processUrl(buf.String(), urlTemplate, link); err == nil && link != "" {
						res.WriteString(link)
					} else {
						return
					}
					buf.Reset()
				} // Else, whitespace right before the link
				res.WriteRune(r)
			} else if r == ',' {
				// End of the current image candidate string
				expectUrl = true
				res.WriteRune(r)
			} else if expectUrl {
				// The link
				buf.WriteRune(r)
			} else {
				// The actual descriptor text
				res.WriteRune(r)
			}
		}

		if buf.Len() > 0 {
			if link, err := processUrl(buf.String(), urlTemplate, link); err == nil && link != "" {
				res.WriteString(link)
			} else {
				return
			}
			buf.Reset()
		}

		s.SetAttr("srcset", res.String())

		changed = true
		return
	})

	return changed
}

func processUrl(link string, urlTemplate *template.Template, articleLink string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", errors.Wrapf(err, "parsing link %s", link)
	}

	if u.Scheme != "" && u.Scheme != "http" {
		return "", nil
	}

	if !u.IsAbs() {
		if ar, err := url.Parse(articleLink); err == nil {
			if u.Scheme == "" {
				u.Scheme = ar.Scheme
			}
			if u.Host == "" {
				u.Host = ar.Host
			}

			if u.Path == "" {
				return "", nil
			}

			if u.Path[0] != '/' {
				u.Path = path.Join(path.Dir(ar.Path), u.Path)
			}
		} else {
			return "", errors.Wrapf(err, "parsing article link %s", articleLink)
		}
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := urlTemplate.Execute(buf, url.QueryEscape(u.String())); err != nil {
		return "", err
	}

	return buf.String(), nil
}
