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

func (r feedRepo) Get(id content.FeedID, user content.User) (content.Feed, error) {
	r.log.Infof("Getting user %s feed %d", user, id)

	query := `query Feed($id: string, $login: string) {
feeds as var(func: uid($id)) @cascade {
	uid feed.link
	%s
}

feeds(func: uid(feeds)) {
	%s
}
	}`
	userFilter := `
	~feed @filter(eq(login, $login))
	`
	if user.Login == "" {
		userFilter = ""
	}

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(query, userFilter, feedPredicates),
		map[string]string{"$id": intToUid(int64(id)), "$login": string(user.Login)},
	)
	if err != nil {
		return content.Feed{}, errors.Wrapf(err, "getting feed %d", id)
	}

	var root struct {
		Feeds []Feed `json:"feeds"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return content.Feed{}, errors.Wrap(err, "unmarshaling feed data")
	}

	if len(root.Feeds) == 0 {
		return content.Feed{}, content.ErrNoContent
	}

	return content.Feed(root.Feeds[0]), nil
}

func (r feedRepo) FindByLink(link string) (content.Feed, error) {
	r.log.Infof("Getting feed by link %s", link)

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(`query Feed($link: string) {
feeds(func: eq(feed.link, $link)) {
	%s
}
}`, feedPredicates),
		map[string]string{"$link": link},
	)
	if err != nil {
		return content.Feed{}, errors.Wrapf(err, "getting feed by link %s", link)
	}

	var root struct {
		Feeds []Feed `json:"feeds"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return content.Feed{}, errors.Wrap(err, "unmarshaling feed data")
	}

	if len(root.Feeds) == 0 {
		return content.Feed{}, content.ErrNoContent
	}

	return content.Feed(root.Feeds[0]), nil
}

