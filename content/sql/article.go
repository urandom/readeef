package sql

import (
	"github.com/urandom/readeef/content"
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
	NamedSQL
	logger webfw.Logger

	db *db.DB
}

type ScoredArticle struct {
	*UserArticle
}

func NewArticle() *Article {
	return &Article{}
}

func NewUserArticle(db *db.DB, logger webfw.Logger) *UserArticle {
	ua := &UserArticle{NamedSQL: NewNamedSQL(), db: db, logger: logger}

	ua.init()

	return ua
}

func NewScoredArticle(db *db.DB, logger webfw.Logger) *ScoredArticle {
	sa := &ScoredArticle{UserArticle: NewUserArticle(db, logger)}

	sa.init()

	return sa
}

func (ua *UserArticle) Read(read bool) content.UserArticle {
	if ua.Err() != nil {
		return ua
	}

	return ua
}

func (ua *UserArticle) Favorite(favorite bool) content.UserArticle {
	if ua.Err() != nil {
		return ua
	}

	return ua
}

func (sa *ScoredArticle) SetScores(asc info.ArticleScores) content.ScoredArticle {
	if sa.Err() != nil {
		return sa
	}

	return sa
}

func (sa *ScoredArticle) Scores() (i info.ArticleScores) {
	if sa.Err() != nil {
		return
	}

	return
}

func (ua *UserArticle) init() {
	// ua.SQL()
}

func (sa *ScoredArticle) init() {
	// sa.UserArticle.SQL()
}
