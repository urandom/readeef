package kv

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

type feedRepo struct {
	db  *storm.DB
	log log.Log
}

type feedUser struct {
	ID     int            `storm:"increment"`
	FeedID content.FeedID `storm:"index"`
	Login  content.Login  `storm:"index"`
}

type feedTag struct {
	ID     int            `storm:"increment"`
	FeedID content.FeedID `storm:"index"`
	TagID  content.TagID  `storm:"index"`
}

type feedSub struct {
	FeedID content.FeedID `storm:"id"`
}

const (
	feedsBucket      = "feeds"
	feedsUsersBucket = "users"
	feedsTagsBucket  = "tags"
	feedsSubsBucket  = "subs"
)

func (r feedRepo) Get(id content.FeedID, user content.User) (content.Feed, error) {
	r.log.Infof("Getting user %s feed %d", user, id)

	if user.Login != "" {
		var feedUsers []feedUser
		if err := r.db.From(feedsBucket, feedsUsersBucket).Find("FeedID", id, &feedUsers); err != nil {
			return content.Feed{}, errors.Wrapf(err, "getting feed users for feed %d", id)
		}

		hasUser := false
		for _, fu := range feedUsers {
			if fu.Login == user.Login {
				hasUser = true
				break
			}
		}

		if !hasUser {
			return content.Feed{}, errors.Wrapf(content.ErrNoContent, "no user %s for feed %d", user.Login, id)
		}
	}

	var feed content.Feed
	if err := r.db.From(feedsBucket).One("ID", id, &feed); err != nil {
		if err == storm.ErrNotFound {
			err = content.ErrNoContent
		}

		return content.Feed{}, errors.Wrapf(err, "getting feed %d", id)
	}

	return feed, nil
}

func (r feedRepo) FindByLink(link string) (content.Feed, error) {
	r.log.Infof("Getting feed by link %s", link)

	var feed content.Feed
	if err := r.db.From(feedsBucket).One("Link", link, &feed); err != nil {
		if err == storm.ErrNotFound {
			err = content.ErrNoContent
		}

		return content.Feed{}, errors.Wrapf(err, "getting feed by link %s", link)
	}

	return feed, nil
}

func (r feedRepo) ForUser(user content.User) ([]content.Feed, error) {
	if err := user.Validate(); err != nil {
		return []content.Feed{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting user %s feeds", user)

	var feedUsers []feedUser
	if err := r.db.From(feedsBucket, feedsUsersBucket).Find("Login", user.Login, &feedUsers); err != nil {
		return nil, errors.Wrapf(err, "getting feeds for user %s", user)
	}

	if len(feedUsers) == 0 {
		return nil, nil
	}

	ids := make([]content.FeedID, len(feedUsers))
	for i := range feedUsers {
		ids[i] = feedUsers[i].FeedID
	}

	var feeds []content.Feed
	if err := r.db.From(feedsBucket).Select(q.In("ID", ids)).Find(&feeds); err != nil {
		return nil, errors.Wrapf(err, "getting user %s feeds", user)
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

	var feedTags []feedTag
	if err := r.db.From(feedsBucket, feedsTagsBucket).Find("TagID", tag.ID, &feedTags); err != nil {
		return nil, errors.Wrapf(err, "getting feeds for tag %s", tag)
	}

	if len(feedTags) == 0 {
		return nil, nil
	}

	var tagUsers []tagUser
	if err := r.db.From(tagsBucket, tagsUsersBucket).Find("TagID", tag.ID, &tagUsers); err != nil {
		return nil, errors.Wrapf(err, "getting users for tag %s", tag)
	}

	hasUser := false
	for _, tu := range tagUsers {
		if tu.Login == user.Login {
			hasUser = true
			break
		}
	}

	if !hasUser {
		return nil, nil
	}

	ids := make([]content.FeedID, len(feedTags))
	for i := range feedTags {
		ids[i] = feedTags[i].FeedID
	}

	var feeds []content.Feed
	if err := r.db.From(feedsBucket).Select(q.In("ID", ids)).Find(&feeds); err != nil {
		return nil, errors.Wrapf(err, "getting user %s feeds", user)
	}

	return feeds, nil
}

func (r feedRepo) All() ([]content.Feed, error) {
	r.log.Infoln("Getting all feeds")

	var feeds []content.Feed
	if err := r.db.From(feedsBucket).All(&feeds); err != nil {
		return nil, errors.Wrap(err, "getting all feeds")
	}

	return feeds, nil
}

func (r feedRepo) IDs() ([]content.FeedID, error) {
	feeds, err := r.All()
	if err != nil {
		return nil, errors.Wrap(err, "getting feed ids")
	}

	ids := make([]content.FeedID, len(feeds))
	for i := range feeds {
		ids[i] = feeds[i].ID
	}

	return ids, nil
}

func (r feedRepo) Unsubscribed() ([]content.Feed, error) {
	panic("not implemented")
}

func (r feedRepo) Update(*content.Feed) ([]content.Article, error) {
	panic("not implemented")
}

func (r feedRepo) Delete(content.Feed) error {
	panic("not implemented")
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

func deleteFeedUserConnections(tx storm.Node, user content.User) error {
	node := tx.From(feedsBucket, feedsUsersBucket)
	var feedUsers []feedUser
	if err := node.Find("Login", user.Login, &feedUsers); err != nil {
		return errors.Wrapf(err, "getting feeds for user %s", user)
	}

	for _, fu := range feedUsers {
		if err := node.DeleteStruct(fu); err != nil {
			return errors.Wrapf(err, "deleting feed user link %d-%s", fu.FeedID, fu.Login)
		}
	}

	return nil
}

func deleteUserFeedConnections(tx storm.Node, feed content.Feed) error {
	node := tx.From(feedsBucket, feedsUsersBucket)
	var feedUsers []feedUser
	if err := node.Find("FeedID", feed.ID, &feedUsers); err != nil {
		return errors.Wrapf(err, "getting users for feed %s", feed)
	}

	for _, fu := range feedUsers {
		if err := node.DeleteStruct(fu); err != nil {
			return errors.Wrapf(err, "deleting feed user link %d-%s", fu.FeedID, fu.Login)
		}
	}

	return nil
}

func deleteTagFeedConnections(tx storm.Node, feed content.Feed) error {
	node := tx.From(feedsBucket, feedsTagsBucket)
	var feedTags []feedTag
	if err := node.Find("FeedID", feed.ID, &feedTags); err != nil {
		return errors.Wrapf(err, "getting tags for feed %s", feed)
	}

	for _, ft := range feedTags {
		if err := node.DeleteStruct(ft); err != nil {
			return errors.Wrapf(err, "deleting feed tag link %d-%d", ft.FeedID, ft.TagID)
		}
	}

	return nil
}

func newFeedRepo(db *storm.DB, log log.Log) (feedRepo, error) {
	if err := db.From(feedsBucket).Init(&content.Feed{}); err != nil {
		return feedRepo{}, errors.Wrap(err, "initializing feeds indices")
	}

	if err := db.From(feedsBucket, feedsUsersBucket).Init(&feedUser{}); err != nil {
		return feedRepo{}, errors.Wrap(err, "initializing feeds-users indices")
	}

	return feedRepo{db, log}, nil
}
