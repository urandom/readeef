package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type tagRepo struct {
	db *db.DB

	log log.Log
}

func (r tagRepo) Get(id content.TagID, user content.User) (content.Tag, error) {
	if err := user.Validate(); err != nil {
		return content.Tag{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tag %d for %s", id, user)

	var tag content.Tag
	if err := r.db.Get(&tag, r.db.SQL().Tag.Get, id, user.Login); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Tag{}, errors.Wrapf(err, "getting tag %d", id)
	}

	tag.ID = id

	return tag, nil
}

func (r tagRepo) ForUser(user content.User) ([]content.Tag, error) {
	if err := user.Validate(); err != nil {
		return []content.Tag{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tags for %s", user)

	var tags []content.Tag

	if err := r.db.Select(&tags, r.db.SQL().Tag.AllForUser, user.Login); err != nil {
		return []content.Tag{}, errors.Wrapf(err, "getting user %s tags", user)
	}

	return tags, nil
}

func (r tagRepo) ForFeed(feed content.Feed, user content.User) ([]content.Tag, error) {
	if err := user.Validate(); err != nil {
		return []content.Tag{}, errors.WithMessage(err, "validating user")
	}

	if err := feed.Validate(); err != nil {
		return []content.Tag{}, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Getting tags for user %s feed %s", user, feed)

	var tags []content.Tag
	if err := r.db.Select(&tags, r.db.SQL().Tag.AllForFeed, user.Login, feed.ID); err != nil {
		return []content.Tag{}, errors.Wrapf(err, "getting user %s feed %s tags", user, feed)
	}

	return tags, nil
}

func (r tagRepo) FeedIDs(tag content.Tag, user content.User) ([]content.FeedID, error) {
	if err := tag.Validate(); err != nil {
		return []content.FeedID{}, errors.WithMessage(err, "validating tag")
	}

	if err := user.Validate(); err != nil {
		return []content.FeedID{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tag %s feed ids", tag)

	var ids []content.FeedID
	if err := r.db.Select(&ids, r.db.SQL().Tag.GetUserFeedIDs, user.Login, tag.ID); err != nil {
		return []content.FeedID{}, errors.Wrap(err, "getting tag feed ids")
	}

	return ids, nil
}

func findTagByValue(value content.TagValue, stmt string, tx *sqlx.Tx) (content.Tag, error) {
	var tag content.Tag
	if err := tx.Get(&tag, stmt, value); err != nil {
		if err == sql.ErrNoRows {
			return content.Tag{}, content.ErrNoContent
		}

		return content.Tag{}, errors.Wrapf(err, "getting tag by value %s", value)
	}

	return tag, nil
}

func createTag(value content.TagValue, tx *sqlx.Tx, db *db.DB) (content.Tag, error) {
	id, err := db.CreateWithID(tx, db.SQL().Tag.Create, value)
	if err != nil {
		return content.Tag{}, errors.Wrapf(err, "creating tag %s", value)
	}

	return content.Tag{ID: content.TagID(id), Value: value}, nil
}
