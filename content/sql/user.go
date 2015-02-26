package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type User struct {
	base.User
	logger webfw.Logger

	db *db.DB
}

func NewUser(db *db.DB, logger webfw.Logger) *User {
	u := &User{db: db, logger: logger}

	return u
}

func (u *User) Update() {
	if u.Err() != nil {
		return
	}

	i := u.Info()
	u.logger.Infof("Updating user %s\n", i.Login)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("update_user"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API, i.Login)
	if err != nil {
		u.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			u.Err(err)
		}

		return
	}

	stmt, err = tx.Preparex(db.SQL("create_user"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login, i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API)
	if err != nil {
		u.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.Err(err)
	}

	return
}

func (u *User) Delete() {
	if u.Err() != nil {
		return
	}

	i := u.Info()
	u.logger.Infof("Deleting user %s\n", i.Login)

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_user"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login)
	if err != nil {
		u.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.Err(err)
	}
}

func (u *User) Feed(id info.FeedId) (uf content.UserFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, id)

	var i info.Feed
	if err := u.db.Get(&i, db.SQL("get_user_feed"), id, login); err != nil && err != sql.ErrNoRows {
		u.Err(err)
		return
	}

	uf.Info(i)

	return
}

func (u *User) AddFeed(f content.Feed) (uf content.UserFeed) {
	if u.Err() != nil {
		return
	}

	if err := f.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Info().Login
	i := f.Info()
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, i.Id)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("create_user_feed"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Info().Login, i.Id)
	if err != nil {
		u.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.Err(err)
	}

	uf = NewUserFeed(u.db, u.logger, u)
	uf.Info(i)

	return
}

func (u *User) AllFeeds() (uf []content.TaggedFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting all feeds for user %s\n", login)

	var info []info.Feed
	if err := u.db.Select(&info, db.SQL("get_user_feeds"), login); err != nil {
		u.Err(err)
		return
	}

	uf = make([]content.TaggedFeed, len(info))
	for i := range info {
		uf[i] = NewTaggedFeed(u.db, u.logger, u)
		uf[i].Info(info[i])
	}

	return
}

func (u *User) AllTaggedFeeds() (tf []content.TaggedFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting all tagged feeds for user %s\n", login)

	var feedIdTags []feedIdTag

	if err := u.db.Select(&feedIdTags, db.SQL("get_user_feed_ids_tags"), login); err != nil {
		u.Err(err)
		return
	}

	tf = u.AllFeeds()
	if u.Err() != nil {
		return
	}

	feedMap := make(map[info.FeedId][]content.Tag)

	for _, tuple := range feedIdTags {
		tag := NewTag(u.db, u.logger, u)
		tag.Set(tuple.TagValue)
		feedMap[tuple.FeedId] = append(feedMap[tuple.FeedId], tag)
	}

	for i := range tf {
		tf[i].SetTags(feedMap[tf[i].Info().Id])
	}

	return
}

func (u *User) Article(id info.ArticleId) (ua content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting article '%d' for user %s\n", id, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "a.id = $2", "", []interface{}{id})

	if u.Err() == nil && len(articles) > 0 {
		return articles[0]
	}

	return
}

func (u *User) ArticlesById(ids ...info.ArticleId) (ua []content.UserArticle) {
	if u.Err() != nil || len(ids) == 0 {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting articles %q for user %s\n", ids, login)

	where := "("

	args := []interface{}{}
	index := 1
	for _, id := range ids {
		if index > 1 {
			where += ` OR `
		}

		where += fmt.Sprintf(`a.id = $%d`, index+1)
		args = append(args, id)
		index = len(args) + 1
	}

	where += ")"

	articles := getArticles(u, u.db, u.logger, u, "", "", where, "", args)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) AllUnreadArticleIds() (ids []info.ArticleId) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting unread article ids for user %s\n", login)

	if err := u.db.Select(&ids, db.SQL("get_all_unread_user_article_ids"), login); err != nil {
		u.Err(err)
		return
	}

	return
}

func (u *User) AllFavoriteIds() (ids []info.ArticleId) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting favorite article ids for user %s\n", login)

	if err := u.db.Select(&ids, db.SQL("get_all_favorite_user_article_ids"), login); err != nil {
		u.Err(err)
		return
	}

	return
}

func (u *User) ArticleCount() (c int64) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting article count for user %s\n", login)

	if err := u.db.Get(&c, db.SQL("get_user_article_count"), login); err != nil && err != sql.ErrNoRows {
		u.Err(err)
		return
	}

	return
}

func (u *User) Articles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting articles for paging %q and user %s\n", paging, login)

	order := "read"

	articles := getArticles(u, u.db, u.logger, u, "", "", "", order, nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) UnreadArticles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting unread articles for paging %q and user %s\n", paging, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "ar.article_id IS NULL", "", nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) ArticlesOrderedById(pivot info.ArticleId, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting articles order by id for paging %q and user %s\n", paging, login)

	u.SortingById()

	articles := getArticles(u, u.db, u.logger, u, "", "", "a.id > $2", "", []interface{}{pivot}, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) FavoriteArticles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting favorite articles for paging %q and user %s\n", paging, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "af.article_id IS NOT NULL", "", nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) ReadBefore(date time.Time, read bool) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Marking user %s articles before %v as read: %v\n", login, date, read)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_all_user_articles_read_by_date"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, date)
	if err != nil {
		u.Err(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(db.SQL("create_all_user_articles_read_by_date"))

		if err != nil {
			u.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, date)
		if err != nil {
			u.Err(err)
			return
		}
	}

	tx.Commit()
}

