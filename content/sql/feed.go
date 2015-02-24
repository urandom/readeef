package sql

import (
	dsql "database/sql"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Feed struct {
	base.Feed
	logger webfw.Logger

	db          *db.DB
	newArticles []content.Article
}

type UserFeed struct {
	base.UserFeed
	*Feed
}

type TaggedFeed struct {
	base.TaggedFeed
	*UserFeed
}

func NewFeed(db *db.DB, logger webfw.Logger) *Feed {
	return &Feed{db: db, logger: logger}
}

func NewUserFeed(db *db.DB, logger webfw.Logger, user content.User) *UserFeed {
	return &UserFeed{Feed: NewFeed(db, logger), UserFeed: base.NewUserFeed(user)}
}

func NewTaggedFeed(db *db.DB, logger webfw.Logger, user content.User) *TaggedFeed {
	return &TaggedFeed{UserFeed: NewUserFeed(db, logger, user)}
}

func (f Feed) NewArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	return f.newArticles
}

func (f *Feed) Update(i info.Feed) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Updating feed %d\n", id)

	tx, err := f.db.Begin()
	if err != nil {
		f.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("update_feed"))
	if err != nil {
		f.SetErr(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.Link, i.Title, i.Description, i.HubLink, i.SiteLink, i.UpdateError, i.SubscribeError, id)
	if err != nil {
		f.SetErr(err)
		return
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		stmt, err := tx.Preparex(db.SQL("create_feed"))
		if err != nil {
			f.SetErr(err)
			return
		}
		defer stmt.Close()

		id, err := f.db.CreateWithId(stmt, i.Link, i.Title, i.Description, i.HubLink, i.SiteLink, i.UpdateError, i.SubscribeError)
		if err != nil {
			f.SetErr(err)
			return
		}

		i.Id = info.FeedId(id)

		f.Set(i)
	}

	articles := f.updateFeedArticles(tx, f.ParsedArticles())

	if f.Err() != nil {
		return
	}

	f.newArticles = articles

	tx.Commit()
}

func (f *Feed) Delete() {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Deleting feed %d\n", id)

	tx, err := f.db.Begin()
	if err != nil {
		f.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_feed"))
	if err != nil {
		f.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		f.SetErr(err)
		return
	}

	tx.Commit()
}

func (f *Feed) AllArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting all feed %d articles\n", id)

	return
}

func (f *Feed) LatestArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting latest feed %d articles\n", id)

	return
}

func (f *Feed) AddArticles([]content.Article) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Adding articles to feed %d\n", id)
}

func (f *Feed) Subscription() (s content.Subscription) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting subcription for feed %d\n", id)

	return
}

func (f *Feed) updateFeedArticles(tx *db.Tx, articles []content.Article) (a []content.Article) {
	if f.Err() != nil || len(articles) == 0 {
		return
	}

	for i := range articles {
		if err := articles[i].Validate(); err != nil {
			f.SetErr(err)
			return
		}

		in := articles[i].Info()
		id := f.Info().Id

		var sql string
		args := []interface{}{in.Title, in.Description, in.Date, id}

		if in.Guid.Valid {
			sql = db.SQL("update_feed_article_with_guid")
			args = append(args, in.Guid)
		} else {
			sql = db.SQL("update_feed_article")
			args = append(args, in.Link)
		}

		stmt, err := tx.Preparex(sql)
		if err != nil {
			f.SetErr(err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(args...)
		if err != nil {
			f.SetErr(err)
			return
		}

		if num, err := res.RowsAffected(); err != nil && err == dsql.ErrNoRows || num == 0 {
			a = append(a, articles[i])

			stmt, err := tx.Preparex(db.SQL("create_feed_article"))
			if err != nil {
				f.SetErr(err)
				return
			}
			defer stmt.Close()

			id, err := f.db.CreateWithId(stmt, id, in.Link, in.Guid,
				in.Title, in.Description, in.Date)
			in.Id = info.ArticleId(id)

			if err != nil {
				f.SetErr(err)
				return
			}
		}
	}

	return
}

func (uf UserFeed) Validate() error {
	err := uf.Feed.Validate()
	if err == nil {
		err = uf.UserFeed.Validate()
	}

	return err
}

func (uf *UserFeed) Users() (u []content.User) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Getting users for feed %d\n", id)

	return
}

func (uf *UserFeed) Detach() {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	login := uf.User().Info().Login
	uf.logger.Infof("Detaching feed %d from user %s\n", id, login)
}

func (uf *UserFeed) Articles(desc bool, paging ...int) (ua []content.UserArticle) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Getting articles for feed %d\n", id)

	return
}

func (uf *UserFeed) UnreadArticles(desc bool, paging ...int) (ua []content.UserArticle) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Getting unread articles for feed %d\n", id)

	return
}

func (uf *UserFeed) ReadBefore(date time.Time, read bool) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Marking feed %d articles before %v as read: %v\n", id, date, read)
}

func (uf *UserFeed) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Getting scored articles for feed %d between %v and %v\n", id, from, to)

	return
}

func (tf *TaggedFeed) Tags() (t []content.Tag) {
	if tf.Err() != nil {
		return
	}

	id := tf.Info().Id
	tf.logger.Infof("Getting tags for feed %d\n", id)

	return
}

func (tf *TaggedFeed) AddTags(tags ...content.Tag) {
	if tf.Err() != nil {
		return
	}

	id := tf.Info().Id
	tf.logger.Infof("Adding tags for feed %d\n", id)
}

func (tf *TaggedFeed) DeleteAllTags() {
	if tf.Err() != nil {
		return
	}

	id := tf.Info().Id
	tf.logger.Infof("Deleting all tags for feed %d\n", id)
}

func init() {
	db.SetSQL("create_feed", createFeed)
	db.SetSQL("update_feed", updateFeed)
	db.SetSQL("delete_feed", deleteFeed)
	db.SetSQL("create_feed_article", createFeedArticle)
	db.SetSQL("update_feed_article", updateFeedArticle)
	db.SetSQL("update_feed_article_with_guid", updateFeedArticleWithGuid)
}

const (
	createFeed = `
INSERT INTO feeds(link, title, description, hub_link, site_link, update_error, subscribe_error)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	updateFeed        = `UPDATE feeds SET link = $1, title = $2, description = $3, hub_link = $4, site_link = $5, update_error = $6, subscribe_error = $7 WHERE id = $8`
	deleteFeed        = `DELETE FROM feeds WHERE id = $1`
	createFeedArticle = `
INSERT INTO articles(feed_id, link, guid, title, description, date)
	SELECT $1, $2, $3, $4, $5, $6 EXCEPT
		SELECT feed_id, link, guid, title, description, date
		FROM articles WHERE feed_id = $1 AND link = $2
`

	updateFeedArticle = `
UPDATE articles SET title = $1, description = $2, date = $3 WHERE feed_id = $4 AND link = $5
`

	updateFeedArticleWithGuid = `
UPDATE articles SET title = $1, description = $2, date = $3 WHERE feed_id = $4 AND guid = $5
`
)
