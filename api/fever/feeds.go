package fever

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type feed struct {
	Id         data.FeedId `json:"id"`
	Title      string      `json:"title"`
	Url        string      `json:"url"`
	SiteUrl    string      `json:"site_url"`
	IsSpark    int         `json:"is_spark"`
	UpdateTime int64       `json:"last_updated_on_time"`
}

func feeds(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
	log.Infoln("Fetching fever feeds")

	var feverFeeds []feed

	feeds := user.AllFeeds()
	if user.HasErr() {
		return errors.Wrap(user.Err(), "getting user feeds")
	}

	now := time.Now().Unix()
	for i := range feeds {
		in := feeds[i].Data()
		feed := feed{
			Id: in.Id, Title: in.Title, Url: in.Link, SiteUrl: in.SiteLink, UpdateTime: now,
		}

		feverFeeds = append(feverFeeds, feed)
	}

	resp["feeds"] = feverFeeds

	err := groups(r, resp, user, log)
	delete(resp, "groups")

	return err
}
