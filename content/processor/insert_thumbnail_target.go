package processor

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

type InsertThumbnailTarget struct {
	log         log.Log
	urlTemplate *template.Template
}

func NewInsertThumbnailTarget(l log.Log) InsertThumbnailTarget {
	return InsertThumbnailTarget{log: l}
}

func (p InsertThumbnailTarget) ProcessArticles(articles []content.Article) []content.Article {
	if len(articles) == 0 {
		return articles
	}

	p.log.Infof("Proxying urls of feed '%d'\n", articles[0].FeedID)

	for i := range articles {
		if articles[i].ThumbnailLink == "" {
			continue
		}

		if d, err := goquery.NewDocumentFromReader(strings.NewReader(articles[i].Description)); err == nil {
			if insertThumbnailTarget(d, articles[i].ThumbnailLink, p.log) {
				if content, err := d.Html(); err == nil {
					// net/http tries to provide valid html, adding html, head and body tags
					content = content[strings.Index(content, "<body>")+6 : strings.LastIndex(content, "</body>")]

					articles[i].Description = content
				}
			}
		}
	}

	return articles
}

func insertThumbnailTarget(d *goquery.Document, thumbnailLink string, log log.Log) bool {
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
