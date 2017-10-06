package sql

import (
	"database/sql"

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

	var extract content.Extract
	if err := r.db.Get(&extract, r.db.SQL().Extract.Get, article.ID); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.Extract{}, errors.Wrapf(err, "getting extract for article %s", article)
	}

	extract.ArticleID = article.ID

	return extract, nil
}

func (r extractRepo) Update(extract content.Extract) error {
	if err := extract.Validate(); err != nil {
		return errors.WithMessage(err, "validating extract")
	}

	r.log.Infof("Updating extract %s", extract)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.Preparex(s.Extract.Update)
	if err != nil {
		return errors.Wrap(err, "preparing extract update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(extract.Title, extract.Content, extract.TopImage, extract.Language, extract.ArticleID)
	if err != nil {
		return errors.Wrap(err, "executing extract update stmt")
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "committing transaction")
		}

		return nil
	}

	stmt, err = tx.Preparex(s.Extract.Create)
	if err != nil {
		return errors.Wrap(err, "preparing extract create stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(extract.ArticleID, extract.Title, extract.Content, extract.TopImage, extract.Language)
	if err != nil {
		return errors.Wrap(err, "executing extract create stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}
