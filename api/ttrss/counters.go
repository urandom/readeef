package ttrss

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type countersContent []counter

type counter struct {
	Id         interface{} `json:"id"`
	Counter    int64       `json:"counter"`
	AuxCounter int64       `json:"auxcounter,omitempty"`
	Kind       string      `json:"kind,omitempty"`
}

func getUnread(req request, user content.User) (interface{}, error) {
	var ar content.ArticleRepo
	o := data.ArticleCountOptions{UnreadOnly: true}

	if req.IsCat {
		tagId := data.TagId(req.FeedId)
		if tagId > 0 {
			ar = user.TagById(tagId)
		} else if tagId == CAT_UNCATEGORIZED {
			ar = user
			o.UntaggedOnly = true
		} else if tagId == CAT_SPECIAL {
			ar = user
			o.FavoriteOnly = true
		}
	} else {
		switch req.FeedId {
		case FAVORITE_ID:
			ar = user
			o.FavoriteOnly = true
		case FRESH_ID:
			ar = user
			o.AfterDate = time.Now().Add(FRESH_DURATION)
		case ALL_ID, 0:
			ar = user
		default:
			if req.FeedId > 0 {
				feed := user.FeedById(req.FeedId)
				if feed.HasErr() {
					return nil, errors.Wrapf(feed.Err(), "getting feed %d", req.FeedId)
				}

				ar = feed
			}

		}

	}

	if ar == nil {
		return genericContent{Unread: "0"}, nil
	}

	return genericContent{Unread: strconv.FormatInt(ar.Count(o), 10)}, nil
}

func getCounters(req request, user content.User) (interface{}, error) {
	if req.OutputMode == "" {
		req.OutputMode = "flc"
	}
	cContent := countersContent{}

	o := data.ArticleCountOptions{UnreadOnly: true}
	unreadCount := user.Count(o)
	cContent = append(cContent,
		counter{Id: "global-unread", Counter: unreadCount})

	feeds := user.AllFeeds()
	cContent = append(cContent,
		counter{Id: "subscribed-feeds", Counter: int64(len(feeds))})

	cContent = append(cContent, counter{Id: ARCHIVED_ID})

	cContent = append(cContent,
		counter{Id: FAVORITE_ID,
			Counter:    user.Count(data.ArticleCountOptions{UnreadOnly: true, FavoriteOnly: true}),
			AuxCounter: user.Count(data.ArticleCountOptions{FavoriteOnly: true})})

	cContent = append(cContent, counter{Id: PUBLISHED_ID})

	freshTime := time.Now().Add(FRESH_DURATION)
	cContent = append(cContent,
		counter{Id: FRESH_ID,
			Counter:    user.Count(data.ArticleCountOptions{UnreadOnly: true, AfterDate: freshTime}),
			AuxCounter: 0})

	cContent = append(cContent,
		counter{Id: ALL_ID,
			Counter:    user.Count(),
			AuxCounter: 0})

	for _, f := range feeds {
		cContent = append(cContent,
			counter{Id: int64(f.Data().Id), Counter: f.Count(o)},
		)

	}

	cContent = append(cContent, counter{Id: CAT_LABELS, Counter: 0, Kind: "cat"})

	for _, t := range user.Tags() {
		cContent = append(cContent,
			counter{
				Id:      int64(t.Data().Id),
				Counter: t.Count(o),
				Kind:    "cat",
			},
		)
	}

	cContent = append(cContent,
		counter{
			Id:      CAT_UNCATEGORIZED,
			Counter: user.Count(data.ArticleCountOptions{UnreadOnly: true, UntaggedOnly: true}),
			Kind:    "cat",
		},
	)

	if user.HasErr() {
		return nil, errors.Wrapf(user.Err(), "getting user %s article counters", user.Data().Login)
	}

	return cContent, nil
}

func init() {
	actions["getUnread"] = getUnread
	actions["getCounters"] = getCounters
}
