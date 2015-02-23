package sql

import (
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

func (u *User) Update() {
	if u.Err() != nil {
		return
	}

	i := u.Info()
	u.logger.Infof("Updating user %s\n", i.Login)

	tx, err := u.db.Begin()
	if err != nil {
		u.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.SQL("update_user"))
	if err != nil {
		u.SetErr(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API, i.Login)
	if err != nil {
		u.SetErr(err)
		return
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			u.SetErr(err)
		}

		return
	}

	stmt, err = tx.Preparex(u.SQL("create_user"))
	if err != nil {
		u.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login, i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API)
	if err != nil {
		u.SetErr(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.SetErr(err)
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
		u.SetErr(err)
		return
	}

	tx, err := u.db.Begin()
	if err != nil {
		u.SetErr(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.SQL("delete_user"))
	if err != nil {
		u.SetErr(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login)
	if err != nil {
		u.SetErr(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.SetErr(err)
	}
}

func (u *User) Feed(id info.FeedId) (uf content.UserFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, id)

	var i info.Feed
	if err := u.db.Get(&i, u.SQL("get_user_feed"), id, login); err != nil {
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

func (u *User) AllFeeds() (uf []content.UserFeed) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) AllTaggedFeeds() (tf []content.TaggedFeed) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) Article(id info.ArticleId) (ua content.UserArticle) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) ArticlesById(ids ...info.ArticleId) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) AllUnreadArticleIds() (ids []info.ArticleId) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) AllFavoriteIds() (ids []info.ArticleId) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) ArticleCount() (c int64) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) Articles(desc bool, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) UnreadArticles(desc bool, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) ArticlesOrderedById(pivot info.ArticleId, desc bool, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) FavoriteArticles(desc bool, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) ReadBefore(date time.Time, read bool) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) ReadAfter(date time.Time, read bool) {
	if u.Err() != nil {
		return
	}

	return
}

func (u *User) ScoredArticles(from, to time.Time, desc bool, paging ...int) (sa []content.ScoredArticle) {
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

func (u *User) init() {
	u.SetSQL("create_user", createUser)
	u.SetSQL("update_user", updateUser)
	u.SetSQL("delete_user", deleteUser)
	u.SetSQL("get_user_feed", getUserFeed)
	u.SetSQL("create_user_feed", createUserFeed)
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
)
