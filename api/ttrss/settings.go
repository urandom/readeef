package ttrss

import (
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

type subscribeContent struct {
	Status struct {
		Code int `json:"code"`
	} `json:"status"`
}

func registerSettingActions(feedManager *readeef.FeedManager, update time.Duration) {
	actions["getPref"] = func(req request, user content.User, service repo.Service) (interface{}, error) {
		return getPref(req, user, update, service)
	}
	actions["shareToPublished"] = shareToPublished
	actions["subscribeToFeed"] = func(req request, user content.User, service repo.Service) (interface{}, error) {
		return subscribeToFeed(req, user, feedManager, service)
	}
	actions["unsubscribeFeed"] = func(req request, user content.User, service repo.Service) (interface{}, error) {
		return unsubscribeFeed(req, user, feedManager, service)
	}
}

func getPref(
	req request,
	user content.User,
	update time.Duration,
	service repo.Service,
) (interface{}, error) {
	switch req.PrefName {
	case "DEFAULT_UPDATE_INTERVAL":
		return genericContent{Value: int(update.Minutes())}, nil
	case "DEFAULT_ARTICLE_LIMIT":
		return genericContent{Value: 200}, nil
	case "HIDE_READ_FEEDS":
		return genericContent{Value: user.ProfileData["unreadOnly"]}, nil
	case "FEEDS_SORT_BY_UNREAD", "ENABLE_FEED_CATS", "SHOW_CONTENT_PREVIEW":
		return genericContent{Value: true}, nil
	case "FRESH_ARTICLE_MAX_AGE":
		return genericContent{Value: (-1 * FRESH_DURATION).Hours()}, nil
	default:
		return unknown(req, user, service)
	}
}

func shareToPublished(req request, user content.User, service repo.Service) (interface{}, error) {
	return nil, errors.WithStack(newErr("unsupported operation", "Publishing failed"))
}

func subscribeToFeed(
	req request,
	user content.User,
	feedManager *readeef.FeedManager,
	service repo.Service,
) (interface{}, error) {
	repo := service.FeedRepo()
	feed, err := repo.FindByLink(req.FeedUrl)
	if content.IsNoContent(err) {
		feed, err = feedManager.AddFeedByLink(req.FeedUrl)
		if err != nil {
			return nil, errors.WithStack(newErr(err.Error(), "INCORRECT_USAGE"))
		}
	} else {
		if err != nil {
			return nil, errors.WithMessage(err, "getting feed by link "+req.FeedUrl)
		}

		users, err := repo.Users(feed)
		if err != nil {
			return nil, errors.WithMessage(err, "getting users for feed")
		}

		for _, u := range users {
			if u.Login == user.Login {
				return subscribeContent{Status: struct {
					Code int `json:"code"`
				}{0}}, nil
			}
		}
	}

	if err = repo.AttachTo(feed, user); err != nil {
		return nil, errors.WithMessage(err, "attaching feed to user")
	}

	return subscribeContent{Status: struct {
		Code int `json:"code"`
	}{1}}, nil
}

func unsubscribeFeed(
	req request,
	user content.User,
	feedManager *readeef.FeedManager,
	service repo.Service,
) (interface{}, error) {
	repo := service.FeedRepo()

	feed, err := repo.Get(req.FeedId, user)
	if err != nil {
		if content.IsNoContent(err) {
			return nil, errors.WithStack(newErr("no feed", "FEED_NOT_FOUND"))
		}
		return nil, errors.WithMessage(err, "getting feed for user")
	}

	if err = repo.DetachFrom(feed, user); err != nil {
		return nil, errors.WithMessage(err, "detaching feed from user")
	}

	users, err := repo.Users(feed)
	if err != nil {
		return nil, errors.WithMessage(err, "getting feed users")
	}

	if len(users) == 0 {
		feedManager.RemoveFeed(feed)
	}

	return genericContent{Status: "OK"}, nil
}
