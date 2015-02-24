package sql

import (
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

func (ua *UserArticle) Read(read bool) {
	if ua.Err() != nil {
		return
	}
}

func (ua *UserArticle) Favorite(favorite bool) {
	if ua.Err() != nil {
		return
	}
}

func (sa *ScoredArticle) SetScores(asc info.ArticleScores) {
	if sa.Err() != nil {
		return
	}
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
