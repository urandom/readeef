package fever

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type item struct {
	Id            content.ArticleID `json:"id"`
	FeedId        content.FeedID    `json:"feed_id"`
	Title         string            `json:"title"`
	Author        string            `json:"author"`
	Html          string            `json:"html"`
	Url           string            `json:"url"`
	IsSaved       int               `json:"is_saved"`
	IsRead        int               `json:"is_read"`
	CreatedOnTime int64             `json:"created_on_time"`
}

func registerItemActions(processors []processor.Article) {
	actions["items"] = func(r *http.Request, resp resp, user content.User, service repo.Service, log log.Log) error {
		return items(r, resp, user, service, processors, log)
	}
}

func items(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	processors []processor.Article,
	log log.Log,
) error {
	log.Infoln("Fetching fever items")

	// Do not use filters in the count, it takes twice as long for something
	// that doesn't need to be precise.
	count, err := service.ArticleRepo().Count(user)

	if err != nil {
		return errors.WithMessage(err, "getting user article count")
	}

	items := []item{}
	if count > 0 {
		var since, max int64

		if val := r.FormValue("since_id"); val != "" {
			// On error, since == 0
			since, _ = strconv.ParseInt(val, 10, 64)
		}

		if val := r.FormValue("max_id"); val != "" {
			// On error, max == 0
			max, _ = strconv.ParseInt(val, 10, 64)
		}

		var articles []content.Article

		// Fever clients do their own paging
		opts := []content.QueryOpt{
			content.Paging(50, 0),
			content.Filters(content.GetUserFilters(user)),
		}

		if withIds, ok := r.Form["with_ids"]; ok {
			stringIds := strings.Split(withIds[0], ",")
			ids := make([]content.ArticleID, 0, len(stringIds))

			for _, stringID := range stringIds {
				stringID = strings.TrimSpace(stringID)

				if id, err := strconv.ParseInt(stringID, 10, 64); err == nil {
					ids = append(ids, content.ArticleID(id))
				}
			}

			opts = append(opts, content.IDs(ids))
		}

		if max > 0 {
			opts = append(opts,
				content.IDRange(0, content.ArticleID(max)),
				content.Sorting(content.DefaultSort, content.DescendingOrder))
		} else {
			opts = append(opts,
				content.IDRange(content.ArticleID(since), 0),
				content.Sorting(content.DefaultSort, content.AscendingOrder))
		}

		articles, err := service.ArticleRepo().ForUser(user, opts...)
		if err != nil {
			return errors.WithMessage(err, "getting user articles")
		}

		articles = processor.Articles(processors).Process(articles)

		for _, a := range articles {
			item := item{
				Id: a.ID, FeedId: a.FeedID, Title: a.Title, Html: a.Description,
				Url: a.Link, CreatedOnTime: a.Date.Unix(),
			}
			if a.Read {
				item.IsRead = 1
			}
			if a.Favorite {
				item.IsSaved = 1
			}
			items = append(items, item)
		}
	}

	resp["total_items"] = count
	resp["items"] = items

	return nil
}
