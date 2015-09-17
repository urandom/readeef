package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type ArticleThumbnail struct {
	base.ArticleThumbnail
	logger webfw.Logger

	db *db.DB
}

func (at *ArticleThumbnail) Update() {
	if at.HasErr() {
		return
	}

	if err := at.Validate(); err != nil {
		at.Err(err)
		return
	}

	data := at.Data()
	s := at.db.SQL()
	at.logger.Infof("Updating thumbnail for article %d", data.ArticleId)

	tx, err := at.db.Beginx()
	if err != nil {
		at.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(s.Article.UpdateThumbnail)
	if err != nil {
		at.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(data.Thumbnail, data.Link, data.Processed, data.ArticleId)
	if err != nil {
		at.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		stmt, err := tx.Preparex(s.Article.CreateThumbnail)
		if err != nil {
			at.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(data.ArticleId, data.Thumbnail, data.Link, data.Processed)
		if err != nil {
			at.Err(err)
			return
		}
	}

	tx.Commit()
}
