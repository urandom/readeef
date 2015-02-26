package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Article struct {
	base.Article
}

type UserArticle struct {
	base.UserArticle
	Article
	logger webfw.Logger

	db *db.DB
}

type ScoredArticle struct {
	UserArticle
}

func (ua *UserArticle) Read(read bool) {
	if ua.Err() != nil {
		return
	}

	id := ua.Info().Id
	login := ua.User().Info().Login
	ua.logger.Infof("Marking user '%s' article '%d' as read: %v\n", login, id, read)

	tx, err := ua.db.Begin()
	if err != nil {
		ua.Err(err)
		return
	}
	defer tx.Rollback()

	var stmt *sqlx.Stmt
	if read {
		stmt, err = tx.Preparex(ua.db.SQL("create_user_article_read"))
	} else {
		stmt, err = tx.Preparex(ua.db.SQL("delete_user_article_read"))
	}

	if err != nil {
		ua.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id)
	ua.Err(err)
}

func (ua *UserArticle) Favorite(favorite bool) {
	if ua.Err() != nil {
		return
	}

	id := ua.Info().Id
	login := ua.User().Info().Login
	ua.logger.Infof("Marking user '%s' article '%d' as favorite: %v\n", login, id, favorite)

	tx, err := ua.db.Begin()
	if err != nil {
		ua.Err(err)
		return
	}
	defer tx.Rollback()

	var stmt *sqlx.Stmt
	if favorite {
		stmt, err = tx.Preparex(ua.db.SQL("create_user_article_favorite"))
	} else {
		stmt, err = tx.Preparex(ua.db.SQL("delete_user_article_favorite"))
	}

	if err != nil {
		ua.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id)
	ua.Err(err)
}

func (sa *ScoredArticle) Scores(in ...info.ArticleScores) (i info.ArticleScores) {
	if sa.Err() != nil {
		return
	}

	id := sa.Info().Id
	if len(in) > 0 {
		asc := in[0]
		sa.logger.Infof("Setting article '%d' scores\n", id)

		tx, err := sa.db.Begin()
		if err != nil {
			sa.Err(err)
			return
		}
		defer tx.Rollback()

		stmt, err := tx.Preparex(sa.db.SQL("update_article_scores"))
		if err != nil {
			sa.Err(err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(asc.Score, asc.Score1, asc.Score2, asc.Score3, asc.Score4, asc.Score5, id)
		if err != nil {
			sa.Err(err)
			return
		}

		if num, err := res.RowsAffected(); err != nil || num == 0 {
			stmt, err := tx.Preparex(sa.db.SQL("create_article_scores"))
			if err != nil {
				sa.Err(err)
				return
			}
			defer stmt.Close()

			_, err = stmt.Exec(id, asc.Score, asc.Score1, asc.Score2, asc.Score3, asc.Score4, asc.Score5)
			if err != nil {
				sa.Err(err)
				return
			}
		}

		tx.Commit()
	} else {
		sa.logger.Infof("Getting article '%d' scores\n", id)

		if err := sa.db.Get(&i, sa.db.SQL("get_article_scores"), id); err != nil && err != sql.ErrNoRows {
			sa.Err(err)
		}
	}

	return
}
