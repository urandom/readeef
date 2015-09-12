package sql

import (
	"time"

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
		"uft.tag = $2 AND uas.article_id IS NULL OR NOT uas.read", "", []interface{}{t.String()}, paging...)

	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (t *Tag) UnreadCount() (count int64) {
	if t.HasErr() {
		return
	}

	if err := t.Validate(); err != nil {
		t.Err(err)
		return
	}

	login := t.User().Data().Login
	tag := t.Value()
	t.logger.Infof("Getting user %s tag %s unread count\n", login, tag)

	if err := t.db.Get(&count, t.db.SQL("get_tag_unread_count"), login, tag); err != nil {
		t.Err(err)
		return
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

	login := t.User().Data().Login
	t.logger.Infof("Marking user %s articles for tag %s before %v as read\n", login, t, date)

	tx, err := t.db.Beginx()
	if err != nil {
		t.Err(err)
		return
	}
	defer tx.Rollback()

	if read {
		stmt, err := tx.Preparex(t.db.SQL("create_missing_user_article_state_by_tag_date"))

		if err != nil {
			t.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, t.String(), date)
		if err != nil {
			t.Err(err)
			return
		}
	}

	stmt, err := tx.Preparex(t.db.SQL("update_all_user_article_state_by_tag_date"))

	if err != nil {
		t.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(read, login, t.String(), date)
	if err != nil {
		t.Err(err)
		return
	}

	if err = tx.Commit(); err != nil {
		t.Err(err)
	}
}

func (t *Tag) ScoredArticles(from, to time.Time, paging ...int) (ua []content.UserArticle) {
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

	ua = getArticles(t.User(), t.db, t.logger, t,
		"asco.score", `INNER JOIN articles_scores asco ON a.id = asco.article_id
		INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login`,
		"uft.tag = $2 AND a.date > $3 AND a.date <= $4", order,
		[]interface{}{t.String(), from, to}, paging...)

	return
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
