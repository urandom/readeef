package api

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/parser"
)

func importOPML(
	repo repo.Feed,
	feedManager feedManager,
	log log.Log,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, stop := userFromRequest(w, r)
		if stop {
			return
		}

		opmlData := r.Form.Get("opml")
		_, dryRun := r.Form["dryRun"]

		opml, err := parser.ParseOpml([]byte(opmlData))
		if err != nil {
			http.Error(w, "Error parsing OPML: "+err.Error(), http.StatusBadRequest)
			return
		}

		feeds, err := repo.ForUser(user)
		if err != nil {
			fatal(w, log, "Error getting user feeds: %+v", err)
			return
		}

		feedSet := feedSet{}
		for i := range feeds {
			feedSet[feeds[i].ID] = struct{}{}
		}
		feeds = make([]content.Feed, 0, 10)

		var skipped []string
		for _, opmlFeed := range opml.Feeds {
			if _, err := repo.FindByLink(opmlFeed.URL); err == nil {
				skipped = append(skipped, opmlFeed.URL)
				continue
			}

			discovered, err := feedManager.DiscoverFeeds(opmlFeed.URL)
			if err != nil {
				skipped = append(skipped, opmlFeed.URL)
				continue
			}

			for _, f := range discovered {
				if _, ok := feedSet[f.ID]; !ok {
					if len(opmlFeed.Tags) > 0 {
						f.Link += "#" + strings.Join(opmlFeed.Tags, ",")
					}

					if dryRun {
						feeds = append(feeds, f)
					} else {
						if feed, err := addFeedByURL(f.Link, user, repo, feedManager); err == nil {
							feeds = append(feeds, feed)
						} else {
							fatal(w, log, "Error adding feed: %+v", err)
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
			fatal(w, log, "Error getting user feeds: %+v", err)
			return
		}

		body := parser.OpmlBody{}
		tagRepo := service.TagRepo()
		for _, f := range feeds {
			tags, err := tagRepo.ForFeed(f, user)
			if err != nil {
				fatal(w, log, "Error getting feed tags: %+v", err)
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
