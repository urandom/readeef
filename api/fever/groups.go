package fever

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type group struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type feedsGroup struct {
	GroupId int64  `json:"group_id"`
	FeedIds string `json:"feed_ids"`
}

func groups(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	log log.Log,
) error {
	log.Infoln("Fetching fever groups")

	tags, err := service.TagRepo().ForUser(user)
	if err != nil {
		return errors.WithMessage(err, "getting user tags")
	}

	g := make([]group, len(tags))
	fg := make([]feedsGroup, len(tags))

	feedRepo := service.FeedRepo()
	for i, tag := range tags {
		g[i] = group{Id: int64(tag.ID), Title: string(tag.Value)}

		feeds, err := feedRepo.ForTag(tag, user)
		if err != nil {
			return errors.WithMessage(err, "getting tag feeds")
		}

		ids := make([]string, len(feeds))
		for j := range feeds {
			ids[j] = strconv.FormatInt(int64(feeds[j].ID), 10)
		}

		fg[i] = feedsGroup{GroupId: int64(tag.ID), FeedIds: strings.Join(ids, ",")}
	}

	resp["groups"], resp["feeds_groups"] = g, fg

	return nil
}

func init() {
	actions["groups"] = groups
}
