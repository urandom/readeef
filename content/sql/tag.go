package sql

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type Tag struct {
	base.Tag
	logger webfw.Logger

	db *db.DB
}

type feedIdTag struct {
	FeedId   data.FeedId   `db:"feed_id"`
	TagValue data.TagValue `db:"tag"`
}

func (t *Tag) AllFeeds() (tf []content.TaggedFeed) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	t.logger.Infof("Getting all feeds for tag %s\n", t)
	u := t.User()

	var d []data.Feed
	if err := t.db.Select(&d, t.db.SQL().Tag.GetUserFeeds, u.Data().Login, t.String()); err != nil {
		t.Err(err)
		return
	}

	repo := t.Repo()
	tf = make([]content.TaggedFeed, len(d))
	for i := range d {
		tf[i] = repo.TaggedFeed(u)
		tf[i].Data(d[i])
	}

	return
}

func (t *Tag) Articles(o ...data.ArticleQueryOptions) (ua []content.UserArticle) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	var opts data.ArticleQueryOptions
	if len(o) > 0 {
		opts = o[0]
	}

	t.logger.Infof("Getting articles for tag %s with options: %#v\n", t, opts)
	u := t.User()

	ua = getArticles(u, t.db, t.logger, opts, t,
		t.db.SQL().Tag.GetArticlesJoin, "", []interface{}{t.String()})

	if u.HasErr() {
		t.Err(u.Err())
	}

	return
}

func (t *Tag) Count(o ...data.ArticleCountOptions) (count int64) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	var opts data.ArticleCountOptions
	if len(o) > 0 {
		opts = o[0]
	}

	u := t.User()
	tag := t.Value()
	t.logger.Infof("Getting user %s tag %s count with options %#v\n", u.Data().Login, tag, opts)

	count = articleCount(u, t.db, t.logger, opts, t.db.SQL().Tag.ArticleCountJoin, "", []interface{}{tag})

	if u.HasErr() {
		t.Err(u.Err())
	}

	return
}

func (t *Tag) ReadState(read bool, o ...data.ArticleUpdateStateOptions) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	var opts data.ArticleUpdateStateOptions
	if len(o) > 0 {
		opts = o[0]
	}

	u := t.User()
	login := u.Data().Login
	tag := t.Value()
	t.logger.Infof("Getting articles for user %s tag %s with options: %#v\n", login, tag, opts)

	s := t.db.SQL()
	args := []interface{}{tag}
	readState(u, t.db, t.logger, opts, read,
		s.Tag.ReadStateInsertJoin, "",
		s.Tag.ReadStateDeleteJoin, "",
		args, args)

	if u.HasErr() {
		t.Err(u.Err())
	}
}

func (t *Tag) Query(term string, sp content.SearchProvider, paging ...int) (ua []content.UserArticle) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	var err error

	feeds := t.AllFeeds()
	if t.HasErr() {
		return
	}

	ids := make([]data.FeedId, len(feeds))
	for i := range feeds {
		ids = append(ids, feeds[i].Data().Id)
	}

	limit, offset := pagingLimit(paging)
	ua, err = sp.Search(term, t.User(), ids, limit, offset)
	t.Err(err)

	return
}
