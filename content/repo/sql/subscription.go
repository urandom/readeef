package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type subscriptionRepo struct {
	db *db.DB

	log log.Log
}

func (r subscriptionRepo) Get(feed content.Feed) (content.Subscription, error) {
	if err := feed.Validate(); err != nil {
		return content.Subscription{}, errors.WithMessage(err, "validating feed")
	}
	r.log.Infoln("Getting feed subscription")

	subscription := content.Subscription{FeedID: feed.ID}
	if err := r.db.WithNamedStmt(r.db.SQL().Subscription.GetForFeed, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&subscription, subscription)
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Subscription{}, errors.Wrapf(err, "getting subscription for feed %s", feed)
	}

	return subscription, nil
}

func (r subscriptionRepo) All() ([]content.Subscription, error) {
	r.log.Infoln("Getting all subscriptions")

	var subscriptions []content.Subscription
	if err := r.db.WithStmt(r.db.SQL().Subscription.All, nil, func(stmt *sqlx.Stmt) error {
		return stmt.Select(&subscriptions)
	}); err != nil {
		return []content.Subscription{}, errors.Wrap(err, "getting hubbub subscriptions")
	}

	return subscriptions, nil
}

func (r subscriptionRepo) Update(subscription content.Subscription) error {
	if err := subscription.Validate(); err != nil {
		return errors.WithMessage(err, "validating subscription")
	}

	r.log.Infof("Updating subscription %s", subscription)

	return r.db.WithTx(func(tx *sqlx.Tx) error {
		s := r.db.SQL()

		r.db.WithNamedStmt(s.Subscription.Update, tx, func(stmt *sqlx.NamedStmt) error {
			res, err := stmt.Exec(subscription)
			if err != nil {
				return errors.Wrap(err, "executing subscription update stmt")
			}

			if num, err := res.RowsAffected(); err == nil && num > 0 {
				return nil
			}

			return r.db.WithNamedStmt(s.Subscription.Create, tx, func(stmt *sqlx.NamedStmt) error {
				if _, err := stmt.Exec(subscription); err != nil {
					return errors.Wrap(err, "executing subscription create stmt")
				}

				return nil
			})
		})

		return nil
	})
}
