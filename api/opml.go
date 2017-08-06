package api

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
)

type importOPMLData struct {
	OPML   string `json:"opml"`
	DryRun bool   `json:"dryRun"`
}

func importOPML(
	repo repo.Feed,
	feedManager *readeef.FeedManager,
	log log.Log,
) http.HandlerFunc {
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

		feeds, err := repo.ForUser(user)
		if err != nil {
			error(w, log, "Error getting user feeds: %+v", err)
			return
		}

		feedSet := feedSet{}
		for i := range feeds {
			feedSet[feeds[i].Id] = struct{}{}
		}
		feeds = make([]content.Feed, 0, 10)

		var skipped []string
		for _, opmlFeed := range opml.Feeds {
			discovered, err := feedManager.DiscoverFeeds(opmlFeed.Url)
			if err != nil {
				skipped = append(skipped, opmlFeed.Url)
				continue
			}

			for _, f := range discovered {
				if !feedSet[f.ID] {
					if len(opmlFeed.Tags) > 0 {
						f.Link += "#" + strings.Join(opmlFeed.Tags, ",")
					}

					feeds = append(feeds, f)

					if !payload.DryRun {
						if err := addFeedByURL(f.Link, user, feedManager); err != nil {
							errors(w, log, "Error adding feed: %+v", err)
							return
						}
					}
				}
			}
		}

		args{"feeds": feeds, "skipped": skipped}.WriteJSON(w)
	}
}

func exportOPML(
	service repo.Service,
	feedManager *readeef.FeedManager,
	log log.Log,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		o := parser.OpmlXml{
			Version: "1.1",
			Head:    parser.OpmlHead{Title: "Feed subscriptions of " + user.String() + " from readeef"},
		}

		feeds, err := service.FeedRepo().ForUser(user)
		if err != nil {
			error(w, log, "Error getting user feeds: %+v", err)
			return
		}

		body := parser.OpmlBody{}
		tagRepo := service.TagRepo()
		for _, f := range feeds {
			tags, err := tagRepo.ForFeed(f, user)
			if err != nil {
				error(w, log, "Error getting feed tags: %+v", err)
				return
			}

			category := make([]string, len(tags))
			for i, t := range tags {
				category[i] = string(t.Value)
			}
			body.Outline = append(body.Outline, parser.OpmlOutline{
				Text:     f.Title,
				Title:    f.Title,
				XmlUrl:   f.Link,
				HtmlUrl:  f.SiteLink,
				Category: strings.Join(category, ","),
				Type:     "rss",
			})
		}

		o.Body = body

		if b, err := xml.MarshalIndent(o, "", "    "); err == nil {
			args{"opml": xml.Header + string(b)}.WriteJSON(w)
		} else {
			http.Error(w, "Error marshaling opml data", http.StatusInternalServerError)
		}
	}
}
