package processor

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
)

type InsertThumbnailTarget struct {
	log         readeef.Logger
	urlTemplate *template.Template
}

func NewInsertThumbnailTarget(l readeef.Logger) InsertThumbnailTarget {
	return InsertThumbnailTarget{log: l}
}

func (p InsertThumbnailTarget) ProcessArticles(articles []content.Article) []content.Article {
	if len(articles) == 0 {
		return articles
	}

	p.log.Infof("Proxying urls of feed '%d'\n", articles[0].Data().FeedId)

	for i := range articles {
		data := articles[i].Data()

		if data.ThumbnailLink == "" {
			continue
		}

		if d, err := goquery.NewDocumentFromReader(strings.NewReader(data.Description)); err == nil {
			if insertThumbnailTarget(d, data.ThumbnailLink, p.log) {
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

func insertThumbnailTarget(d *goquery.Document, thumbnailLink string, log readeef.Logger) bool {
	changed := false

	if d.Find(".top-image").Length() > 0 {
		return changed
	}

	thumbDoc, err := goquery.NewDocumentFromReader(strings.NewReader(fmt.Sprintf(`<img src="%s">`, thumbnailLink)))
	if err != nil {
		log.Infof("Error generating thumbnail image node: %v\n", err)
		return changed
	}

	d.Find("body").PrependSelection(thumbDoc.Find("img"))
	changed = true

	return changed
}
