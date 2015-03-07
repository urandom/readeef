package sql

import (
	"time"

	"github.com/blevesearch/bleve"
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
	if err := t.db.Select(&d, t.db.SQL("get_user_tag_feeds"), u.Data().Login, t.String()); err != nil {
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

func (t *Tag) Articles(paging ...int) (ua []content.UserArticle) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	t.logger.Infof("Getting articles for tag %s\n", t)

	articles := getArticles(t.User(), t.db, t.logger, t,
		"", "INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login",
		"uft.tag = $2", "read", []interface{}{t.String()}, paging...)

	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (t *Tag) UnreadArticles(paging ...int) (ua []content.UserArticle) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	t.logger.Infof("Getting unread articles for tag %s\n", t)

	articles := getArticles(t.User(), t.db, t.logger, t,
		"", "INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login",
		"uft.tag = $2 AND ar.article_id IS NULL", "", []interface{}{t.String()}, paging...)

	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (t *Tag) ReadBefore(date time.Time, read bool) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	t.logger.Infof("Marking articles for tag %s before %v as read\n", t, date)

	tx, err := t.db.Begin()
	if err != nil {
		t.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(t.db.SQL("delete_all_user_tag_articles_read_by_date"))

	if err != nil {
		t.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.User().Data().Login, t.String(), date)
	if err != nil {
		t.Err(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(t.db.SQL("create_all_user_tag_articles_read_by_date"))

		if err != nil {
			t.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(t.User().Data().Login, t.String(), date)
		if err != nil {
			t.Err(err)
			return
		}
	}

	tx.Commit()
}

func (t *Tag) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	t.logger.Infof("Getting scored articles for tag %s\n", t)

	order := "asco.score"
	if t.Order() == data.DescendingOrder {
		order = "asco.score DESC"
	}

	ua := getArticles(t.User(), t.db, t.logger, t,
		"asco.score", `INNER JOIN articles_scores asco ON a.id = asco.article_id
		INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login`,
		"uft.tag = $2 AND a.date > $3 AND a.date <= $4", order,
		[]interface{}{t.String(), from, to}, paging...)

	sa = make([]content.ScoredArticle, len(ua))
	for i := range ua {
		sa[i] = t.Repo().ScoredArticle()
		sa[i].Data(ua[i].Data())
	}

	return
}

func (t *Tag) Query(term string, index bleve.Index, paging ...int) (ua []content.UserArticle) {
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

	ua, err = query(term, t.Highlight(), index, t.User(), ids, paging...)
	t.Err(err)

	return
}
