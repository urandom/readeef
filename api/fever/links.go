package fever

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/processor"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type link struct {
	Id          content.ArticleID `json:"id"`
	FeedId      content.FeedID    `json:"feed_id"`
	ItemId      content.ArticleID `json:"item_id"`
	Temperature float64           `json:"temperature"`
	IsItem      int               `json:"is_item"`
	IsLocal     int               `json:"is_local"`
	IsSaved     int               `json:"is_saved"`
	Title       string            `json:"title"`
	Url         string            `json:"url"`
	ItemIds     string            `json:"item_ids"`
}

func registerLinkActions(processors []processor.Article) {
	actions["links"] = func(r *http.Request, resp resp, user content.User, service repo.Service, log log.Log) error {
		return links(r, resp, user, service, processors, log)
	}
}

func links(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	processors []processor.Article,
	log log.Log,
) error {
	log.Infoln("Fetching fever links")
	offset, _ := strconv.ParseInt(r.FormValue("offset"), 10, 64)

	rng, e := strconv.ParseInt(r.FormValue("range"), 10, 64)
	if e != nil {
		rng = 7
	}

	page, err := strconv.ParseInt(r.FormValue("page"), 10, 64)
	if e != nil {
		return errors.Wrapf(err, "parsing page value %s", r.FormValue("page"))
	}

	if page > 3 {
		resp["links"] = []link{}
		return nil
	}

	var from, to time.Time

	if offset == 0 {
		from = time.Now().AddDate(0, 0, int(-1*rng))
		to = time.Now()
	} else {
		from = time.Now().AddDate(0, 0, int(-1*rng-offset))
		to = time.Now().AddDate(0, 0, int(-1*offset))
	}

	articles, err := service.ArticleRepo().ForUser(
		user,
		content.TimeRange(from, to),
		content.Paging(50, 50*int(page-1)),
		content.IncludeScores,
		content.Sorting(content.DefaultSort, content.DescendingOrder),
		content.Filters(content.GetUserFilters(user)),
	)

	if err != nil {
		return errors.WithMessage(err, "getting user articles")
	}

	articles = processor.Articles(processors).Process(articles)

	links := make([]link, len(articles))
	for i, a := range articles {
		link := link{
			Id: a.ID, FeedId: a.FeedID, ItemId: a.ID, IsItem: 1,
			IsLocal: 1, Title: a.Title, Url: a.Link, ItemIds: fmt.Sprintf("%d", a.ID),
		}

		if a.Score == 0 {
			link.Temperature = 0
		} else {
			link.Temperature = math.Log10(float64(a.Score)) / math.Log10(1.1)
		}

		if a.Favorite {
			link.IsSaved = 1
		}

		links[i] = link
	}
	resp["links"] = links

	return nil
}
