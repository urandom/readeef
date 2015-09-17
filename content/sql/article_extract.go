package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type ArticleExtract struct {
	base.ArticleExtract
	logger webfw.Logger

	db *db.DB
}

func (ae *ArticleExtract) Update() {
	if ae.HasErr() {
		return
	}

	if err := ae.Validate(); err != nil {
		ae.Err(err)
		return
	}

	data := ae.Data()
	s := ae.db.SQL()
	ae.logger.Infof("Updating extract for article %d", data.ArticleId)

	tx, err := ae.db.Beginx()
	if err != nil {
		ae.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(s.Article.UpdateExtract)
	if err != nil {
		ae.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(data.Title, data.Content, data.TopImage, data.Language, data.ArticleId)
	if err != nil {
		ae.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		stmt, err := tx.Preparex(s.Article.CreateExtract)
		if err != nil {
			ae.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(data.ArticleId, data.Title, data.Content, data.TopImage, data.Language)
		if err != nil {
			ae.Err(err)
			return
		}
	}

	tx.Commit()
}
