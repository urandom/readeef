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
	*base.Article
}

type UserArticle struct {
	*base.UserArticle
	Article
	logger webfw.Logger

	db *db.DB
}

type ScoredArticle struct {
	*UserArticle
}

func NewArticle() *Article {
	return &Article{}
}

func NewUserArticle(db *db.DB, logger webfw.Logger, user content.User) *UserArticle {
	return &UserArticle{UserArticle: base.NewUserArticle(user), db: db, logger: logger}
}

func NewScoredArticle(db *db.DB, logger webfw.Logger, user content.User) *ScoredArticle {
	return &ScoredArticle{UserArticle: NewUserArticle(db, logger, user)}
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
		ua.SetErr(err)
		return
	}
	defer tx.Rollback()

	var stmt *sqlx.Stmt
	if read {
		stmt, err = tx.Preparex(db.SQL("create_user_article_read"))
	} else {
		stmt, err = tx.Preparex(db.SQL("delete_user_article_read"))
	}

	if err != nil {
		ua.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id)
	ua.SetErr(err)
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
		ua.SetErr(err)
		return
	}
	defer tx.Rollback()

	var stmt *sqlx.Stmt
	if favorite {
		stmt, err = tx.Preparex(db.SQL("create_user_article_favorite"))
	} else {
		stmt, err = tx.Preparex(db.SQL("delete_user_article_favorite"))
	}

	if err != nil {
		ua.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id)
	ua.SetErr(err)
}

func (sa *ScoredArticle) SetScores(asc info.ArticleScores) {
	if sa.Err() != nil {
		return
	}

	id := sa.Info().Id
	sa.logger.Infof("Setting article '%d' scores\n", id)

	tx, err := sa.db.Begin()
	if err != nil {
		sa.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("update_article_scores"))
	if err != nil {
		sa.SetErr(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(asc.Score, asc.Score1, asc.Score2, asc.Score3, asc.Score4, asc.Score5, id)
	if err != nil {
		sa.SetErr(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		stmt, err := tx.Preparex(db.SQL("create_article_scores"))
		if err != nil {
			sa.SetErr(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(id, asc.Score, asc.Score1, asc.Score2, asc.Score3, asc.Score4, asc.Score5)
		if err != nil {
			sa.SetErr(err)
			return
		}
	}

	tx.Commit()
}

func (sa *ScoredArticle) Scores() (i info.ArticleScores) {
	if sa.Err() != nil {
		return
	}

	id := sa.Info().Id
	sa.logger.Infof("Getting article '%d' scores\n", id)

	if err := sa.db.Get(&i, db.SQL("get_article_scores"), id); err != nil && err != sql.ErrNoRows {
		sa.SetErr(err)
	}

	return
}

func init() {
	db.SetSQL("create_user_article_read", createUserArticleRead)
	db.SetSQL("delete_user_article_read", deleteUserArticleRead)
	db.SetSQL("create_user_article_favorite", createUserArticleFavorite)
	db.SetSQL("delete_user_article_favorite", deleteUserArticleFavorite)
	db.SetSQL("get_article_scores", getArticleScores)
	db.SetSQL("create_article_scores", createArticleScores)
	db.SetSQL("update_article_scores", updateArticleScores)
}

const (
	createUserArticleRead = `
INSERT INTO users_articles_read(user_login, article_id)
	SELECT $1, $2 EXCEPT
		SELECT user_login, article_id
		FROM users_articles_read WHERE user_login = $1 AND article_id = $2
`
	deleteUserArticleRead = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id = $2`
	createUserArticleFavorite = `
INSERT INTO users_articles_fav(user_login, article_id)
	SELECT $1, $2 EXCEPT
		SELECT user_login, article_id
		FROM users_articles_fav WHERE user_login = $1 AND article_id = $2
`
	deleteUserArticleFavorite = `
DELETE FROM users_articles_fav WHERE  user_login = $1 AND article_id = $2
`
	getArticleScores = `
SELECT asco.score, asco.score1, asco.score2, asco.score3, asco.score4, asco.score5
FROM articles_scores asco
WHERE asco.article_id = $1
`
	createArticleScores = `
INSERT INTO articles_scores(article_id, score, score1, score2, score3, score4, score5)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT article_id, score, score1, score2, score3, score4, score5 FROM articles_scores WHERE article_id = $1`
	updateArticleScores = `UPDATE articles_scores SET score = $1, score1 = $2, score2 = $3, score3 = $4, score4 = $5, score5 = $6 WHERE article_id = $7`
)
