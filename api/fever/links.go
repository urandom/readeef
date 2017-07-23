package fever

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type link struct {
	Id          data.ArticleId `json:"id"`
	FeedId      data.FeedId    `json:"feed_id"`
	ItemId      data.ArticleId `json:"item_id"`
	Temperature float64        `json:"temperature"`
	IsItem      int            `json:"is_item"`
	IsLocal     int            `json:"is_local"`
	IsSaved     int            `json:"is_saved"`
	Title       string         `json:"title"`
	Url         string         `json:"url"`
	ItemIds     string         `json:"item_ids"`
}

func links(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
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

	user.SortingByDate()
	user.Order(data.DescendingOrder)

	articles := user.Articles(data.ArticleQueryOptions{
		BeforeDate:    to,
		AfterDate:     from,
		Limit:         50,
		Offset:        50 * int(page-1),
		IncludeScores: true,
	})
	if user.HasErr() {
		return errors.Wrap(user.Err(), "getting user articles")
	}

	links := make([]link, len(articles))
	for i := range articles {
		in := articles[i].Data()

		link := link{
			Id: in.Id, FeedId: in.FeedId, ItemId: in.Id, IsItem: 1,
			IsLocal: 1, Title: in.Title, Url: in.Link, ItemIds: fmt.Sprintf("%d", in.Id),
		}

		if in.Score == 0 {
			link.Temperature = 0
		} else {
			link.Temperature = math.Log10(float64(in.Score)) / math.Log10(1.1)
		}

		if in.Favorite {
			link.IsSaved = 1
		}

		links[i] = link
	}
	resp["links"] = links

	return nil
}
