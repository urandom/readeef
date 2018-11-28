package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

const (
	feedPredicates = `uid title description feed.link siteLink hubLink updateError subscribeError ttl skipHours skipDays`
)

type Feed content.Feed

func (f Feed) MarshalJSON() ([]byte, error) {
	res := feedInter{
		UID:            NewUID(int64(f.ID)),
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

	f.ID = content.FeedID(res.UID.ToInt())
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
	UID
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
	r.log.Infoln("Getting all feeds")

	resp, err := r.dg.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
feed(func: has(feed.link)) {
	%s
}
}`, feedPredicates))

	if err != nil {
		return nil, errors.Wrap(err, "getting all feeds")
	}

	var root struct {
		Feeds []Feed `json:"feed"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling feed data")
	}

	feeds := make([]content.Feed, 0, len(root.Feeds))

	for _, f := range root.Feeds {
		if f.Link == "" {
			continue
		}
		feeds = append(feeds, content.Feed(f))
	}

	return feeds, nil
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
		feed.ID = content.FeedID(UID{resp.Uids["blank-0"]}.ToInt())
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

	uid := NewUID(int64(feed.ID))
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

func (r feedRepo) Users(feed content.Feed) ([]content.User, error) {
	if err := feed.Validate(); err != nil {
		return []content.User{}, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Getting users for feed %s", feed)
	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(context.Background(), fmt.Sprintf(`
query Users($id: string) {
	users(func: uid($id)) @normalize {
		~feed {
			%s
		}
	}
}`, aliasPredicates(userPredicates)), map[string]string{"$id": intToUid(int64(feed.ID))})

	if err != nil {
		return nil, errors.Wrapf(err, "getting users of feed %s", feed)
	}

	var root struct {
		Users []User `json:"users"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling user data")
	}

	users := make([]content.User, 0, len(root.Users))

	for _, u := range root.Users {
		users = append(users, u.User)
	}

	return users, nil
}

func (r feedRepo) AttachTo(feed content.Feed, user content.User) error {
	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Attaching feed %s to %s", feed, user)

	ctx := context.Background()
	tx := r.dg.NewTxn()
	defer tx.Discard(ctx)

	uid, err := userUID(ctx, tx, user)
	if err != nil {
		return err
	}

	if !uid.Valid() {
		return errors.Errorf("Invalid user %s", user)
	}

	_, err = tx.Mutate(ctx, &api.Mutation{
		CommitNow: true,
		Set:       []*api.NQuad{{Subject: uid.Value, Predicate: "feed", ObjectId: NewUID(int64(feed.ID)).Value}},
	})

	if err != nil {
		return errors.Wrapf(err, "updating user %s, feed %s, link", user, feed)
	}

	return nil
}

func (r feedRepo) DetachFrom(feed content.Feed, user content.User) error {
	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Detaching feed %s from %s", feed, user)

	ctx := context.Background()
	tx := r.dg.NewTxn()
	defer tx.Discard(ctx)

	uid, err := userUID(ctx, tx, user)
	if err != nil {
		return err
	}

	if !uid.Valid() {
		return errors.Errorf("Invalid user %s", user)
	}

	_, err = tx.Mutate(ctx, &api.Mutation{
		CommitNow: true,
		Del:       []*api.NQuad{{Subject: uid.Value, Predicate: "feed", ObjectId: NewUID(int64(feed.ID)).Value}},
	})

	if err != nil {
		return errors.Wrapf(err, "updating user %s, feed %s, link", user, feed)
	}

	return nil
}

func (r feedRepo) SetUserTags(content.Feed, content.User, []*content.Tag) error {
	panic("not implemented")
}
