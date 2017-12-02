package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type scoresRepo struct {
	db *db.DB

	log log.Log
}

func (r scoresRepo) Get(article content.Article) (content.Scores, error) {
	if err := article.Validate(); err != nil {
		return content.Scores{}, errors.WithMessage(err, "validating article")
	}

	r.log.Debugf("Getting scores for article %s", article)

	scores := content.Scores{ArticleID: article.ID}
	if err := r.db.WithNamedStmt(r.db.SQL().Scores.Get, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&scores, scores)
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Scores{}, errors.Wrapf(err, "getting scores for article %s", article)
	}

	return scores, nil
}

func (r scoresRepo) Update(scores content.Scores) error {
	if err := scores.Validate(); err != nil {
		return errors.WithMessage(err, "validating scores")
	}

	r.log.Infof("Updating scores %s", scores)

	s := r.db.SQL()
	return r.db.WithTx(func(tx *sqlx.Tx) error {
		return r.db.WithNamedStmt(s.Scores.Update, tx, func(stmt *sqlx.NamedStmt) error {
			res, err := stmt.Exec(scores)
			if err != nil {
				return errors.Wrap(err, "executing scores update stmt")
			}

			if num, err := res.RowsAffected(); err == nil && num > 0 {
				return nil
			}

			return r.db.WithNamedStmt(s.Scores.Create, tx, func(stmt *sqlx.NamedStmt) error {
				if _, err = stmt.Exec(scores); err != nil {
					return errors.Wrap(err, "executing scores create stmt")
				}

				return nil
			})
		})
	})
}
