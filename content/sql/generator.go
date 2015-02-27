package sql

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
)

func (r Repo) Article() content.Article {
	return &Article{}
}

func (r Repo) ScoredArticle() content.ScoredArticle {
	return &ScoredArticle{Article: Article{}}
}

func (r Repo) UserArticle(u content.User) content.UserArticle {
	ua := r.userArticle(u)
	return &ua
}

func (r Repo) ArticleScores() content.ArticleScores {
	return &ArticleScores{db: r.db, logger: r.logger}
}

func (r Repo) Feed() content.Feed {
	f := r.feed()
	return &f
}

func (r Repo) UserFeed(u content.User) content.UserFeed {
	uf := r.userFeed(u)
	return &uf
}

func (r Repo) TaggedFeed(u content.User) content.TaggedFeed {
	return &TaggedFeed{UserFeed: r.userFeed(u)}
}

func (r Repo) Subscription() content.Subscription {
	return &Subscription{db: r.db, logger: r.logger}
}

func (r Repo) Tag(u content.User) content.Tag {
	t := &base.Tag{}
	t.User(u)
	return &Tag{Tag: *t, db: r.db, logger: r.logger}
}

func (r Repo) User() content.User {
	return &User{db: r.db, logger: r.logger}
}

func (r Repo) userArticle(u content.User) UserArticle {
	ua := &base.UserArticle{}
	ua.User(u)

	return UserArticle{UserArticle: *ua, db: r.db, logger: r.logger}
}

func (r Repo) feed() Feed {
	return Feed{db: r.db, logger: r.logger}
}

func (r Repo) userFeed(u content.User) UserFeed {
	uf := &base.UserFeed{}
	uf.User(u)
	return UserFeed{Feed: r.feed(), UserFeed: *uf}
}