func (u *User) ReadAfter(date time.Time, read bool) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Marking user %s articles after %v as read: %v\n", login, date, read)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(db.SQL("delete_newer_user_articles_read_by_date"))

	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, date)
	if err != nil {
		u.Err(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(db.SQL("create_newer_user_articles_read_by_date"))

		if err != nil {
			u.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, date)
		if err != nil {
			u.Err(err)
			return
		}
	}

	return
}

func (u *User) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting scored articles for paging %q and user %s\n", paging, login)

	order := "asco.score"
	if u.Order() == info.DescendingOrder {
		order = "asco.score DESC"
	}

	ua := getArticles(u, u.db, u.logger, u, "asco.score",
		"INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"a.date > $2 AND a.date <= $3", order,
		[]interface{}{from, to}, paging...)

	sa = make([]content.ScoredArticle, len(ua))
	for i := range ua {
		sa[i] = &ScoredArticle{UserArticle: *ua[i]}
	}

	return sa
}

func (u *User) Tags() (tags []content.Tag) {
	if u.Err() != nil {
		return
	}

	return
}

func getArticles(u content.User, dbo *db.DB, logger webfw.Logger, sorting content.ArticleSorting, columns, join, where, order string, args []interface{}, paging ...int) (ua []*UserArticle) {
	if u.Err() != nil {
		return
	}

	sql := db.SQL("get_article_columns")
	if columns != "" {
		sql += ", " + columns
	}

	sql += db.SQL("get_article_tables")
	if join != "" {
		sql += " " + join
	}

	sql += db.SQL("get_article_joins")

	args = append([]interface{}{u.Info().Login}, args...)
	if where != "" {
		sql += " AND " + where
	}

	sortingField := sorting.Field()
	sortingOrder := sorting.Order()

	fields := []string{}
	if order != "" {
		fields = append(fields, order)
	}
	switch sortingField {
	case info.SortById:
		fields = append(fields, "a.id")
	case info.SortByDate:
		fields = append(fields, "a.date")
	}
	if len(fields) > 0 {
		sql += " ORDER BY "

		sql += strings.Join(fields, ",")

		if sortingOrder == info.DescendingOrder {
			sql += " DESC"
		}
	}

	if len(paging) > 0 {
		limit, offset := pagingLimit(paging)

		sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	var info []info.Article
	if err := dbo.Select(&info, sql, args...); err != nil {
		u.Err(err)
		return
	}

	ua = make([]*UserArticle, len(info))
	for i := range info {
		ua[i] = NewUserArticle(dbo, logger, u)
		ua[i].Info(info[i])
	}

	return
}

func init() {
	db.SetSQL("create_user", createUser)
	db.SetSQL("update_user", updateUser)
	db.SetSQL("delete_user", deleteUser)
	db.SetSQL("get_user_feed", getUserFeed)
	db.SetSQL("create_user_feed", createUserFeed)
	db.SetSQL("get_user_feeds", getUserFeeds)
	db.SetSQL("get_user_tag_feeds", getUserTagFeeds)
	db.SetSQL("get_user_feed_ids_tags", getUserFeedIdsTags)
	db.SetSQL("get_article_columns", getArticleColumns)
	db.SetSQL("get_article_tables", getArticleTables)
	db.SetSQL("get_article_joins", getArticleJoins)
	db.SetSQL("get_all_unread_user_article_ids", getAllUnreadUserArticleIds)
	db.SetSQL("get_all_favorite_user_article_ids", getAllFavoriteUserArticleIds)
	db.SetSQL("get_user_article_count", getUserArticleCount)
	db.SetSQL("create_all_user_articles_read_by_date", createAllUserArticlesReadByDate)
	db.SetSQL("delete_all_user_articles_read_by_date", deleteAllUserArticlesReadByDate)
	db.SetSQL("create_newer_user_articles_read_by_date", createNewerUserArticlesReadByDate)
	db.SetSQL("delete_newer_user_articles_read_by_date", deleteNewerUserArticlesReadByDate)
}

const (
	createUser = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 EXCEPT
	SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	updateUser = `
UPDATE users SET first_name = $1, last_name = $2, email = $3, admin = $4, active = $5, profile_data = $6, hash_type = $7, salt = $8, hash = $9, md5_api = $10
	WHERE login = $11`
	deleteUser  = `DELETE FROM users WHERE login = $1`
	getUserFeed = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND f.id = $1 AND uf.user_login = $2
`
	createUserFeed = `
INSERT INTO users_feeds(user_login, feed_id)
	SELECT $1, $2 EXCEPT SELECT user_login, feed_id FROM users_feeds WHERE user_login = $1 AND feed_id = $2`
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY LOWER(f.title)
`
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY LOWER(f.title)
`
	getUserFeedIdsTags = `SELECT feed_id, tag FROM users_feeds_tags WHERE user_login = $1 ORDER BY feed_id`
	getArticleColumns  = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read,
CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
`

	getArticleTables = `
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
`

	getArticleJoins = `
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND uf.user_login = ar.user_login
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND uf.user_login = af.user_login
WHERE uf.user_login = $1
`
	getAllUnreadUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND uf.user_login = ar.user_login
WHERE ar.article_id IS NULL
`
	getAllFavoriteUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_fav af
	ON a.id = af.article_id AND uf.user_login = af.user_login
WHERE af.article_id IS NOT NULL
`
	getUserArticleCount = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
`
	createAllUserArticlesReadByDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $2)
`
	deleteAllUserArticlesReadByDate = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE date IS NULL OR date < $2
)
`
	createNewerUserArticlesReadByDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1
		AND a.id IN (SELECT id FROM articles WHERE date > $2)
`
	deleteNewerUserArticlesReadByDate = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id FROM articles WHERE date > $2
)
`
)
