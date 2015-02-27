package sql

import (
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type ArticleScores struct {
	base.ArticleScores
	logger webfw.Logger

	db *db.DB
}

func (asc *ArticleScores) Update() {
	if asc.Err() != nil {
		return
	}

	info := asc.Info()
	asc.logger.Infof("Updating scores for article", info.ArticleId)

	tx, err := asc.db.Begin()
	if err != nil {
		asc.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(asc.db.SQL("update_article_scores"))
	if err != nil {
		asc.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(info.Score, info.Score1, info.Score2, info.Score3, info.Score4, info.Score5, info.ArticleId)
	if err != nil {
		asc.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		stmt, err := tx.Preparex(asc.db.SQL("create_article_scores"))
		if err != nil {
			asc.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(info.ArticleId, info.Score, info.Score1, info.Score2, info.Score3, info.Score4, info.Score5)
		if err != nil {
			asc.Err(err)
			return
		}
	}

	tx.Commit()
}
