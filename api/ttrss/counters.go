package ttrss

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

type countersContent []counter

type counter struct {
	Id         interface{} `json:"id"`
	Counter    int64       `json:"counter"`
	AuxCounter int64       `json:"auxcounter,omitempty"`
	Kind       string      `json:"kind,omitempty"`
}

func getUnread(req request, user content.User, service repo.Service) (interface{}, error) {
	opts := []content.QueryOpt{
		content.Filters(content.GetUserFilters(user)),
	}

	var feedGenerator func() ([]content.Feed, error)

	if req.IsCat {
		if req.FeedId > 0 {
			tag, err := service.TagRepo().Get(content.TagID(req.FeedId), user)
			if err != nil {
				return nil, errors.WithMessage(err, "getting tag for user")
			}

			feedGenerator = func() ([]content.Feed, error) {
				return service.FeedRepo().ForTag(tag, user)
			}
		} else if req.FeedId == CAT_UNCATEGORIZED {
			opts = append(opts, content.UntaggedOnly)
		} else if req.FeedId == CAT_SPECIAL {
			opts = append(opts, content.FavoriteOnly)
		}
	} else {
		switch req.FeedId {
		case FAVORITE_ID:
			opts = append(opts, content.FavoriteOnly)
		case FRESH_ID:
			opts = append(opts, content.TimeRange(time.Now().Add(FRESH_DURATION), time.Time{}))
		default:
			if req.FeedId > 0 {
				feed, err := service.FeedRepo().Get(req.FeedId, user)
				if err != nil {
					return nil, errors.WithMessage(err, "getting user feed")
				}

				feedGenerator = func() ([]content.Feed, error) {
					return []content.Feed{feed}, nil
				}
			}

		}
	}

	if feedGenerator != nil {
		feeds, err := feedGenerator()
		if err != nil {
			return nil, errors.WithMessage(err, "getting feeds")
		}

		ids := make([]content.FeedID, len(feeds))
		for i := range feeds {
			ids[i] = feeds[i].ID
		}

		opts = append(opts, content.FeedIDs(ids))
	}

	count, err := service.ArticleRepo().Count(user, opts...)
	if err != nil {
		return nil, err
	}

	return genericContent{Unread: strconv.FormatInt(count, 10)}, nil
}

func getCounters(req request, user content.User, service repo.Service) (interface{}, error) {
	if req.OutputMode == "" {
		req.OutputMode = "flc"
	}
	cContent := countersContent{}

	articleRepo := service.ArticleRepo()
	unreadCount, err := articleRepo.Count(user, content.UnreadOnly,
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting user unread count")
	}
	cContent = append(cContent,
		counter{Id: "global-unread", Counter: unreadCount})

	feeds, err := service.FeedRepo().ForUser(user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting user feeds")
	}
	cContent = append(cContent,
		counter{Id: "subscribed-feeds", Counter: int64(len(feeds))})

	cContent = append(cContent, counter{Id: ARCHIVED_ID})

	unreadFavCount, err := articleRepo.Count(user, content.UnreadOnly, content.FavoriteOnly,
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting favorite unread count")
	}

	favCount, err := articleRepo.Count(user, content.FavoriteOnly,
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting favorite count")
	}

	cContent = append(cContent,
		counter{Id: FAVORITE_ID,
			Counter:    unreadFavCount,
			AuxCounter: favCount})

	cContent = append(cContent, counter{Id: PUBLISHED_ID})

	freshTime := time.Now().Add(FRESH_DURATION)
	freshCount, err := articleRepo.Count(user, content.UnreadOnly,
		content.TimeRange(freshTime, time.Time{}),
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting fresh unread count")
	}
	cContent = append(cContent,
		counter{Id: FRESH_ID,
			Counter:    freshCount,
			AuxCounter: 0})

	userCount, err := articleRepo.Count(user,
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting user count")
	}
	cContent = append(cContent,
		counter{Id: ALL_ID,
			Counter:    userCount,
			AuxCounter: 0})

	for _, f := range feeds {
		feedCount, err := articleRepo.Count(user,
			content.FeedIDs([]content.FeedID{f.ID}),
			content.Filters(content.GetUserFilters(user)),
		)
		if err != nil {
			return nil, errors.WithMessage(err, "getting feed count")
		}
		cContent = append(cContent,
			counter{Id: int64(f.ID), Counter: feedCount},
		)

	}

	cContent = append(cContent, counter{Id: CAT_LABELS, Counter: 0, Kind: "cat"})

	tagRepo := service.TagRepo()
	tags, err := tagRepo.ForUser(user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting user tags")
	}
	for _, tag := range tags {
		ids, err := tagRepo.FeedIDs(tag, user)
		if err != nil {
			return nil, errors.WithMessage(err, "getting tag feed ids")
		}

		tagCount, err := articleRepo.Count(user, content.UnreadOnly,
			content.FeedIDs(ids),
			content.Filters(content.GetUserFilters(user)),
		)
		if err != nil {
			return nil, errors.WithMessage(err, "getting tag unread count")
		}
		cContent = append(cContent,
			counter{
				Id:      int64(tag.ID),
				Counter: tagCount,
				Kind:    "cat",
			},
		)
	}

	unreadUntaggedCount, err := articleRepo.Count(user,
		content.UnreadOnly, content.UntaggedOnly,
		content.Filters(content.GetUserFilters(user)),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "getting unread untagged count")
	}
	cContent = append(cContent,
		counter{
			Id:      CAT_UNCATEGORIZED,
			Counter: unreadUntaggedCount,
			Kind:    "cat",
		},
	)

	return cContent, nil
}

func init() {
	actions["getUnread"] = getUnread
	actions["getCounters"] = getCounters
}
