package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type ArticleScores struct {
	base.ArticleScores
	logger webfw.Logger

	db *db.DB
}

func (asc *ArticleScores) Update() {
	if asc.HasErr() {
		return
	}

	if err := asc.Validate(); err != nil {
		asc.Err(err)
		return
	}

	data := asc.Data()
	s := asc.db.SQL()
	if data.Score == 0 {
		data.Score = asc.Calculate()
		asc.Data(data)
	}
	asc.logger.Infof("Updating scores for article %d", data.ArticleId)

	tx, err := asc.db.Beginx()
	if err != nil {
		asc.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(s.Article.UpdateScores)
	if err != nil {
		asc.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(data.Score, data.Score1, data.Score2, data.Score3, data.Score4, data.Score5, data.ArticleId)
	if err != nil {
		asc.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		stmt, err := tx.Preparex(s.Article.CreateScores)
		if err != nil {
			asc.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(data.ArticleId, data.Score, data.Score1, data.Score2, data.Score3, data.Score4, data.Score5)
		if err != nil {
			asc.Err(err)
			return
		}
	}

	tx.Commit()
}
