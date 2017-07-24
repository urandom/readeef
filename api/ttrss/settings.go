package ttrss

import (
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
)

type subscribeContent struct {
	Status struct {
		Code int `json:"code"`
	} `json:"status"`
}

func registerSettingActions(feedManager *readeef.FeedManager, update time.Duration) {
	actions["getPref"] = func(req request, user content.User) (interface{}, error) {
		return getPref(req, user, update)
	}
	actions["shareToPublished"] = shareToPublished
	actions["subscribeToFeed"] = func(req request, user content.User) (interface{}, error) {
		return subscribeToFeed(req, user, feedManager)
	}
	actions["unsubscribeFeed"] = func(req request, user content.User) (interface{}, error) {
		return unsubscribeFeed(req, user, feedManager)
	}
}

func getPref(req request, user content.User, update time.Duration) (interface{}, error) {
	switch req.PrefName {
	case "DEFAULT_UPDATE_INTERVAL":
		return genericContent{Value: int(update.Minutes())}, nil
	case "DEFAULT_ARTICLE_LIMIT":
		return genericContent{Value: 200}, nil
	case "HIDE_READ_FEEDS":
		return genericContent{Value: user.Data().ProfileData["unreadOnly"]}, nil
	case "FEEDS_SORT_BY_UNREAD", "ENABLE_FEED_CATS", "SHOW_CONTENT_PREVIEW":
		return genericContent{Value: true}, nil
	case "FRESH_ARTICLE_MAX_AGE":
		return genericContent{Value: (-1 * FRESH_DURATION).Hours()}, nil
	default:
		return unknown(req, user)
	}
}

func shareToPublished(req request, user content.User) (interface{}, error) {
	return nil, errors.WithStack(newErr("unsupported operation", "Publishing failed"))
}

func subscribeToFeed(req request, user content.User, feedManager *readeef.FeedManager) (interface{}, error) {
	f := user.Repo().FeedByLink(req.FeedUrl)
	for _, u := range f.Users() {
		if u.Data().Login == user.Data().Login {
			return subscribeContent{Status: struct {
				Code int `json:"code"`
			}{0}}, nil
		}
	}

	if f.HasErr() {
		return nil, errors.Wrapf(f.Err(), "getting feed by link %s", req.FeedUrl)
	}

	f, err := feedManager.AddFeedByLink(req.FeedUrl)
	if err != nil {
		return nil, errors.WithStack(newErr(err.Error(), "INCORRECT_USAGE"))
	}

	uf := user.AddFeed(f)
	if uf.HasErr() {
		return nil, errors.Wrapf(uf.Err(), "adding feed %s to user", req.FeedUrl)
	}

	return subscribeContent{Status: struct {
		Code int `json:"code"`
	}{1}}, nil
}

func unsubscribeFeed(req request, user content.User, feedManager *readeef.FeedManager) (interface{}, error) {
	f := user.FeedById(req.FeedId)
	f.Detach()
	users := f.Users()

	if f.HasErr() {
		err := f.Err()
		if err == content.ErrNoContent {
			return nil, errors.WithStack(newErr("no feed", "FEED_NOT_FOUND"))
		}

		return nil, errors.Wrapf(err, "getting feed by id %d", req.FeedId)
	}

	if len(users) == 0 {
		feedManager.RemoveFeed(f)
	}

	return genericContent{Status: "OK"}, nil
}