func (r feedRepo) ForUser(user content.User) ([]content.Feed, error) {
	if err := user.Validate(); err != nil {
		return []content.Feed{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting user %s feeds", user)

	query := `query Feed($login: string) {
feeds as var(func: has(feed.link)) @cascade {
	uid feed.link
	~feed @filter(eq(login, $login))
}

feeds(func: uid(feeds)) {
	%s
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(query, feedPredicates),
		map[string]string{"$login": string(user.Login)},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "getting feeds for user %s", user)
	}

	var root struct {
		Feeds []Feed `json:"feeds"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling feed data")
	}

	feeds := make([]content.Feed, len(root.Feeds))
	for i := range root.Feeds {
		feeds[i] = content.Feed(root.Feeds[i])
	}

	return feeds, nil
}

func (r feedRepo) ForTag(tag content.Tag, user content.User) ([]content.Feed, error) {
	if err := tag.Validate(); err != nil {
		return []content.Feed{}, errors.WithMessage(err, "validating tag")
	}

	if err := user.Validate(); err != nil {
		return []content.Feed{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tag %s feeds", tag)

	query := `query Feed($login: string, $id: string) {
feeds as var(func: has(feed.link)) @cascade {
	uid feed.link
	~feed @filter(eq(login, $login))
}

forTag as var(func: uid(feeds)) @cascade {
	~feed @filter(uid($id))
}

feeds(func: uid(forTag)) {
	%s
}
	}`

	resp, err := r.dg.NewReadOnlyTxn().QueryWithVars(
		context.Background(), fmt.Sprintf(query, feedPredicates),
		map[string]string{"$id": intToUid(int64(tag.ID)), "$login": string(user.Login)},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "getting feeds for tag %s and user %s", tag, user)
	}

	var root struct {
		Feeds []Feed `json:"feeds"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling feed data")
	}

	feeds := make([]content.Feed, len(root.Feeds))
	for i := range root.Feeds {
		feeds[i] = content.Feed(root.Feeds[i])
	}

	return feeds, nil
}

func (r feedRepo) All() ([]content.Feed, error) {
	r.log.Infoln("Getting all feeds")

	resp, err := r.dg.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
  feeds as var(func: has(feed.link)) @cascade {
    uid
    feed.link
  }
  feeds(func: uid(feeds)) {
    %s
  }
}`, feedPredicates))

	if err != nil {
		return nil, errors.Wrap(err, "getting all feeds")
	}

	var root struct {
		Feeds []Feed `json:"feeds"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling feed data")
	}

	feeds := make([]content.Feed, 0, len(root.Feeds))

	for _, f := range root.Feeds {
		feeds = append(feeds, content.Feed(f))
	}

	return feeds, nil
}

func (r feedRepo) IDs() ([]content.FeedID, error) {
	r.log.Info("Getting feed IDs")

	resp, err := r.dg.NewReadOnlyTxn().Query(context.Background(), `{
  feeds as var(func: has(feed.link)) @cascade {
    uid
    feed.link
  }
    
  ids(func: uid(feeds)) {
    uid
  }
}`)

	if err != nil {
		return nil, errors.Wrap(err, "getting all feed ids")
	}

	var root struct {
		UIDs []UID `json:"ids"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return nil, errors.Wrap(err, "unmarshaling feed data")
	}

	ids := make([]content.FeedID, 0, len(root.UIDs))

	for _, uid := range root.UIDs {
		ids = append(ids, content.FeedID(uid.ToInt()))
	}

	return ids, nil
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

	_, err = tx.Mutate(ctx, &api.Mutation{
		CommitNow: true,
		Del:       []*api.NQuad{{Subject: uid.Value, Predicate: "feed", ObjectId: NewUID(int64(feed.ID)).Value}},
	})

	if err != nil {
		return errors.Wrapf(err, "updating user %s, feed %s, link", user, feed)
	}

	return nil
}

func (r feedRepo) SetUserTags(feed content.Feed, user content.User, tags []*content.Tag) error {
	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	if users, err := r.Users(feed); err == nil {
		found := false
		for _, u := range users {
			if u.Login == user.Login {
				found = true
				break
			}
		}

		if !found {
			return errors.Errorf("feed %s does not belong to user %s", feed, user)
		}
	} else {
		return errors.Wrap(err, "getting feed users")
	}

	r.log.Infof("Setting feed %s user %s tags", feed, user)

	ctx := context.Background()
	tx := r.dg.NewTxn()
	defer tx.Discard(ctx)

	query := `query Q($id: string) {
feedTags as var(func: has(tag.value)) {
	feed @filter(uid($id))
}

ids(func: uid(feedTags)) {
	uid
}
}
`

	feedUID := NewUID(int64(feed.ID)).Value
	resp, err := tx.QueryWithVars(ctx, query, map[string]string{"$id": feedUID})
	if err != nil {
		return errors.Wrapf(err, "getting tags for feed %s", feed)
	}

	var root struct {
		UIDs []UID `json:"ids"`
	}

	if err := json.Unmarshal(resp.Json, &root); err != nil {
		return errors.Wrap(err, "unmarshaling tag ids")
	}

	if len(root.UIDs) > 0 {
		nquads := make([]*api.NQuad, len(root.UIDs))
		for i := range root.UIDs {
			nquads[i] = &api.NQuad{Subject: root.UIDs[i].Value, Predicate: "feed", ObjectId: feedUID}
		}

		// Clean the current feed's tags
		_, err = tx.Mutate(ctx, &api.Mutation{
			Del: nquads,
		})
		if err != nil {
			return errors.Wrapf(err, "deleting tags for feed %s", feed)
		}
	}

	// Create any tags that do not exist yet
	for i := range tags {
		if tags[i].ID == 0 {
			uid, err := tagUID(ctx, tx, *tags[i])
			if err == nil {
				tags[i].ID = content.TagID(uid.ToInt())
			} else {
				if content.IsNoContent(err) {
					b, err := json.Marshal(Tag(*tags[i]))
					if err != nil {
						return errors.Wrapf(err, "marshaling tag %s", tags[i].Value)
					}

					if resp, err := tx.Mutate(ctx, &api.Mutation{SetJson: b}); err != nil {
						return errors.Wrapf(err, "creating tag %s", tags[i].Value)
					} else if len(resp.Uids) > 0 {
						tags[i].ID = content.TagID(UID{resp.Uids["blank-0"]}.ToInt())
					}
				} else {
					return errors.Wrapf(err, "getting tag by value %s", tags[i].Value)
				}
			}
		}
	}

	// Link tags to user and feed
	uid, err := userUID(ctx, tx, user)
	if err != nil {
		return err
	}

	nquads := make([]*api.NQuad, len(tags)*2)
	for i := range tags {
		tuid := NewUID(int64(tags[i].ID)).Value
		nquads[i] = &api.NQuad{Subject: uid.Value, Predicate: "tag", ObjectId: tuid}
		nquads[len(tags)+i] = &api.NQuad{Subject: tuid, Predicate: "feed", ObjectId: feedUID}
	}

	if _, err := tx.Mutate(ctx, &api.Mutation{Set: nquads}); err != nil {
		return errors.Wrapf(err, "settings tags for user %s and feed %s", user, feed)
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrapf(err, "committing tx for user %s feed %s tags", user, feed)
	}

	return nil
}
