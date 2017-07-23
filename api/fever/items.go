package fever

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type item struct {
	Id            data.ArticleId `json:"id"`
	FeedId        data.FeedId    `json:"feed_id"`
	Title         string         `json:"title"`
	Author        string         `json:"author"`
	Html          string         `json:"html"`
	Url           string         `json:"url"`
	IsSaved       int            `json:"is_saved"`
	IsRead        int            `json:"is_read"`
	CreatedOnTime int64          `json:"created_on_time"`
}

func items(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
	log.Infoln("Fetching fever items")

	count := user.Count()
	if user.HasErr() {
		return errors.Wrap(user.Err(), "getting user article count")
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

		var articles []content.UserArticle
		// Fever clients do their own paging
		o := data.ArticleQueryOptions{Limit: 50, Offset: 0, SkipSessionProcessors: true}

		if withIds, ok := r.Form["with_ids"]; ok {
			stringIds := strings.Split(withIds[0], ",")
			ids := make([]data.ArticleId, 0, len(stringIds))

			for _, stringId := range stringIds {
				stringId = strings.TrimSpace(stringId)

				if id, err := strconv.ParseInt(stringId, 10, 64); err == nil {
					ids = append(ids, data.ArticleId(id))
				}
			}

			articles = user.ArticlesById(ids, data.ArticleQueryOptions{SkipSessionProcessors: true})
			if user.HasErr() {
				return errors.Wrapf(user.Err(), "getting user articles by ids [%v]", ids)
			}
		} else if max > 0 {
			user.Order(data.DescendingOrder)
			o.BeforeId = data.ArticleId(max)
			articles = user.Articles(o)
			if user.HasErr() {
				return errors.Wrap(user.Err(), "getting user articles")
			}
		} else {
			user.Order(data.AscendingOrder)
			o.AfterId = data.ArticleId(since)
			articles = user.Articles(o)
			if user.HasErr() {
				return errors.Wrap(user.Err(), "getting user articles")
			}
		}

		for i := range articles {
			in := articles[i].Data()
			item := item{
				Id: in.Id, FeedId: in.FeedId, Title: in.Title, Html: in.Description,
				Url: in.Link, CreatedOnTime: in.Date.Unix(),
			}
			if in.Read {
				item.IsRead = 1
			}
			if in.Favorite {
				item.IsSaved = 1
			}
			items = append(items, item)
		}
	}

	resp["total_items"] = count
	resp["items"] = items

	return nil
}
