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
	NamedSQL
	logger webfw.Logger

	db *db.DB
}

func NewUser(db *db.DB, logger webfw.Logger) *User {
	u := &User{NamedSQL: NewNamedSQL(), db: db, logger: logger}

	u.init()

	return u
}

func (u *User) Update() content.User {
	if u.Err() != nil {
		return u
	}

	i := u.Info()
	u.logger.Infof("Updating user %s\n", i.Login)

	tx, err := u.db.Begin()
	if err != nil {
		u.SetErr(err)
		return u
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.SQL("update_user"))
	if err != nil {
		u.SetErr(err)
		return u
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API, i.Login)
	if err != nil {
		u.SetErr(err)
		return u
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			u.SetErr(err)
		}

		return u
	}

	stmt, err = tx.Preparex(u.SQL("create_user"))
	if err != nil {
		u.SetErr(err)
		return u
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login, i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API)
	if err != nil {
		u.SetErr(err)
		return u
	}

	if err := tx.Commit(); err != nil {
		u.SetErr(err)
	}

	return u
}

func (u *User) Delete() content.User {
	if u.Err() != nil {
		return u
	}

	i := u.Info()
	u.logger.Infof("Deleting user %s\n", i.Login)

	if err := u.Validate(); err != nil {
		u.SetErr(err)
		return u
	}

	tx, err := u.db.Begin()
	if err != nil {
		u.SetErr(err)
		return u
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.SQL("delete_user"))
	if err != nil {
		u.SetErr(err)
		return u
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login)
	if err != nil {
		u.SetErr(err)
		return u
	}

	if err := tx.Commit(); err != nil {
		u.SetErr(err)
	}

	return u
}

func (u *User) Feed(id info.FeedId) (uf content.UserFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, id)

	var i info.Feed
	if err := u.db.Get(&i, u.SQL("get_user_feed"), id, login); err != nil && err != sql.ErrNoRows {
		u.SetErr(err)
		return
	}

	uf.Set(i)

	return
}

func (u *User) AddFeed(f content.Feed) (uf content.UserFeed) {
	if u.Err() != nil {
		return
	}

	if err := f.Validate(); err != nil {
		u.SetErr(err)
		return
	}

	login := u.Info().Login
	i := f.Info()
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, i.Id)

	tx, err := u.db.Begin()
	if err != nil {
		u.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.SQL("create_user_feed"))
	if err != nil {
		u.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Info().Login, i.Id)
	if err != nil {
		u.SetErr(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.SetErr(err)
	}

	uf = NewUserFeed(u.db, u.logger, u)
	uf.Set(i)

	return
}

func (u *User) AllFeeds() (uf []content.TaggedFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting all feeds for user %s\n", login)

	var info []info.Feed
	if err := u.db.Select(&info, u.SQL("get_user_feeds"), login); err != nil {
		u.SetErr(err)
		return
	}

	uf = make([]content.TaggedFeed, len(info))
	for i := range info {
		uf[i].Set(info[i])
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

	if err := u.db.Select(&feedIdTags, u.SQL("get_user_feed_ids_tags"), login); err != nil {
		u.SetErr(err)
		return
	}

	tf = u.AllFeeds()
	if u.Err() != nil {
		return
	}

	feedMap := make(map[info.FeedId][]content.Tag)

	for _, tuple := range feedIdTags {
		tag := NewTag(u.db, u.logger)
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

	articles := u.getArticles("", "", "a.id = $2", "", []interface{}{id})

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

	return u.getArticles("", "", where, "", args)
}

func (u *User) AllUnreadArticleIds() (ids []info.ArticleId) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting unread article ids for user %s\n", login)

	if err := u.db.Select(&ids, u.SQL("get_all_unread_user_article_ids"), login); err != nil {
		u.SetErr(err)
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

	if err := u.db.Select(&ids, u.SQL("get_all_favorite_user_article_ids"), login); err != nil {
		u.SetErr(err)
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

	if err := u.db.Get(&c, u.SQL("get_user_article_count"), login); err != nil && err != sql.ErrNoRows {
		u.SetErr(err)
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

	return u.getArticles("", "", "", order, nil, paging...)
}

func (u *User) UnreadArticles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting unread articles for paging %q and user %s\n", paging, login)

	return u.getArticles("", "", "ar.article_id IS NULL", "", nil, paging...)
}

func (u *User) ArticlesOrderedById(pivot info.ArticleId, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting articles order by id for paging %q and user %s\n", paging, login)

	u.SortingById()

	ua = u.getArticles("", "", "a.id > $2", "", []interface{}{pivot}, paging...)

	return
}

func (u *User) FavoriteArticles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting favorite articles for paging %q and user %s\n", paging, login)

	return u.getArticles("", "", "af.article_id IS NOT NULL", "", nil, paging...)
}

func (u *User) ReadBefore(date time.Time, read bool) content.User {
	if u.Err() != nil {
		return u
	}

	return u
}

func (u *User) ReadAfter(date time.Time, read bool) content.User {
	if u.Err() != nil {
		return u
	}

	return u
}

func (u *User) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) Tags() (tags []content.Tag) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) getArticles(columns, join, where, order string, args []interface{}, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	sql := u.SQL("get_article_columns")
	if columns != "" {
		sql += ", " + columns
	}

	sql += u.SQL("get_article_tables")
	if join != "" {
		sql += " " + join
	}

	sql += u.SQL("get_article_joins")

	args = append([]interface{}{u.Info().Login}, args...)
	if where != "" {
		sql += " AND " + where
	}

	sortingField := u.User.ArticleSorting.SortingField
	desc := u.User.ArticleSorting.ReverseSorting

	fields := []string{}
	if order != "" {
		fields = append(fields, order)
	}
	switch sortingField {
	case base.SortById:
		fields = append(fields, "a.id")
	case base.SortByDate:
		fields = append(fields, "a.date")
	}
	if len(fields) > 0 {
		sql += " ORDER BY "

		sql += strings.Join(fields, ",")

		if desc {
			sql += " DESC"
		}
	}

	if len(paging) > 0 {
		limit, offset := pagingLimit(paging)

		sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	var info []info.Article
	if err := u.db.Select(&info, sql, args...); err != nil {
		u.SetErr(err)
		return
	}

	ua = make([]content.UserArticle, len(info))
	for i := range info {
		ua[i].Set(info[i])
	}

	return
}

func (u *User) init() {
	u.SetSQL("create_user", createUser)
	u.SetSQL("update_user", updateUser)
	u.SetSQL("delete_user", deleteUser)
	u.SetSQL("get_user_feed", getUserFeed)
	u.SetSQL("create_user_feed", createUserFeed)
	u.SetSQL("get_user_feeds", getUserFeeds)
	u.SetSQL("get_user_tag_feeds", getUserTagFeeds)
	u.SetSQL("get_user_feed_ids_tags", getUserFeedIdsTags)
	u.SetSQL("get_article_columns", getArticleColumns)
	u.SetSQL("get_article_tables", getArticleTables)
	u.SetSQL("get_article_joins", getArticleJoins)
	u.SetSQL("get_all_unread_user_article_ids", getAllUnreadUserArticleIds)
	u.SetSQL("get_all_favorite_user_article_ids", getAllFavoriteUserArticleIds)
	u.SetSQL("get_user_article_count", getUserArticleCount)
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
)