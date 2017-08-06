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

func (r feedRepo) Get(id content.FeedID, user content.User) (content.Feed, error) {
	r.log.Infof("Getting user %s feed %d", user, id)

	var feed content.Feed
	if user.Login == "" {
		err = r.db.Get(&feed, r.db.SQL().Repo.GetFeed, id)
	} else {
		err = r.db.Get(&feed, r.db.SQL().User.GetFeed, id, user.Login)
	}

	if err != nil {
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
	if err := r.db.Get(&feed, r.db.SQL().Repo.GetFeedByLink, link); err != nil {
		if err == sql.ErrNoRows {
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

	var feeds []content.Feed
	if err := r.db.Select(&feeds, r.db.SQL().User.GetFeeds, user.Login); err != nil {
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
	if err := r.db.Select(&feeds, r.db.SQL().Tag.GetUserFeeds, user.Login, tag.Value); err != nil {
		return []content.Feed{}, errors.Wrap(err, "getting tag feeds")
	}

	return feeds, nil
}

func (r feedRepo) All() ([]content.Feed, error) {
	r.log.Infoln("Getting all feeds")

	var feeds []content.Feed
	if err := r.db.Select(&feeds, r.db.SQL().Repo.GetFeeds); err != nil {
		return []content.Feed{}, errors.Wrap(err, "getting all feeds")
	}

	return feeds, nil
}
func (r feedRepo) IDs() ([]content.FeedID, error) {
	r.log.Info("Getting feed IDs")

	var ids []content.FeedID
	if err := r.db.Select(&ids, r.db.SQL().Feed.IDs); err != nil {
		return []content.FeedID{}, errors.Wrap(err, "getting feed ids")
	}

	return ids, nil
}

func (r feedRepo) Unsubscribed() ([]content.Feed, error) {
	r.log.Infoln("Getting all unsubscribed feeds")

	var feeds []content.Feed
	if err := r.db.Select(&feeds, r.db.SQL().Repo.GetUnsubscribedFeeds); err != nil {
		return []content.Feed{}, errors.Wrap(err, "getting unsubscribed feeds")
	}

	return feeds, nil
}

// Update updates or creates the feed data in the database.
// It returns a list of all new articles, or an error.
func (r feedRepo) Update(feed content.Feed) ([]content.Article, error) {
	newArticles := []content.Article{}

	if err := feed.Validate(); err != nil && feed.ID != 0 {
		return newArticles, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Updating feed %s", feed)

	tx, err := r.db.Beginx()
	if err != nil {
		return newArticles, errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	var changed = 0
	if feed.ID != 0 {
		s := r.db.SQL()
		stmt, err := tx.Preparex(s.Feed.Update)
		if err != nil {
			return newArticles, errors.Wrap(err, "preparing feed update stmt")
		}
		defer stmt.Close()

		res, err := stmt.Exec(feed.Link, feed.Title, feed.Description, feed.HubLink, feed.SiteLink, feed.UpdateError, feed.SubscribeError, feed.ID)
		if err != nil {
			return newArticles, errors.Wrap(err, "executing feed update stmt")
		}

		if changed, err = res.RowsAffected(); err != nil {
			changed = 0
		}
	}

	if changed == 0 {
		id, err := r.db.CreateWithID(tx, s.Feed.Create, feed.Link, feed.Title, feed.Description, feed.HubLink, feed.SiteLink, feed.UpdateError, feed.SubscribeError)

		if err != nil {
			return newArticles, errors.Wrap(err, "creating feed")
		}

		feed.ID = content.FeedID(id)
	}

	if newArticles, err = r.updateFeedArticles(feed, tx); err != nil {
		return newArticles, errors.WithMessage(err, "updating feed articles")
	}

	if err = tx.Commit(); err != nil {
		return []content.Articl{}, errors.Wrap("committing transaction")
	}

	return newArticles, nil
}

// Delete deleted the feed from the database.
func (r feedRepo) Delete(feed content.Feed) error {
	newArticles := []content.Article{}

	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Deleting feed %s", feed)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(r.db.SQL().Feed.Delete)
	if err != nil {
		return errors.Wrap(err, "preparing feed delete stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(feed.ID)
	if err != nil {
		return errors.Wrap(err, "executing feed delete stmt")
	}

	if err = tx.Commit(); err != nil {
		return []content.Articl{}, errors.Wrap("committing transaction")
	}

	return nil
}

func (r feedRepo) Users(feed content.Feed) ([]content.User, error) {
	panic("not implemented")
}

func (r feedRepo) AttachTo(content.Feed, content.User) error {
	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Attaching feed %s to %s", feed, user)

	tx, err := uf.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(uf.db.SQL().Feed.Attach)
	if err != nil {
		return errors.Wrap(err, "preparing feed attach stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Login, feed.ID)
	if err != nil {
		return errors.Wrap(err, "executing feed attach stmt")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
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

	tx, err := uf.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(uf.db.SQL().Feed.Detach)
	if err != nil {
		return errors.Wrap(err, "preparing feed detach stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Login, feed.ID)
	if err != nil {
		return errors.Wrap(err, "executing feed detach stmt")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

func (r feedRepo) SetUserTags(feed content.Feed, user content.User, tags []content.Tag) error {
	if err := feed.Validate(); err != nil {
		return errors.WithMessage(err, "validating feed")
	}

	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Setting feed %s user %s tags", feed, user)

	tx, err := uf.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.Preparex(s.Feed.DeleteUserTags)
	if err != nil {
		tf.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Login, feed.ID)
	if err != nil {
		return errors.Wrapf(err, "deleting user %s feed %s tags", user, feed)
	}

	for i := range tags {
		if tags[i].ID == 0 {
			tag, err := findTagByValue(tags[i].Value, s.Tag.GetByValue, tx)
			if err != nil {
				if content.IsNoContent(err) {
					tag, err = createTag(tags[i].Value, tx, r.db)
					if err != nil {
						return errors.Wrapf(err, "creating tag %s", tags[i].Value)
					}
				} else {
					return errors.Wrapf(err, "getting tag by value %s", tags[i].Value)
				}
			}

			tags[i] = tag
		}
	}

	if len(tags) > 0 {
		stmt, err := tx.Preparex(s.Feed.CreateUserTag)
		if err != nil {
			tf.Err(err)
			return
		}
		defer stmt.Close()

		for i := range tags {
			_, err = stmt.Exec(user.Login, feed.ID, tag[i].ID)
			if err != nil {
				return errors.Wrapf(err, "creating user %s feed %s tag %s", user, feed, tag)
			}
		}
	}

	if _, err = tx.Exec(s.Tag.DeleteStale); err != nil {
		return errors.Wrap(err, "deleting orphan tags")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

func (r feedRepo) updateFeedArticles(feed content.Feed, tx *sqlx.Tx) ([]content.Article, error) {
	articles := []content.ARticle{}

	for _, a := range feed.ParsedArticles() {
		a.FeedID = id

		var err error
		if a, err = updateArticle(a, tx, r.db, r.log); err != nil {
			return errors.WithMessage(err, "updating feed articles")
		}

		if a.IsNew {
			articles = append(articles, a)
		}
	}

	return articles
}
