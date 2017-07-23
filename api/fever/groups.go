package fever

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
)

type group struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type feedsGroup struct {
	GroupId int64  `json:"group_id"`
	FeedIds string `json:"feed_ids"`
}

func groups(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
	log.Infoln("Fetching fever groups")

	tags := user.Tags()

	if user.HasErr() {
		return errors.Wrap(user.Err(), "getting user tags")
	}

	g := make([]group, len(tags))
	fg := make([]feedsGroup, len(tags))

	for i := range tags {
		td := tags[i].Data()

		g[i] = group{Id: int64(td.Id), Title: string(td.Value)}

		feeds := tags[i].AllFeeds()
		if tags[i].HasErr() {
			return errors.Wrap(tags[i].Err(), "getting tag feeds")
		}

		ids := make([]string, len(feeds))
		for j := range feeds {
			ids[j] = strconv.FormatInt(int64(feeds[j].Data().Id), 10)
		}

		fg[i] = feedsGroup{GroupId: int64(td.Id), FeedIds: strings.Join(ids, ",")}
	}

	resp["groups"], resp["feeds_groups"] = g, fg

	return nil
}
