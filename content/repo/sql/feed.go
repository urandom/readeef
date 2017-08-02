package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
)

type feedRepo struct {
	db *db.DB

	log readeef.Logger
}

func (r feedRepo) Get(id content.FeedID) (content.Feed, error) {
	r.log.Infof("Getting feed %d", id)

	var feed content.Feed
	if err := r.db.Get(&feed, r.db.SQL().Repo.GetFeed, id); err != nil {
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

	if err := feed.Validate(); err != nil {
		return newArticles, errors.WithMessage(err, "validating feed")
	}

	r.log.Infof("Updating feed %s", feed)

	tx, err := r.db.Beginx()
	if err != nil {
		return newArticles, errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.Preparex(s.Feed.Update)
	if err != nil {
		return newArticles, errors.Wrap(err, "preparing feed update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.Link, i.Title, i.Description, i.HubLink, i.SiteLink, i.UpdateError, i.SubscribeError, id)
	if err != nil {
		return newArticles, errors.Wrap(err, "executimg feed update stmt")
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		id, err := r.db.CreateWithID(tx, s.Feed.Create, feed.Link, feed.Title, feed.Description, feed.HubLink, feed.SiteLink, feed.UpdateError, feed.SubscribeError)

		if err != nil {
			return newArticles, errors.Wrap(err, "creating feed")
		}

		feed.ID = data.FeedID(id)
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
func (r feedRepo) Delete(content.Feed) error {
	panic("not implemented")
}

func (r feedRepo) Users(content.Feed) ([]content.User, error) {
	panic("not implemented")
}

func (r feedRepo) DetachFrom(content.Feed, content.User) error {
	panic("not implemented")
}

func (r feedRepo) Tags(content.Feed) ([]content.Tag, error) {
	panic("not implemented")
}

func (r feedRepo) UpdateTags(content.Feed, []content.Tag) error {
	panic("not implemented")
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
