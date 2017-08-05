package processor

import (
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/parser/processor"
)

type ProxyHTTP struct {
	urlTemplate *template.Template
	logger      readeef.Logger
}

func NewProxyHTTP(urlTemplate string, log readeef.Logger) (ProxyHTTP, error) {
	log.Infof("URL Template: %s\n", urlTemplate)
	t, err := template.New("proxy-http-url-template").Parse(urlTemplate)
	if err != nil {
		return ProxyHTTP{}, errors.Wrap(err, "parsing template")
	}

	return ProxyHTTP{logger: log, urlTemplate: t}, nil
}

func (p ProxyHTTP) Process(articles []content.Article) []content.Article {
	if len(articles) == 0 {
		return articles
	}

	p.logger.Infof("Proxying urls of feed '%d'\n", articles[0].Data().FeedId)

	for i := range articles {
		data := articles[i].Data()

		if d, err := goquery.NewDocumentFromReader(strings.NewReader(data.Description)); err == nil {
			if processor.ProxyArticleLinks(d, p.urlTemplate, data.Link) {
				if content, err := d.Html(); err == nil {
					// net/http tries to provide valid html, adding html, head and body tags
					content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]

					data.Description = content
					articles[i].Data(data)
				}
			}
		}
	}

	return articles
}
