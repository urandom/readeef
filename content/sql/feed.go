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
	Feed
}

type TaggedFeed struct {
	base.TaggedFeed
	UserFeed
}

func NewFeed(db *db.DB, logger webfw.Logger) *Feed {
	return &Feed{db: db, logger: logger}
}

func NewUserFeed(db *db.DB, logger webfw.Logger, user content.User) *UserFeed {
	return &UserFeed{Feed: Feed{db: db, logger: logger}, UserFeed: base.NewUserFeed(user)}
}

func NewTaggedFeed(db *db.DB, logger webfw.Logger, user content.User) *TaggedFeed {
	return &TaggedFeed{UserFeed: *NewUserFeed(db, logger, user)}
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

	var info []info.Article
	if err := f.db.Select(&info, db.SQL("get_all_feed_articles"), id); err != nil {
		f.SetErr(err)
		return
	}

	a = make([]content.Article, len(info))
	for i := range info {
		a[i] = NewArticle()
		a[i].Set(info[i])
	}

	return
}

func (f *Feed) LatestArticles() (a []content.Article) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting latest feed %d articles\n", id)

	var info []info.Article
	if err := f.db.Select(&info, db.SQL("get_latest_feed_articles"), id); err != nil {
		f.SetErr(err)
		return
	}

	a = make([]content.Article, len(info))
	for i := range info {
		a[i] = NewArticle()
		a[i].Set(info[i])
	}

	return
}

func (f *Feed) AddArticles(articles []content.Article) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Adding articles to feed %d\n", id)

	tx, err := f.db.Begin()
	if err != nil {
		f.SetErr(err)
		return
	}
	defer tx.Rollback()

	newArticles := f.updateFeedArticles(tx, articles)

	if f.Err() != nil {
		return
	}

	tx.Commit()

	f.newArticles = newArticles
}

func (f *Feed) Subscription() (s content.Subscription) {
	if f.Err() != nil {
		return
	}

	id := f.Info().Id
	f.logger.Infof("Getting subcription for feed %d\n", id)

	var in info.Subscription
	if err := f.db.Get(&in, db.SQL("get_hubbub_subscription"), id); err != nil && err != dsql.ErrNoRows {
		f.SetErr(err)
		return
	}

	s = NewSubscription(f.db, f.logger)
	s.Set(in)

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

		id := f.Info().Id
		in := articles[i].Info()
		in.FeedId = id

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

			if err != nil {
				f.SetErr(err)
				return
			}

			in.Id = info.ArticleId(id)
			articles[i].Set(in)
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

	var in []info.User
	if err := uf.db.Select(&in, db.SQL("get_feed_users"), id); err != nil {
		uf.SetErr(err)
		return
	}

	u = make([]content.User, len(in))
	for i := range in {
		u[i] = NewUser(uf.db, uf.logger)
		u[i].Set(in[i])

		if u[i].Err() != nil {
			uf.SetErr(u[i].Err())
			return
		}
	}

	return
}

func (uf *UserFeed) Detach() {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	login := uf.User().Info().Login
	uf.logger.Infof("Detaching feed %d from user %s\n", id, login)

	tx, err := uf.db.Begin()
	if err != nil {
		uf.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_user_feed"))
	if err != nil {
		uf.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id)
	if err != nil {
		uf.SetErr(err)
		return
	}

	tx.Commit()
}

func (uf *UserFeed) Articles(paging ...int) (ua []content.UserArticle) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Getting articles for feed %d\n", id)

	order := "read"
	articles := uf.getArticles("", order, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (uf *UserFeed) UnreadArticles(paging ...int) (ua []content.UserArticle) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	uf.logger.Infof("Getting unread articles for feed %d\n", id)

	articles := uf.getArticles("ar.article_id IS NULL", "", paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (uf *UserFeed) ReadBefore(date time.Time, read bool) {
	if uf.Err() != nil {
		return
	}

	id := uf.Info().Id
	login := uf.User().Info().Login
	uf.logger.Infof("Marking user %s feed %d articles before %v as read: %v\n", login, id, date, read)

	tx, err := uf.db.Begin()
	if err != nil {
		uf.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_all_users_articles_read_by_feed_date"))
	if err != nil {
		uf.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, id, date)
	if err != nil {
		uf.SetErr(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(db.SQL("create_all_users_articles_read_by_feed_date"))
		if err != nil {
			uf.SetErr(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, id, date)
		if err != nil {
			uf.SetErr(err)
			return
		}
	}

	tx.Commit()
}

func (uf *UserFeed) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if uf.Err() != nil {
		return
	}

	u := uf.User()
	id := uf.Info().Id
	login := u.Info().Login
	uf.logger.Infof("Getting scored articles for user %s feed %d between %v and %v\n", login, id, from, to)

	order := ""
	if uf.Order() == info.DescendingOrder {
		order = " DESC"
	}
	ua := getArticles(u, uf.db, uf.logger, uf, "asco.score",
		"INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"uf.feed_id = $2 AND a.date > $3 AND a.date <= $4", "asco.score"+order,
		[]interface{}{id, from, to}, paging...)

	sa = make([]content.ScoredArticle, len(ua))
	for i := range ua {
		sa[i] = &ScoredArticle{UserArticle: *ua[i]}
	}

	return
}

func (uf *UserFeed) getArticles(where, order string, paging ...int) (ua []*UserArticle) {
	if uf.Err() != nil {
		return
	}

	if where == "" {
		where = "uf.feed_id = $2"
	} else {
		where = "uf.feed_id = $2 AND " + where
	}

	u := uf.User()
	ua = getArticles(u, uf.db, uf.logger, uf, "", "", where, order, []interface{}{uf.Info().Id}, paging...)

	if u.Err() != nil {
		uf.SetErr(u.Err())
	}

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
	db.SetSQL("get_all_feed_articles", getAllFeedArticles)
	db.SetSQL("get_latest_feed_articles", getLatestFeedArticles)
	db.SetSQL("get_hubbub_subscription", getHubbubSubscription)
	db.SetSQL("get_feed_users", getFeedUsers)
	db.SetSQL("delete_user_feed", deleteUserFeed)
	db.SetSQL("create_all_users_articles_read_by_feed_date", createAllUsersArticlesReadByFeedDate)
	db.SetSQL("delete_all_users_articles_read_by_feed_date", deleteAllUsersArticlesReadByFeedDate)
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
	getAllFeedArticles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.guid, a.date
FROM articles a
WHERE a.feed_id = $1
`
	getLatestFeedArticles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > NOW() - INTERVAL '5 days'
`
	getHubbubSubscription = `
SELECT link, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions WHERE feed_id = $1`
	getFeedUsers = `
SELECT u.login, u.first_name, u.last_name, u.email, u.admin, u.active,
	u.profile_data, u.hash_type, u.salt, u.hash, u.md5_api
FROM users u, users_feeds uf
WHERE u.login = uf.user_login AND uf.feed_id = $1
`
	deleteUserFeed                       = `DELETE FROM users_feeds WHERE user_login = $1 AND feed_id = $2`
	createAllUsersArticlesReadByFeedDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1 AND uf.feed_id = $2
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`

	deleteAllUsersArticlesReadByFeedDate = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE feed_id = $2 AND (date IS NULL OR date < $3)
)
`
)
