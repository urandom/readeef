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

type tagQuery struct {
	ID        content.TagID    `db:"id"`
	Value     content.TagValue `db:"value"`
	UserLogin content.Login    `db:"user_login"`
	FeedID    content.FeedID   `db:"feed_id"`
}

func (r tagRepo) Get(id content.TagID, user content.User) (content.Tag, error) {
	if err := user.Validate(); err != nil {
		return content.Tag{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tag %d for %s", id, user)

	tag := content.Tag{ID: id}
	if err := r.db.WithNamedStmt(r.db.SQL().Tag.Get, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&tag, tagQuery{ID: id, UserLogin: user.Login})
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Tag{}, errors.Wrapf(err, "getting tag %d", id)
	}

	return tag, nil
}

func (r tagRepo) ForUser(user content.User) ([]content.Tag, error) {
	if err := user.Validate(); err != nil {
		return []content.Tag{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting tags for %s", user)

	var tags []content.Tag
	if err := r.db.WithNamedStmt(r.db.SQL().Tag.AllForUser, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&tags, tagQuery{UserLogin: user.Login})
	}); err != nil {
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
	if err := r.db.WithNamedStmt(r.db.SQL().Tag.AllForFeed, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&tags, tagQuery{UserLogin: user.Login, FeedID: feed.ID})
	}); err != nil {
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
	if err := r.db.WithNamedStmt(r.db.SQL().Tag.GetUserFeedIDs, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&ids, tagQuery{UserLogin: user.Login, ID: tag.ID})
	}); err != nil {
		return []content.FeedID{}, errors.Wrap(err, "getting tag feed ids")
	}

	return ids, nil
}

func findTagByValue(value content.TagValue, stmt string, db *db.DB, tx *sqlx.Tx) (content.Tag, error) {
	var tag content.Tag
	if err := db.WithNamedStmt(stmt, tx, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&tag, tagQuery{Value: value})
	}); err != nil {
		if err == sql.ErrNoRows {
			return content.Tag{}, content.ErrNoContent
		}

		return content.Tag{}, errors.Wrapf(err, "getting tag by value %s", value)
	}

	return tag, nil
}

func createTag(tag content.Tag, tx *sqlx.Tx, db *db.DB) (content.Tag, error) {
	id, err := db.CreateWithID(tx, db.SQL().Tag.Create, tag)
	if err != nil {
		return content.Tag{}, errors.Wrapf(err, "creating tag %s", tag)
	}

	tag.ID = content.TagID(id)

	return tag, nil
}
