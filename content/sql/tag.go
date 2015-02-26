package sql

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Tag struct {
	base.Tag
	logger webfw.Logger

	db *db.DB
}

type feedIdTag struct {
	FeedId   info.FeedId `db:"feed_id"`
	TagValue info.TagValue
}

func NewTag(db *db.DB, logger webfw.Logger, user content.User) *Tag {
	return &Tag{Tag: base.NewTag(user), db: db, logger: logger}
}

func (t *Tag) AllFeeds() (tf []content.TaggedFeed) {
	if t.Err() != nil {
		return
	}

	t.logger.Infof("Getting all feeds for tag %s\n", t)

	var i []info.Feed
	if err := t.db.Select(&i, db.SQL("get_user_tag_feeds"), t.User().Info().Login, t.String()); err != nil {
		t.Err(err)
		return
	}

	return
}

func (t *Tag) Articles(paging ...int) (ua []content.UserArticle) {
	if t.Err() != nil {
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
	if t.Err() != nil {
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
	if t.Err() != nil {
		return
	}

	t.logger.Infof("Marking articles for tag %s before %v as read\n", t, date)

	tx, err := t.db.Begin()
	if err != nil {
		t.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_all_user_tag_articles_read_by_date"))

	if err != nil {
		t.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.User().Info().Login, t.String(), date)
	if err != nil {
		t.Err(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(db.SQL("create_all_user_tag_articles_read_by_date"))

		if err != nil {
			t.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(t.User().Info().Login, t.String(), date)
		if err != nil {
			t.Err(err)
			return
		}
	}

	tx.Commit()
}

func (t *Tag) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if t.Err() != nil {
		return
	}

	t.logger.Infof("Getting scored articles for tag %s\n", t)

	order := "asco.score"
	if t.Order() == info.DescendingOrder {
		order = "asco.score DESC"
	}

	ua := getArticles(t.User(), t.db, t.logger, t,
		"asco.score", `INNER JOIN articles_scores asco ON a.id = asco.article_id
		INNER JOIN users_feeds_tags uft ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login`,
		"uft.tag = $2 AND a.date > $3 AND a.date <= $4", order,
		[]interface{}{t.String(), from, to}, paging...)

	sa = make([]content.ScoredArticle, len(ua))
	for i := range ua {
		sa[i] = &ScoredArticle{UserArticle: *ua[i]}
	}

	return
}

func init() {
	db.SetSQL("create_all_user_tag_articles_read_by_date", createAllUserTagArticlesByDate)
	db.SetSQL("delete_all_user_tag_articles_read_by_date", deleteAllUserTagArticlesByDate)
}

const (
	createAllUserTagArticlesByDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id, uf.feed_id
	FROM users_feeds uf INNER JOIN users_feeds_tags uft
		ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
			AND uft.user_login = $1 AND uft.tag = $2
	INNER JOIN articles a
		ON uf.feed_id = a.feed_id
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`

	deleteAllUserTagArticlesByDate = `
DELETE FROM users_articles_read WHERE user_login = $1
	AND article_feed_id IN (
		SELECT feed_id FROM users_feeds_tags WHERE user_login = $1 AND tag = $2
	) AND article_id IN (
		SELECT id FROM articles WHERE date IS NULL OR date < $3
	)
`
)
