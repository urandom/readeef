package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type feedRepo struct {
	db *db.DB

	log log.Log
}

type feedQuery struct {
	ID        content.FeedID   `db:"id"`
	Link      string           `db:"link"`
	UserLogin content.Login    `db:"user_login"`
	TagValue  content.TagValue `db:"tag_value"`
}

func (r feedRepo) Get(id content.FeedID, user content.User) (content.Feed, error) {
	r.log.Infof("Getting user %s feed %d", user, id)

	args := feedQuery{ID: id, UserLogin: user.Login}
	query := r.db.SQL().Feed.GetForUser
	if user.Login == "" {
		query = r.db.SQL().Feed.Get
	}

	feed := content.Feed{ID: id}
	if err := r.db.WithNamedStmt(query, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&feed, args)
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Feed{}, errors.Wrapf(err, "getting feed %d", id)
	}

	return feed, nil
}

func (r feedRepo) FindByLink(link string) (content.Feed, error) {
	r.log.Infof("Getting feed by link %s", link)

	var feed content.Feed

	if err := r.db.WithNamedStmt(r.db.SQL().Feed.GetByLink, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&feed, feedQuery{Link: link})
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Feed{}, errors.Wrapf(err, "getting feed by link %s", link)
	}

	feed.Link = link

	return feed, nil
}

func (r feedRepo) ForUser(user content.User) ([]content.Feed, error) {
	if err := user.Validate(); err != nil {
		return []content.Feed{}, errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Getting user %s feeds", user)

	var feeds []content.Feed

	if err := r.db.WithNamedStmt(r.db.SQL().Feed.AllForUser, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&feeds, feedQuery{UserLogin: user.Login})
	}); err != nil {
		return []content.Feed{}, errors.Wrapf(err, "getting user %s feeds", user)
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

	var feeds []content.Feed

	if err := r.db.WithNamedStmt(r.db.SQL().Feed.AllForTag, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&feeds, feedQuery{UserLogin: user.Login, TagValue: tag.Value})
	}); err != nil {
		return []content.Feed{}, errors.Wrap(err, "getting tag feeds")
	}

	return feeds, nil
}

func (r feedRepo) All() ([]content.Feed, error) {
	r.log.Infoln("Getting all feeds")

	var feeds []content.Feed
	if err := r.db.WithStmt(r.db.SQL().Feed.All, nil, func(stmt *sqlx.Stmt) error {
		return stmt.Select(&feeds)
	}); err != nil {
		return []content.Feed{}, errors.Wrap(err, "getting all feeds")
	}

	return feeds, nil
}
func (r feedRepo) IDs() ([]content.FeedID, error) {
	r.log.Info("Getting feed IDs")

	var ids []content.FeedID
	if err := r.db.WithStmt(r.db.SQL().Feed.IDs, nil, func(stmt *sqlx.Stmt) error {
		return stmt.Select(&ids)
	}); err != nil {
		return []content.FeedID{}, errors.Wrap(err, "getting feed ids")
	}

	return ids, nil
}

func (r feedRepo) Unsubscribed() ([]content.Feed, error) {
	r.log.Infoln("Getting all unsubscribed feeds")

	var feeds []content.Feed
	if err := r.db.WithStmt(r.db.SQL().Feed.Unsubscribed, nil, func(stmt *sqlx.Stmt) error {
		return stmt.Select(&feeds)
	}); err != nil {
		return []content.Feed{}, errors.Wrap(err, "getting unsubscribed feeds")
	}

	return feeds, nil
}

