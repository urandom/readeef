package fever

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

func unreadRecent(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	log log.Log,
) error {
	log.Infoln("Marking recently read fever items as unread")

	t := time.Now().Add(-24 * time.Hour)
	err := service.ArticleRepo().Read(false, user,
		content.TimeRange(t, time.Now()),
		content.Filters(content.GetUserFilters(user)),
	)

	if err != nil {
		return errors.WithMessage(err, "marking recently read articles as unread")
	}

	return nil
}

func markItem(
	r *http.Request,
	resp resp,
	user content.User,
	service repo.Service,
	log log.Log,
) error {
	val := r.FormValue("mark")
	opts := []content.QueryOpt{
		content.Filters(content.GetUserFilters(user)),
	}

	id, err := strconv.ParseInt(r.PostFormValue("id"), 10, 64)
	if err != nil {
		return errors.Wrapf(err, "parsing id %s", r.FormValue("id"))
	}

	switch val {
	case "item":
		opts = append(opts, content.IDs([]content.ArticleID{content.ArticleID(id)}))
	case "group", "feed":
		timestamp, err := strconv.ParseInt(r.FormValue("before"), 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parsing before value %s", r.FormValue("before"))
		}

		opts = append(opts, content.TimeRange(time.Time{}, time.Unix(timestamp, 0)))

		if id == 0 {
			break
		}

		if val == "feed" {
			opts = append(opts, content.FeedIDs([]content.FeedID{content.FeedID(id)}))
		} else {
			tagRepo := service.TagRepo()
			tag, err := tagRepo.Get(content.TagID(id), user)
			if err != nil {
				return errors.WithMessage(err, "getting user tag")
			}

			ids, err := tagRepo.FeedIDs(tag, user)
			if err != nil {
				return errors.WithMessage(err, "getting tag feed ids")
			}

			opts = append(opts, content.FeedIDs(ids))
		}
	default:
		return errors.Errorf("unknown mark type %s", val)
	}

	switch action := r.FormValue("as"); action {
	case "read":
		return service.ArticleRepo().Read(true, user, opts...)
	case "saved":
		return service.ArticleRepo().Favor(true, user, opts...)
	case "unsaved":
		return service.ArticleRepo().Favor(false, user, opts...)
	default:
		return errors.Errorf("unknown action %s", action)
	}
}

func init() {
	actions["unread_recently_read"] = unreadRecent
	actions["mark"] = markItem
}
