package sql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type Article struct {
	base.Article

	logger webfw.Logger
	db     *db.DB
}

type UserArticle struct {
	base.UserArticle
	Article
}

func (a *Article) Update() {
	a.logger.Infof("Updating article %d\n", a.Data().Id)

	tx, err := a.db.Beginx()
	if err != nil {
		a.Err(err)
		return
	}
	defer tx.Rollback()

	updateArticle(a, tx, a.db, a.logger)

	if !a.HasErr() {
		tx.Commit()
	}
}

func (a *Article) Thumbnail() (at content.ArticleThumbnail) {
	at = a.Repo().ArticleThumbnail()
	if a.HasErr() {
		at.Err(a.Err())
		return
	}

	id := a.Data().Id
	if id == 0 {
		a.Err(content.NewValidationError(errors.New("Invalid article id")))
		return
	}

	a.logger.Infof("Getting article '%d' thumbnail\n", id)

	var i data.ArticleThumbnail
	if err := a.db.Get(&i, a.db.SQL().Article.GetThumbnail, id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		at.Err(err)
	}

	i.ArticleId = id
	at.Data(i)

	return
}

func (a *Article) Extract() (ae content.ArticleExtract) {
	ae = a.Repo().ArticleExtract()
	if a.HasErr() {
		ae.Err(a.Err())
		return
	}

	id := a.Data().Id
	if id == 0 {
		a.Err(content.NewValidationError(errors.New("Invalid article id")))
		return
	}

	a.logger.Infof("Getting article '%d' extract\n", id)

	var i data.ArticleExtract
	if err := a.db.Get(&i, a.db.SQL().Article.GetExtract, id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		ae.Err(err)
	}

	i.ArticleId = id
	ae.Data(i)

	return
}

func (a *Article) Scores() (as content.ArticleScores) {
	as = a.Repo().ArticleScores()
	if a.HasErr() {
		as.Err(a.Err())
		return
	}

	id := a.Data().Id
	if id == 0 {
		a.Err(content.NewValidationError(errors.New("Invalid article id")))
		return
	}

	a.logger.Infof("Getting article '%d' scores\n", id)

	var i data.ArticleScores
	if err := a.db.Get(&i, a.db.SQL().Article.GetScores, id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		as.Err(err)
	}

	i.ArticleId = id
	as.Data(i)

	return
}

func updateArticle(a content.Article, tx *sqlx.Tx, db *db.DB, logger webfw.Logger) {
	if a.HasErr() {
		return
	}

	if err := a.Validate(); err != nil {
		a.Err(err)
		return
	}

	logger.Infof("Updating article %s\n", a)

	d := a.Data()
	s := db.SQL()

	stmt, err := tx.Preparex(s.Article.Update)
	if err != nil {
		a.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(d.Title, d.Description, d.Date, d.Guid, d.Link, d.FeedId)
	if err != nil {
		a.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil && err == sql.ErrNoRows || num == 0 {
		logger.Infof("Creating article %s\n", a)

		aId, err := db.CreateWithId(tx, s.Article.Create, d.FeedId, d.Link, d.Guid,
			d.Title, d.Description, d.Date)

		if err != nil {
			a.Err(fmt.Errorf("Error updating article %s (guid - %v, link - %s): %v", a, d.Guid, d.Link, err))
			return
		}

		d.Id = data.ArticleId(aId)
		d.IsNew = true
		a.Data(d)
	}
}

func (ua *UserArticle) Read(read bool) {
	if ua.HasErr() {
		return
	}

	if err := ua.Validate(); err != nil {
		ua.Err(err)
		return
	}

	login := ua.User().Data().Login
	ua.logger.Infof("Setting read state %v on %s for %s\n", read, ua, login)

	d := ua.Data()

	tx, err := ua.db.Beginx()
	if err != nil {
		ua.Err(err)
		return
	}
	defer tx.Rollback()

	if read {
		stmt, err := tx.Preparex(ua.db.SQL().Article.DeleteUserUnread)
		if err != nil {
			ua.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, d.Id)
		if err != nil {
			ua.Err(err)
			return
		}
	} else {
		stmt, err := tx.Preparex(ua.db.SQL().Article.CreateUserUnread)
		if err != nil {
			ua.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, d.Id)
		if err != nil {
			ua.Err(err)
			return
		}
	}

	d.Read = read
	ua.Data(d)

	if err := tx.Commit(); err != nil {
		ua.Err(err)
	}
}

func (ua *UserArticle) Favorite(favorite bool) {
	if ua.HasErr() {
		return
	}

	if err := ua.Validate(); err != nil {
		ua.Err(err)
		return
	}

	login := ua.User().Data().Login
	ua.logger.Infof("Setting favorite state %v on %s for %s\n", favorite, ua, login)

	d := ua.Data()

	tx, err := ua.db.Beginx()
	if err != nil {
		ua.Err(err)
		return
	}
	defer tx.Rollback()

	if favorite {
		stmt, err := tx.Preparex(ua.db.SQL().Article.CreateUserFavorite)
		if err != nil {
			ua.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, d.Id)
		if err != nil {
			ua.Err(err)
			return
		}
	} else {
		stmt, err := tx.Preparex(ua.db.SQL().Article.DeleteUserFavorite)
		if err != nil {
			ua.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, d.Id)
		if err != nil {
			ua.Err(err)
			return
		}
	}

	d.Favorite = favorite
	ua.Data(d)

	if err := tx.Commit(); err != nil {
		ua.Err(err)
	}
}
