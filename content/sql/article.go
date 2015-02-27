package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Article struct {
	base.Article
}

type ScoredArticle struct {
	Article
	logger webfw.Logger

	db *db.DB
}

type UserArticle struct {
	base.UserArticle
	Article
	logger webfw.Logger

	db *db.DB
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

func (sa *ScoredArticle) Scores() (asc content.ArticleScores) {
	if sa.Err() != nil {
		return
	}

	id := sa.Info().Id
	sa.logger.Infof("Getting article '%d' scores\n", id)

	var i info.ArticleScores
	if err := sa.db.Get(&i, sa.db.SQL("get_article_scores"), id); err != nil && err != sql.ErrNoRows {
		sa.Err(err)
	}

	asc = sa.Repo().ArticleScores()
	asc.Info(i)

	return
}
