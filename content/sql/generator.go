package sql

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
)

func (r *Repo) Article() content.Article {
	a := r.article()

	a.Repo(r)

	return &a
}

func (r *Repo) UserArticle(u content.User) content.UserArticle {
	ua := r.userArticle(u)

	ua.Repo(r)

	return &ua
}

func (r *Repo) ArticleScores() content.ArticleScores {
	asc := &ArticleScores{db: r.db, log: r.log}

	asc.Repo(r)

	return asc
}

func (r *Repo) ArticleThumbnail() content.ArticleThumbnail {
	at := &ArticleThumbnail{db: r.db, log: r.log}

	at.Repo(r)

	return at
}

func (r *Repo) ArticleExtract() content.ArticleExtract {
	ae := &ArticleExtract{db: r.db, log: r.log}

	ae.Repo(r)

	return ae
}

func (r *Repo) Feed() content.Feed {
	f := r.feed()

	f.Repo(r)

	return &f
}

func (r *Repo) UserFeed(u content.User) content.UserFeed {
	uf := r.userFeed(u)

	uf.Repo(r)

	return &uf
}

func (r *Repo) TaggedFeed(u content.User) content.TaggedFeed {
	tf := &TaggedFeed{UserFeed: r.userFeed(u)}

	tf.Repo(r)

	return tf
}

func (r *Repo) Subscription() content.Subscription {
	s := &Subscription{db: r.db, log: r.log}

	s.Repo(r)

	return s
}

func (r *Repo) Tag(u content.User) content.Tag {
	t := &base.Tag{}
	t.User(u)
	t.Repo(r)
	return &Tag{Tag: *t, db: r.db, log: r.log}
}

func (r *Repo) User() content.User {
	u := &User{db: r.db, log: r.log}

	u.Repo(r)

	return u
}

func (r *Repo) userArticle(u content.User) UserArticle {
	ua := &base.UserArticle{}
	ua.User(u)

	return UserArticle{Article: r.article(), UserArticle: *ua}
}

func (r Repo) article() Article {
	return Article{db: r.db, log: r.log}
}

func (r Repo) feed() Feed {
	return Feed{db: r.db, log: r.log}
}

func (r Repo) userFeed(u content.User) UserFeed {
	uf := &base.UserFeed{}
	uf.User(u)
	return UserFeed{Feed: r.feed(), UserFeed: *uf}
}
