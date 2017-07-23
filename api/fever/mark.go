package fever

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func unreadRecent(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
	log.Infoln("Marking recently read fever items as unread")

	t := time.Now().Add(-24 * time.Hour)
	user.ReadState(false, data.ArticleUpdateStateOptions{
		BeforeDate: time.Now(),
		AfterDate:  t,
	})

	if user.HasErr() {
		return errors.Wrap(user.Err(), "marking recently read articles as unread")
	}

	return nil
}

func markItem(r *http.Request, resp resp, user content.User, log readeef.Logger) error {
	val := r.FormValue("mark")
	if val == "item" {
		log.Infof("Marking fever item '%s' as '%s'\n", r.PostFormValue("id"), r.PostFormValue("as"))

		id, err := strconv.ParseInt(r.PostFormValue("id"), 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parsing id %s", r.FormValue("id"))
		}

		article := user.ArticleById(data.ArticleId(id), data.ArticleQueryOptions{SkipSessionProcessors: true})
		if user.HasErr() {
			return errors.Wrapf(user.Err(), "getting user article %d", id)
		}

		switch r.FormValue("as") {
		case "read":
			article.Read(true)
		case "saved":
			article.Favorite(true)
		case "unsaved":
			article.Favorite(false)
		default:
			err = errors.New("Unknown 'as' action")
		}
		if err == nil {
			err = article.Err()
		}

		if err != nil {
			return errors.Wrapf(err, "executing mark command %s", r.FormValue("as"))
		}
	} else if val == "feed" || val == "group" {
		log.Infof("Marking fever %s '%s' as '%s'\n", val, r.FormValue("id"), r.FormValue("as"))
		if r.FormValue("as") != "read" {
			return errors.Errorf("unknown command %s", r.FormValue("as"))
		}

		id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parsing id %s", r.FormValue("id"))
		}

		timestamp, err := strconv.ParseInt(r.FormValue("before"), 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parsing before value %s", r.FormValue("before"))
		}

		t := time.Unix(timestamp, 0)

		if val == "feed" {
			feed := user.FeedById(data.FeedId(id))
			feed.ReadState(true, data.ArticleUpdateStateOptions{
				BeforeDate: t,
			})

			if feed.HasErr() {
				return errors.Wrapf(feed.Err(), "operating on feed %d", id)
			}
		} else if val == "group" {
			if id == 1 || id == 0 {
				user.ReadState(true, data.ArticleUpdateStateOptions{
					BeforeDate: t,
				})
				if user.HasErr() {
					return errors.Wrap(user.Err(), "marking user articles as read")
				}
			} else {
				return errors.Errorf("Unknown group %d\n", id)
			}
		}
	}

	return nil
}
