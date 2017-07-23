package api

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/parser"
)

type importOPMLData struct {
	OPML   string `json:"opml"`
	DryRun bool   `json:"dryRun"`
}

func importOPML(feedManager *readeef.FeedManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload importOPMLData
		if stop := readJSON(w, r.Body, &payload); stop {
			return
		}

		opml, err := parser.ParseOpml([]byte(payload.OPML))
		if err != nil {
			http.Error(w, "Error parsing OPML: "+err.Error(), http.StatusBadRequest)
			return
		}

		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		uf := user.AllFeeds()
		if user.HasErr() {
			http.Error(w, user.Err().Error(), http.StatusInternalServerError)
			return
		}

		userFeedMap := make(map[data.FeedId]bool)
		for i := range uf {
			userFeedMap[uf[i].Data().Id] = true
		}

		var feeds []content.Feed
		var skipped []string
		for _, opmlFeed := range opml.Feeds {
			discovered, err := feedManager.DiscoverFeeds(opmlFeed.Url)
			if err != nil {
				skipped = append(skipped, opmlFeed.Url)
				continue
			}

			for _, f := range discovered {
				in := f.Data()

				if !userFeedMap[in.Id] {
					if len(opmlFeed.Tags) > 0 {
						in.Link += "#" + strings.Join(opmlFeed.Tags, ",")
					}

					f.Data(in)

					feeds = append(feeds, f)

					if !payload.DryRun {
						if err := addFeedByURL(in.Link, user, feedManager); err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)

							return
						}
					}
				}
			}
		}

		args{"feeds": feeds, "skipped": skipped}.WriteJSON(w)
	}
}

func exportOPML(feedManager *readeef.FeedManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o := parser.OpmlXml{
			Version: "1.1",
			Head:    parser.OpmlHead{Title: "Feed subscriptions of " + user.String() + " from readeef"},
		}

		if feeds := user.AllTaggedFeeds(); user.HasErr() {
			http.Error(w, user.Err().Error(), http.StatusInternalServerError)
			return
		} else {
			body := parser.OpmlBody{}
			for _, f := range feeds {
				d := f.Data()

				tags := f.Tags()
				category := make([]string, len(tags))
				for i, t := range tags {
					category[i] = string(t.Data().Value)
				}
				body.Outline = append(body.Outline, parser.OpmlOutline{
					Text:     d.Title,
					Title:    d.Title,
					XmlUrl:   d.Link,
					HtmlUrl:  d.SiteLink,
					Category: strings.Join(category, ","),
					Type:     "rss",
				})
			}

			o.Body = body
		}

		if b, err := xml.MarshalIndent(o, "", "    "); err == nil {
			args{"opml": xml.Header + string(b)}.WriteJSON(w)
		} else {
			http.Error(w, "Error marshaling opml data", http.StatusInternalServerError)
		}
	}
}
