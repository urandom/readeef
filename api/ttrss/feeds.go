package ttrss

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type feedsContent []feed

type feed struct {
	Id          data.FeedId `json:"id"`
	Title       string      `json:"title"`
	Unread      int64       `json:"unread"`
	CatId       int         `json:"cat_id"`
	FeedUrl     string      `json:"feed_url,omitempty"`
	LastUpdated int64       `json:"last_updated,omitempty"`
	OrderId     int         `json:"order_id,omitempty"`
}

type category struct {
	Identifier string      `json:"identifier,omitempty"`
	Label      string      `json:"label,omitempty"`
	Items      []category  `json:"items,omitempty"`
	Id         string      `json:"id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Type       string      `json:"type,omitempty"`
	Unread     int64       `json:"unread,omitempty"`
	BareId     data.FeedId `json:"bare_id,omitempty"`
	Param      string      `json:"param,omitempty"`
}

type feedTreeContent struct {
	Categories category `json:"categories"`
}

func getFeeds(req request, user content.User) (interface{}, error) {
	fContent := feedsContent{}

	if req.CatId == CAT_ALL || req.CatId == CAT_SPECIAL {
		unreadFav := user.Count(data.ArticleCountOptions{UnreadOnly: true, FavoriteOnly: true})

		if unreadFav > 0 || !req.UnreadOnly {
			fContent = append(fContent, feed{
				Id:     FAVORITE_ID,
				Title:  specialTitle(FAVORITE_ID),
				Unread: unreadFav,
				CatId:  FAVORITE_ID,
			})
		}

		freshTime := time.Now().Add(FRESH_DURATION)
		unreadFresh := user.Count(data.ArticleCountOptions{UnreadOnly: true, AfterDate: freshTime})

		if unreadFresh > 0 || !req.UnreadOnly {
			fContent = append(fContent, feed{
				Id:     FRESH_ID,
				Title:  specialTitle(FRESH_ID),
				Unread: unreadFresh,
				CatId:  FAVORITE_ID,
			})
		}

		unreadAll := user.Count(data.ArticleCountOptions{UnreadOnly: true})

		if unreadAll > 0 || !req.UnreadOnly {
			fContent = append(fContent, feed{
				Id:     ALL_ID,
				Title:  specialTitle(ALL_ID),
				Unread: unreadAll,
				CatId:  FAVORITE_ID,
			})
		}
	}

	var feeds []content.UserFeed
	var catId int
	if req.CatId == CAT_ALL || req.CatId == CAT_ALL_EXCEPT_VIRTUAL {
		feeds = user.AllFeeds()
	} else {
		if req.CatId == CAT_UNCATEGORIZED {
			tagged := user.AllTaggedFeeds()
			for _, t := range tagged {
				if len(t.Tags()) == 0 {
					feeds = append(feeds, t)
				}
			}
		} else if req.CatId > 0 {
			catId = int(req.CatId)
			t := user.TagById(req.CatId)
			tagged := t.AllFeeds()
			if t.HasErr() {
				return nil, errors.Wrapf(t.Err(), "getting tag %s feeds", t.Data().Value)
			}
			for _, t := range tagged {
				feeds = append(feeds, t)
			}
		}
	}

	if len(feeds) > 0 {
		o := data.ArticleCountOptions{UnreadOnly: true}
		for i := range feeds {
			if req.Limit > 0 {
				if i < req.Offset || i >= req.Limit+req.Offset {
					continue
				}
			}

			d := feeds[i].Data()
			unread := feeds[i].Count(o)

			if unread > 0 || !req.UnreadOnly {
				fContent = append(fContent, feed{
					Id:          d.Id,
					Title:       d.Title,
					FeedUrl:     d.Link,
					CatId:       catId,
					Unread:      unread,
					LastUpdated: time.Now().Unix(),
					OrderId:     0,
				})
			}
		}
	}

	if user.HasErr() {
		return nil, errors.Wrap(user.Err(), "getting user feeds")
	}

	return fContent, nil
}

func updateFeed(req request, user content.User) (interface{}, error) {
	return genericContent{Status: "OK"}, nil
}

func catchupFeed(req request, user content.User) (interface{}, error) {
	var ar content.ArticleRepo
	o := data.ArticleUpdateStateOptions{BeforeDate: time.Now()}

	if req.IsCat {
		tagId := data.TagId(req.FeedId)
		ar = user.TagById(tagId)

		if tagId == CAT_UNCATEGORIZED {
			o.UntaggedOnly = true
		}
	} else {
		ar = user.FeedById(req.FeedId)
	}

	if ar != nil {
		ar.ReadState(true, o)

		if e, ok := ar.(content.Error); ok {
			if e.HasErr() {
				return nil, errors.Wrap(e.Err(), "setting read state")
			}
		}

		return genericContent{Status: "OK"}, nil
	}

	return nil, errors.New("no article repo")
}

func getFeedTree(req request, user content.User) (interface{}, error) {
	items := []category{}

	special, err := createSpecialCategory(user)
	if err != nil {
		return nil, errors.Wrap(err, "getting special categories")
	}
	items = append(items, special)

	tf := user.AllTaggedFeeds()

	uncat := category{Id: "CAT:0", Items: []category{}, BareId: 0, Name: "Uncategorized", Type: "category"}
	tagCategories := map[content.Tag]category{}

	for _, f := range tf {
		tags := f.Tags()

		item, err := feedListCategoryFeed(user, f, f.Data().Id, true)
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
						Id:     "CAT:" + strconv.FormatInt(int64(t.Data().Id), 10),
						BareId: data.FeedId(t.Data().Id),
						Name:   string(t.Data().Value),
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

	if user.HasErr() {
		return nil, errors.Wrap(user.Err(), "getting feed tree")
	} else {
		return feedTreeContent{Categories: fl}, nil
	}
}

func specialTitle(id data.FeedId) (t string) {
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

func feedListCategoryFeed(u content.User, f content.UserFeed, id data.FeedId, includeUnread bool) (category, error) {
	c := category{BareId: id, Id: "FEED:" + strconv.FormatInt(int64(id), 10), Type: "feed"}

	copts := data.ArticleCountOptions{UnreadOnly: true}
	if f != nil {
		c.Name = f.Data().Title
		c.Unread = f.Count(copts)

		if f.HasErr() {
			return category{}, errors.Wrap(f.Err(), "getting feed unread count")
		}
	} else {
		c.Name = specialTitle(id)
		switch id {
		case FAVORITE_ID:
			copts.FavoriteOnly = true
			c.Unread = u.Count(copts)
		case FRESH_ID:
			copts.AfterDate = time.Now().Add(FRESH_DURATION)
			c.Unread = u.Count(copts)
		case ALL_ID:
			c.Unread = u.Count(copts)
		}

		if u.HasErr() {
			return category{}, errors.Wrap(u.Err(), "getting user unread count")
		}
	}

	return c, nil
}

func createSpecialCategory(user content.User) (category, error) {
	ids := [...]data.FeedId{ALL_ID, FRESH_ID, FAVORITE_ID, PUBLISHED_ID, ARCHIVED_ID, RECENTLY_READ_ID}

	special := category{Id: "CAT:-1", Items: make([]category, len(ids)), Name: "Special", Type: "category", BareId: -1}

	var err error
	for i, id := range ids {
		special.Items[i], err = feedListCategoryFeed(user, nil, id, false)

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
