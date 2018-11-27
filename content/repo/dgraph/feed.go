package dgraph

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

type Feed content.Feed

func (f Feed) MarshalJSON() ([]byte, error) {
	res := feedInter{
		Uid:            NewUid(int64(f.ID)),
		Title:          f.Title,
		Description:    f.Description,
		Link:           f.Link,
		SiteLink:       f.SiteLink,
		HubLink:        f.HubLink,
		UpdateError:    f.UpdateError,
		SubscribeError: f.SubscribeError,
		TTL:            f.TTL,
	}
	if len(f.SkipHours) > 0 {
		b, err := json.Marshal(f.SkipHours)
		if err != nil {
			return nil, errors.WithMessage(err, "marshaling skip hours")
		}

		res.SkipHours = string(b)
	}

	if len(f.SkipDays) > 0 {
		b, err := json.Marshal(f.SkipDays)
		if err != nil {
			return nil, errors.WithMessage(err, "marshaling skip days")
		}

		res.SkipDays = string(b)
	}

	return json.Marshal(res)
}

func (f *Feed) UnmarshalJSON(b []byte) error {
	res := feedInter{}
	if err := json.Unmarshal(b, &res); err != nil {
		return errors.WithMessage(err, "unmarshaling intermediate feed data")
	}

	if res.SkipHours != "" {
		if err := json.Unmarshal([]byte(res.SkipHours), &f.SkipHours); err != nil {
			return errors.WithMessage(err, "unmarshaling skip hours")
		}
	}

	if res.SkipDays != "" {
		if err := json.Unmarshal([]byte(res.SkipDays), &f.SkipDays); err != nil {
			return errors.WithMessage(err, "unmarshaling skip hours")
		}
	}

	f.ID = content.FeedID(res.Uid.ToInt())
	f.Title = res.Title
	f.Description = res.Description
	f.Link = res.Link
	f.SiteLink = res.SiteLink
	f.HubLink = res.HubLink
	f.UpdateError = res.UpdateError
	f.SubscribeError = res.SubscribeError
	f.TTL = res.TTL

	return nil
}

type feedInter struct {
	Uid
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Link           string        `json:"feed.link"`
	SiteLink       string        `json:"siteLink"`
	HubLink        string        `json:"hubLink"`
	UpdateError    string        `json:"updateError"`
	SubscribeError string        `json:"subscribeError"`
	TTL            time.Duration `json:"ttl"`
	SkipHours      string        `json:"skipHours"`
	SkipDays       string        `json:"skipDays"`
}

type feedRepo struct {
	dg *dgo.Dgraph

	log log.Log
}

func (r feedRepo) Get(content.FeedID, content.User) (content.Feed, error) {
	panic("not implemented")
}

func (r feedRepo) FindByLink(link string) (content.Feed, error) {
	panic("not implemented")
}

func (r feedRepo) ForUser(content.User) ([]content.Feed, error) {
	panic("not implemented")
}

func (r feedRepo) ForTag(content.Tag, content.User) ([]content.Feed, error) {
	panic("not implemented")
}

func (r feedRepo) All() ([]content.Feed, error) {
	panic("not implemented")
}

func (r feedRepo) IDs() ([]content.FeedID, error) {
	panic("not implemented")
}

func (r feedRepo) Unsubscribed() ([]content.Feed, error) {
	panic("not implemented")
}

func (r feedRepo) Update(feed *content.Feed) ([]content.Article, error) {
	newArticles := []content.Article{}

	if err := feed.Validate(); err != nil && feed.ID != 0 {
		return newArticles, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Updating feed %s", feed)

	b, err := json.Marshal(Feed(*feed))
	if err != nil {
		return newArticles, errors.Wrapf(err, "marshaling feed %s", feed)
	}

	resp, err := r.dg.NewTxn().Mutate(context.Background(), &api.Mutation{
		CommitNow: true,
		SetJson:   b,
	})

	if err != nil {
		return newArticles, errors.Wrapf(err, "updating feed %s", feed)
	}

	if len(resp.Uids) > 0 {
		feed.ID = content.FeedID(Uid{resp.Uids["blank-0"]}.ToInt())
	}

	/* FIXME:
	if newArticles, err = r.updateFeedArticles(*feed, tx); err != nil {
		return newArticles, errors.WithMessage(err, "updating feed articles")
	}
	*/

	r.log.Debugf("Feed %s new articles: %d", feed, len(newArticles))

	return newArticles, nil
}

func (r feedRepo) Delete(feed content.Feed) error {
	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Deleting feed %s", feed)

	uid := NewUid(int64(feed.ID))
	b, err := json.Marshal(uid)
	if err != nil {
		return errors.Wrapf(err, "marshaling uid for feed %s", feed)
	}

	_, err = r.dg.NewTxn().Mutate(context.Background(), &api.Mutation{
		CommitNow:  true,
		DeleteJson: b,
	})

	if err != nil {
		return errors.Wrapf(err, "deleting feed %s", feed)
	}

	return nil
}

func (r feedRepo) Users(content.Feed) ([]content.User, error) {
	panic("not implemented")
}

func (r feedRepo) AttachTo(content.Feed, content.User) error {
	panic("not implemented")
}

func (r feedRepo) DetachFrom(content.Feed, content.User) error {
	panic("not implemented")
}

func (r feedRepo) SetUserTags(content.Feed, content.User, []*content.Tag) error {
	panic("not implemented")
}
