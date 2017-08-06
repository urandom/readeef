package sql

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/log"
)

type extractRepo struct {
	db *db.DB

	log log.Log
}

func (r extractRepo) Get(id content.ArticleID) (content.Extract, error) {
	r.log.Infof("Getting extract for article id %d", id)

	var extract content.Extract
	err := r.db.Get(&extract, r.db.SQL().Article.GetExtract, id)
	if err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.User{}, errors.Wrapf(err, "getting extract for article id %d", id)
	}

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
	stmt, err := tx.Preparex(s.Article.UpdateExtract)
	if err != nil {
		return errors.Wrap(err, "preparing extract update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(extract.Title, extract.Content, extract.TopImage, extract.Language, extract.ArticleId)
	if err != nil {
		return errors.Wrap(err, "executimg extract update stmt")
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "committing transaction")
		}

		return nil
	}

	stmt, err = tx.Preparex(s.Article.CreateExtract)
	if err != nil {
		return errors.Wrap(err, "preparing extract create stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(extract.ArticleID, extract.Title, extract.Content, extract.TopImage, extract.Language)
	if err != nil {
		return errors.Wrap(err, "executimg extract create stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}