// Update updates or creates the feed data in the database.
// It returns a list of all new articles, or an error.
func (r feedRepo) Update(feed *content.Feed) ([]content.Article, error) {
	newArticles := []content.Article{}

	if err := feed.Validate(); err != nil && feed.ID != 0 {
		return newArticles, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Updating feed %s", feed)

	s := r.db.SQL()
	if err := r.db.WithTx(func(tx *sqlx.Tx) error {
		var changed int64
		var err error

		if feed.ID != 0 {
			if err = r.db.WithNamedStmt(s.Feed.Update, tx, func(stmt *sqlx.NamedStmt) error {
				res, err := stmt.Exec(feed)
				if err != nil {
					return errors.Wrap(err, "executing feed update stmt")
				}

				if changed, err = res.RowsAffected(); err != nil {
					changed = 0
				}

				return nil
			}); err != nil {
				return errors.Wrap(err, "executing feed update stmt")
			}
		}

		if changed == 0 {
			id, err := r.db.CreateWithID(tx, r.db.SQL().Feed.Create, feed)

			if err != nil {
				return errors.Wrap(err, "executing feed create stmt")
			}

			feed.ID = content.FeedID(id)
		}

		if newArticles, err = r.updateFeedArticles(*feed, tx); err != nil {
			return errors.WithMessage(err, "updating feed articles")
		}

		return nil
	}); err != nil {
		return newArticles, err
	}

	r.log.Debugf("Feed %s new articles: %d", feed, len(newArticles))

	return newArticles, nil
}

// Delete deleted the feed from the database.
func (r feedRepo) Delete(feed content.Feed) error {
	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Deleting feed %s", feed)

	return r.db.WithNamedTx(r.db.SQL().Feed.Delete, func(stmt *sqlx.NamedStmt) error {
		if _, err := stmt.Exec(feed); err != nil {
			return errors.Wrap(err, "executing feed delete stmt")
		}
		return nil
	})
}

func (r feedRepo) Users(feed content.Feed) ([]content.User, error) {
	if err := feed.Validate(); err != nil {
		return []content.User{}, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Getting users for feed %s", feed)

	var users []content.User
	if err := r.db.WithNamedStmt(r.db.SQL().Feed.GetUsers, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&users, feed)
	}); err != nil {
		return []content.User{}, errors.Wrap(err, "getting feed users")
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

	if err := r.db.WithNamedTx(r.db.SQL().Feed.Attach, func(stmt *sqlx.NamedStmt) error {
		_, err := stmt.Exec(feedQuery{UserLogin: user.Login, ID: feed.ID})
		return err
	}); err != nil {
		return errors.Wrap(err, "executing feed attach stmt")
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

	if err := r.db.WithNamedTx(r.db.SQL().Feed.Detach, func(stmt *sqlx.NamedStmt) error {
		_, err := stmt.Exec(feedQuery{UserLogin: user.Login, ID: feed.ID})
		return err
	}); err != nil {
		return errors.Wrap(err, "executing feed detach stmt")
	}

	return nil
}

type userFeedTag struct {
	UserLogin content.Login  `db:"user_login"`
	FeedID    content.FeedID `db:"feed_id"`
	TagID     content.TagID  `db:"tag_id"`
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

	return r.db.WithTx(func(tx *sqlx.Tx) error {
		s := r.db.SQL()

		if err := r.db.WithNamedStmt(s.Feed.DeleteUserTags, tx, func(stmt *sqlx.NamedStmt) error {
			_, err := stmt.Exec(userFeedTag{UserLogin: user.Login, FeedID: feed.ID})
			return err
		}); err != nil {
			return errors.Wrapf(err, "deleting tags for feed %s", feed)
		}

		for i := range tags {
			if tags[i].ID == 0 {
				tag, err := findTagByValue(tags[i].Value, s.Tag.GetByValue, r.db, tx)
				if err != nil {
					if content.IsNoContent(err) {
						tag, err = createTag(*tags[i], tx, r.db)
						if err != nil {
							return errors.Wrapf(err, "creating tag %s", tags[i].Value)
						}
					} else {
						return errors.Wrapf(err, "getting tag by value %s", tags[i].Value)
					}
				}

				tags[i].ID = tag.ID
			}
		}

		if len(tags) > 0 {
			if err := r.db.WithNamedStmt(s.Feed.CreateUserTag, tx, func(stmt *sqlx.NamedStmt) error {
				for i := range tags {
					_, err := stmt.Exec(userFeedTag{user.Login, feed.ID, tags[i].ID})
					if err != nil {
						return errors.Wrapf(err, "creating user %s feed %s tag %s", user, feed, tags[i])
					}
				}

				return nil
			}); err != nil {
				return err
			}
		}

		return r.db.WithStmt(s.Tag.DeleteStale, tx, func(stmt *sqlx.Stmt) error {
			_, err := stmt.Exec()
			return err
		})
	})
}

func (r feedRepo) updateFeedArticles(feed content.Feed, tx *sqlx.Tx) ([]content.Article, error) {
	articles := []content.Article{}

	for _, a := range feed.ParsedArticles() {
		a.FeedID = feed.ID

		var err error
		if a, err = updateArticle(a, tx, r.db, r.log); err != nil {
			return []content.Article{}, errors.Wrap(err, "updating feed articles")
		}

		if a.IsNew {
			articles = append(articles, a)
		}
	}

	return articles, nil
}
