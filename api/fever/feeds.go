package fever

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type feed struct {
	Id         content.FeedID `json:"id"`
	Title      string         `json:"title"`
	Url        string         `json:"url"`
	SiteUrl    string         `json:"site_url"`
	IsSpark    int            `json:"is_spark"`
	UpdateTime int64          `json:"last_updated_on_time"`
}

func feeds(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	log log.Log,
) error {
	log.Infoln("Fetching fever feeds")

	var feverFeeds []feed

	feeds, err := service.FeedRepo().ForUser(user)
	if err != nil {
		return errors.WithMessage(err, "getting user feeds")
	}

	now := time.Now().Unix()
	for _, f := range feeds {
		feed := feed{
			Id: f.ID, Title: f.Title, Url: f.Link, SiteUrl: f.SiteLink, UpdateTime: now,
		}

		feverFeeds = append(feverFeeds, feed)
	}

	resp["feeds"] = feverFeeds

	err = groups(r, resp, user, service, log)
	delete(resp, "groups")

	return err
}

func init() {
	actions["feeds"] = feeds
}
