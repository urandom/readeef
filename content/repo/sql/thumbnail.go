package sql

import (
	"database/sql"

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

	var thumbnail content.Thumbnail
	if err := r.db.Get(&thumbnail, r.db.SQL().Thumbnail.Get, article.ID); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Thumbnail{}, errors.Wrapf(err, "getting thumbnail for article %s", article)
	}

	thumbnail.ArticleID = article.ID

	return thumbnail, nil
}

func (r thumbnailRepo) Update(thumbnail content.Thumbnail) error {
	if err := thumbnail.Validate(); err != nil {
		return errors.WithMessage(err, "validating thumbnail")
	}

	r.log.Infof("Updating thumbnail %s", thumbnail)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.Preparex(s.Thumbnail.Update)
	if err != nil {
		return errors.Wrap(err, "preparing thumbnail update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(thumbnail.Thumbnail, thumbnail.Link, thumbnail.Processed, thumbnail.ArticleID)
	if err != nil {
		return errors.Wrap(err, "executing thumbnail update stmt")
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "committing transaction")
		}

		return nil
	}

	stmt, err = tx.Preparex(s.Thumbnail.Create)
	if err != nil {
		return errors.Wrap(err, "preparing thumbnail create stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(thumbnail.ArticleID, thumbnail.Thumbnail, thumbnail.Link, thumbnail.Processed)
	if err != nil {
		return errors.Wrap(err, "executing thumbnail create stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}
