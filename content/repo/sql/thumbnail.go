package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type thumbnailRepo struct {
	db *db.DB

	log log.Log
}

func (r thumbnailRepo) Get(article content.Article) (content.Thumbnail, error) {
	if err := article.Validate(); err != nil {
		return content.Thumbnail{}, errors.WithMessage(err, "validating article")
	}

	r.log.Infof("Getting thumbnail for article %s", article)

	thumbnail := content.Thumbnail{ArticleID: article.ID}
	if err := r.db.WithNamedStmt(r.db.SQL().Thumbnail.Get, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&thumbnail, thumbnail)
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Thumbnail{}, errors.Wrapf(err, "getting thumbnail for article %s", article)
	}

	return thumbnail, nil
}

func (r thumbnailRepo) Update(thumbnail content.Thumbnail) error {
	if err := thumbnail.Validate(); err != nil {
		return errors.WithMessage(err, "validating thumbnail")
	}

	r.log.Infof("Updating thumbnail %s", thumbnail)

	return r.db.WithTx(func(tx *sqlx.Tx) error {
		s := r.db.SQL()
		return r.db.WithNamedStmt(s.Thumbnail.Update, tx, func(stmt *sqlx.NamedStmt) error {
			res, err := stmt.Exec(thumbnail)
			if err != nil {
				return errors.Wrap(err, "executing thumbnail update stmt")
			}

			if num, err := res.RowsAffected(); err == nil && num > 0 {
				return nil
			}

			return r.db.WithNamedStmt(s.Thumbnail.Create, tx, func(stmt *sqlx.NamedStmt) error {
				if _, err := stmt.Exec(thumbnail); err != nil {
					return errors.Wrap(err, "executing thumbnail create stmt")
				}

				return nil
			})
		})
	})
}
