package ttrss

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

type feedsContent []feed

type feed struct {
	Id          content.FeedID `json:"id"`
	Title       string         `json:"title"`
	Unread      int64          `json:"unread"`
	CatId       int            `json:"cat_id"`
	FeedUrl     string         `json:"feed_url,omitempty"`
	LastUpdated int64          `json:"last_updated,omitempty"`
	OrderId     int            `json:"order_id,omitempty"`
}

type category struct {
	Identifier string         `json:"identifier,omitempty"`
	Label      string         `json:"label,omitempty"`
	Items      []category     `json:"items,omitempty"`
	Id         string         `json:"id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Type       string         `json:"type,omitempty"`
	Unread     int64          `json:"unread,omitempty"`
	BareId     content.FeedID `json:"bare_id,omitempty"`
	Param      string         `json:"param,omitempty"`
}

type feedTreeContent struct {
	Categories category `json:"categories"`
}

func getFeeds(req request, user content.User, service repo.Service) (interface{}, error) {
	fContent := feedsContent{}

	articleRepo := service.ArticleRepo()
	if req.CatId == CAT_ALL || req.CatId == CAT_SPECIAL {
		unreadFav, err := articleRepo.Count(user,
			content.UnreadOnly, content.FavoriteOnly,
			content.Filters(content.GetUserFilters(user)),
		)
		if err != nil {
			return nil, errors.WithMessage(err, "getting unread favorite count")
		}

		if unreadFav > 0 || !req.UnreadOnly {
			fContent = append(fContent, feed{
				Id:     FAVORITE_ID,
				Title:  specialTitle(FAVORITE_ID),
				Unread: unreadFav,
				CatId:  FAVORITE_ID,
			})
		}

		freshTime := time.Now().Add(FRESH_DURATION)
		unreadFresh, err := articleRepo.Count(user,
			content.TimeRange(freshTime, time.Time{}), content.UnreadOnly,
			content.Filters(content.GetUserFilters(user)),
		)
		if err != nil {
			return nil, errors.WithMessage(err, "getting unread fresh count")
		}

		if unreadFresh > 0 || !req.UnreadOnly {
			fContent = append(fContent, feed{
				Id:     FRESH_ID,
				Title:  specialTitle(FRESH_ID),
				Unread: unreadFresh,
				CatId:  FAVORITE_ID,
			})
		}

		unreadAll, err := articleRepo.Count(user, content.UnreadOnly,
			content.Filters(content.GetUserFilters(user)),
		)
		if err != nil {
			return nil, errors.WithMessage(err, "getting unread count")
		}

		if unreadAll > 0 || !req.UnreadOnly {
			fContent = append(fContent, feed{
				Id:     ALL_ID,
				Title:  specialTitle(ALL_ID),
				Unread: unreadAll,
				CatId:  FAVORITE_ID,
			})
		}
	}

	var feeds []content.Feed
	var err error
	var catID int
	if req.CatId == CAT_ALL || req.CatId == CAT_ALL_EXCEPT_VIRTUAL {
		feeds, err = service.FeedRepo().ForUser(user)
		if err != nil {
			return nil, errors.WithMessage(err, "getting user feeds")
		}
	} else {
		if req.CatId == CAT_UNCATEGORIZED {
			allFeeds, err := service.FeedRepo().ForUser(user)
			if err != nil {
				return nil, errors.WithMessage(err, "getting user feeds")
			}

			for _, feed := range allFeeds {
				tags, err := service.TagRepo().ForFeed(feed, user)
				if err != nil {
					return nil, errors.WithMessage(err, "getting feed tags")
				}

				if len(tags) == 0 {
					feeds = append(feeds, feed)
				}
			}
		} else if req.CatId > 0 {
			catID = int(req.CatId)
			tag, err := service.TagRepo().Get(req.CatId, user)
			if err != nil {
				return nil, errors.WithMessage(err, "getting user tag")
			}

			tagged, err := service.FeedRepo().ForTag(tag, user)
			if err != nil {
				return nil, errors.WithMessage(err, "getting tag feeds")
			}
			for _, t := range tagged {
				feeds = append(feeds, t)
			}
		}
	}

	if len(feeds) > 0 {
		for i, f := range feeds {
			if req.Limit > 0 {
				if i < req.Offset || i >= req.Limit+req.Offset {
					continue
				}
			}

			unread, err := articleRepo.Count(
				user, content.UnreadOnly,
				content.FeedIDs([]content.FeedID{f.ID}),
				content.Filters(content.GetUserFilters(user)),
			)
			if err != nil {
				return nil, errors.WithMessage(err, "getting feed unread count")
			}

			if unread > 0 || !req.UnreadOnly {
				fContent = append(fContent, feed{
					Id:          f.ID,
					Title:       f.Title,
					FeedUrl:     f.Link,
					CatId:       catID,
					Unread:      unread,
					LastUpdated: time.Now().Unix(),
					OrderId:     0,
				})
			}
		}
	}

	return fContent, nil
}

func updateFeed(req request, user content.User, service repo.Service) (interface{}, error) {
	return genericContent{Status: "OK"}, nil
}

func catchupFeed(req request, user content.User, service repo.Service) (interface{}, error) {
	o := []content.QueryOpt{
		content.TimeRange(time.Time{}, time.Now()),
		content.Filters(content.GetUserFilters(user)),
	}

	var feedGenerator func() ([]content.Feed, error)

	if req.IsCat {
		tagID := content.TagID(req.FeedId)

		if tagID == CAT_UNCATEGORIZED {
			o = append(o, content.UntaggedOnly)
		} else {
			tag, err := service.TagRepo().Get(tagID, user)
			if err != nil {
				return nil, errors.WithMessage(err, "getting tag for user")
			}

			feedGenerator = func() ([]content.Feed, error) {
				return service.FeedRepo().ForTag(tag, user)
			}
		}
	} else {
		feedGenerator = func() ([]content.Feed, error) {
			feed, err := service.FeedRepo().Get(req.FeedId, user)
			return []content.Feed{feed}, err
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

		o = append(o, content.FeedIDs(ids))
	}

	if err := service.ArticleRepo().Read(true, user, o...); err != nil {
		return nil, errors.WithMessage(err, "setting read state")
	}

	return genericContent{Status: "OK"}, nil
}

func getFeedTree(req request, user content.User, service repo.Service) (interface{}, error) {
	items := []category{}

	special, err := createSpecialCategory(service.ArticleRepo(), user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting special categories")
	}
	items = append(items, special)

	feeds, err := service.FeedRepo().ForUser(user)
	if err != nil {
		return nil, errors.WithMessage(err, "getting user feeds")
	}

	uncat := category{Id: "CAT:0", Items: []category{}, BareId: 0, Name: "Uncategorized", Type: "category"}
	tagCategories := map[content.Tag]category{}

	for _, f := range feeds {
		tags, err := service.TagRepo().ForFeed(f, user)
		if err != nil {
			return nil, errors.WithMessage(err, "getting feed tags")
		}

		item, err := feedListCategoryFeed(service.ArticleRepo(), user, f, f.ID)
		if err != nil {
			return nil, err
		}

		if len(tags) > 0 {
			for _, t := range tags {
				var c category
				if cached, ok := tagCategories[t]; ok {
					c = cached
				} else {
					c = category{
						Id:     "CAT:" + strconv.FormatInt(int64(t.ID), 10),
						BareId: content.FeedID(t.ID),
						Name:   string(t.Value),
						Type:   "category",
						Items:  []category{},
					}
				}

				c.Items = append(c.Items, item)
				tagCategories[t] = c
			}
		} else {
			uncat.Items = append(uncat.Items, item)
		}
	}

	categories := []category{uncat}
	for _, c := range tagCategories {
		categories = append(categories, c)
	}

	for _, c := range categories {
		if len(c.Items) == 1 {
			c.Param = "(1 feed)"
		} else {
			c.Param = fmt.Sprintf("(%d feed)", len(c.Items))
		}
		items = append(items, c)
	}

	fl := category{Identifier: "id", Label: "name"}
	fl.Items = items

	return feedTreeContent{Categories: fl}, nil
}

func specialTitle(id content.FeedID) (t string) {
	switch id {
	case FAVORITE_ID:
		t = "Starred articles"
	case FRESH_ID:
		t = "Fresh articles"
	case ALL_ID:
		t = "All articles"
	case PUBLISHED_ID:
		t = "Published articles"
	case ARCHIVED_ID:
		t = "Archived articles"
	case RECENTLY_READ_ID:
		t = "Recently read"
	}

	return
}

func feedListCategoryFeed(
	repo repo.Article,
	user content.User,
	feed content.Feed,
	id content.FeedID,
) (category, error) {
	c := category{BareId: id, Id: "FEED:" + strconv.FormatInt(int64(id), 10), Type: "feed"}

	var err error
	if feed.ID > 0 {
		c.Name = feed.Title
		c.Unread, err = repo.Count(
			user, content.UnreadOnly,
			content.FeedIDs([]content.FeedID{feed.ID}),
			content.Filters(content.GetUserFilters(user)),
		)

		if err != nil {
			return category{}, errors.WithMessage(err, "getting feed unread count")
		}
	} else {
		c.Name = specialTitle(id)
		switch id {
		case FAVORITE_ID:
			c.Unread, err = repo.Count(user, content.UnreadOnly, content.FavoriteOnly,
				content.Filters(content.GetUserFilters(user)),
			)
		case FRESH_ID:
			c.Unread, err = repo.Count(
				user, content.UnreadOnly,
				content.TimeRange(time.Now().Add(FRESH_DURATION), time.Time{}),
				content.Filters(content.GetUserFilters(user)),
			)
		case ALL_ID:
			c.Unread, err = repo.Count(user, content.UnreadOnly,
				content.Filters(content.GetUserFilters(user)),
			)
		}

		if err != nil {
			return category{}, errors.WithMessage(err, "getting feed count")
		}
	}

	return c, nil
}

func createSpecialCategory(repo repo.Article, user content.User) (category, error) {
	ids := [...]content.FeedID{ALL_ID, FRESH_ID, FAVORITE_ID, PUBLISHED_ID, ARCHIVED_ID, RECENTLY_READ_ID}

	special := category{Id: "CAT:-1", Items: make([]category, len(ids)), Name: "Special", Type: "category", BareId: -1}

	var err error
	for i, id := range ids {
		special.Items[i], err = feedListCategoryFeed(repo, user, content.Feed{}, id)

		if err != nil {
			return category{}, err
		}
	}

	return special, nil
}

func init() {
	actions["getFeeds"] = getFeeds
	actions["updateFeed"] = updateFeed
	actions["catchupFeed"] = catchupFeed
	actions["getFeedTree"] = getFeedTree
}
