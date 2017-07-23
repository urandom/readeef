package processor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/parser/processor"
)

type RelativeURL struct {
	log readeef.Logger
}

func NewRelativeURL(log readeef.Logger) RelativeURL {
	return RelativeURL{log: log}
}

func (p RelativeURL) ProcessArticles(ua []content.UserArticle) []content.UserArticle {
	if len(ua) == 0 {
		return ua
	}

	p.log.Infof("Proxying urls of feed '%d'\n", ua[0].Data().FeedId)

	for i := range ua {
		data := ua[i].Data()

		if d, err := goquery.NewDocumentFromReader(strings.NewReader(data.Description)); err == nil {
			if processor.RelativizeArticleLinks(d) {
				if content, err := d.Html(); err == nil {
					// net/http tries to provide valid html, adding html, head and body tags
					content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]

					data.Description = content
					ua[i].Data(data)
				}
			}
		}
	}

	return ua
}
