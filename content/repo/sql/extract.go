package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type extractRepo struct {
	db *db.DB

	log log.Log
}

func (r extractRepo) Get(article content.Article) (content.Extract, error) {
	if err := article.Validate(); err != nil {
		return content.Extract{}, errors.WithMessage(err, "validating article")
	}

	r.log.Infof("Getting extract for article %s", article)

	extract := content.Extract{ArticleID: article.ID}
	if err := r.db.WithNamedStmt(r.db.SQL().Extract.Get, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&extract, extract)
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Extract{}, errors.Wrapf(err, "getting extract for article %s", article)
	}

	return extract, nil
}

func (r extractRepo) Update(extract content.Extract) error {
	if err := extract.Validate(); err != nil {
		return errors.WithMessage(err, "validating extract")
	}

	r.log.Infof("Updating extract %s", extract)

	return r.db.WithTx(func(tx *sqlx.Tx) error {
		s := r.db.SQL()
		return r.db.WithNamedStmt(s.Extract.Update, tx, func(stmt *sqlx.NamedStmt) error {
			res, err := stmt.Exec(extract)
			if err != nil {
				return errors.Wrap(err, "executing extract update stmt")
			}

			if num, err := res.RowsAffected(); err == nil && num > 0 {
				return nil
			}

			return r.db.WithNamedStmt(s.Extract.Create, tx, func(stmt *sqlx.NamedStmt) error {
				if _, err = stmt.Exec(extract); err != nil {
					return errors.Wrap(err, "executing extract create stmt")
				}
				return nil
			})
		})
	})
}
