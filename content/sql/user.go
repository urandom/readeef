package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type User struct {
	base.User
	logger webfw.Logger

	db *db.DB
}

func (u *User) Update() {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	i := u.Data()
	u.logger.Infof("Updating user %s\n", i.Login)

	tx, err := u.db.Beginx()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("update_user"))
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

	stmt, err = tx.Preparex(u.db.SQL("create_user"))
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
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	i := u.Data()
	u.logger.Infof("Deleting user %s\n", i.Login)

	tx, err := u.db.Beginx()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("delete_user"))
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

func (u *User) FeedById(id data.FeedId) (uf content.UserFeed) {
	uf = u.Repo().UserFeed(u)
	if u.HasErr() {
		uf.Err(u.Err())
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, id)

	var i data.Feed
	if err := u.db.Get(&i, u.db.SQL("get_user_feed"), id, login); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		uf.Err(err)
		return
	}

	uf.Data(i)

	return
}

func (u *User) AddFeed(f content.Feed) (uf content.UserFeed) {
	uf = u.Repo().UserFeed(u)
	if u.HasErr() {
		uf.Err(u.Err())
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	d := f.Data()
	if f.HasErr() {
		uf.Data(d)
		uf.Err(f.Err())
		return
	}

	if err := f.Validate(); err != nil {
		uf.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, d.Id)

	tx, err := u.db.Beginx()
	if err != nil {
		uf.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("create_user_feed"))
	if err != nil {
		uf.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Data().Login, d.Id)
	if err != nil {
		uf.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		uf.Err(err)
	}

	uf.Data(d)

	return
}

func (u *User) AllFeeds() (uf []content.UserFeed) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting all feeds for user %s\n", login)

	var data []data.Feed
	if err := u.db.Select(&data, u.db.SQL("get_user_feeds"), login); err != nil {
		u.Err(err)
		return
	}

	uf = make([]content.UserFeed, len(data))
	for i := range data {
		uf[i] = u.Repo().UserFeed(u)
		uf[i].Data(data[i])
	}

	return
}

func (u *User) AllTaggedFeeds() (tf []content.TaggedFeed) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting all tagged feeds for user %s\n", login)

	var feedIdTags []feedIdTag

	if err := u.db.Select(&feedIdTags, u.db.SQL("get_user_feed_ids_tags"), login); err != nil {
		u.Err(err)
		return
	}

	uf := u.AllFeeds()
	if u.HasErr() {
		return
	}

	feedMap := make(map[data.FeedId][]content.Tag)
	repo := u.Repo()

	for _, tuple := range feedIdTags {
		tag := repo.Tag(u)
		tag.Value(tuple.TagValue)
		feedMap[tuple.FeedId] = append(feedMap[tuple.FeedId], tag)
	}

	tf = make([]content.TaggedFeed, len(uf))
	for i := range uf {
		tf[i] = repo.TaggedFeed(u)
		tf[i].Data(uf[i].Data())
		tf[i].Tags(feedMap[tf[i].Data().Id])
	}

	return
}

func (u *User) ArticleById(id data.ArticleId) (ua content.UserArticle) {
	ua = u.Repo().UserArticle(u)
	if u.HasErr() {
		ua.Err(u.Err())
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting article '%d' for user %s\n", id, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "a.id = $2", "", []interface{}{id})

	if len(articles) > 0 {
		if u.HasErr() {
			articles[0].Err(u.Err())
		}
		return articles[0]
	}

	ua.Err(content.ErrNoContent)

	return
}

func (u *User) ArticlesById(ids []data.ArticleId) (ua []content.UserArticle) {
	if u.HasErr() || len(ids) == 0 {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
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

func (u *User) AllUnreadArticleIds() (ids []data.ArticleId) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting unread article ids for user %s\n", login)

	if err := u.db.Select(&ids, u.db.SQL("get_all_unread_user_article_ids"), login); err != nil {
		u.Err(err)
		return
	}

	return
}

func (u *User) AllFavoriteArticleIds() (ids []data.ArticleId) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting favorite article ids for user %s\n", login)

	if err := u.db.Select(&ids, u.db.SQL("get_all_favorite_user_article_ids"), login); err != nil {
		u.Err(err)
		return
	}

	return
}

func (u *User) ArticleCount() (c int64) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting article count for user %s\n", login)

	if err := u.db.Get(&c, u.db.SQL("get_user_article_count"), login); err != nil && err != sql.ErrNoRows {
		u.Err(err)
		return
	}

	return
}

func (u *User) Articles(paging ...int) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
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
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting unread articles for paging %q and user %s\n", paging, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "ar.article_id IS NULL", "", nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) UnreadCount() (count int64) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting user %s unread count\n", login)

	if err := u.db.Get(&count, u.db.SQL("get_user_unread_count"), login); err != nil {
		u.Err(err)
		return
	}

	return
}

func (u *User) ArticlesOrderedById(pivot data.ArticleId, paging ...int) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting articles order by id for paging %q and user %s\n", paging, login)

	u.SortingById()

	var where string
	switch u.Order() {
	case data.AscendingOrder:
		where = "a.id > $2"
	case data.DescendingOrder:
		where = "a.id < $2"
	}

	articles := getArticles(u, u.db, u.logger, u, "", "", where, "", []interface{}{pivot}, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) FavoriteArticles(paging ...int) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting favorite articles for paging %q and user %s\n", paging, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "af.article_id IS NOT NULL", "", nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) Query(term string, sp content.SearchProvider, paging ...int) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	var err error

	limit, offset := pagingLimit(paging)
	ua, err = sp.Search(term, u, []data.FeedId{}, limit, offset)
	u.Err(err)

	return
}

func (u *User) ReadBefore(date time.Time, read bool) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Marking user %s articles before %v as read: %v\n", login, date, read)

	tx, err := u.db.Beginx()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("delete_all_user_articles_read_by_date"))
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
		stmt, err = tx.Preparex(u.db.SQL("create_all_user_articles_read_by_date"))

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
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Marking user %s articles after %v as read: %v\n", login, date, read)

	tx, err := u.db.Beginx()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("delete_newer_user_articles_read_by_date"))

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
		stmt, err = tx.Preparex(u.db.SQL("create_newer_user_articles_read_by_date"))

		if err != nil {
			u.Err(err)
			return
		}
		defer stmt.Close()

		_, err := stmt.Exec(login, date)
		if err != nil {
			u.Err(err)
			return
		}
	}

	tx.Commit()

	return
}

func (u *User) ScoredArticles(from, to time.Time, paging ...int) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting scored articles for paging %q and user %s\n", paging, login)

	order := "asco.score"
	if u.Order() == data.DescendingOrder {
		order = "asco.score DESC"
	}

	ua = getArticles(u, u.db, u.logger, u, "asco.score",
		"INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"a.date > $2 AND a.date <= $3", order,
		[]interface{}{from, to}, paging...)

	return
}

func (u *User) Tags() (tags []content.Tag) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Data().Login
	u.logger.Infof("Getting all tags for user %s\n", login)

	var feedIdTags []feedIdTag

	if err := u.db.Select(&feedIdTags, u.db.SQL("get_user_tags"), login); err != nil {
		u.Err(err)
		return
	}

	tags = make([]content.Tag, len(feedIdTags))
	for i, tuple := range feedIdTags {
		tag := u.Repo().Tag(u)
		tag.Value(tuple.TagValue)

		tags[i] = tag
	}

	return
}

func getArticles(u content.User, dbo *db.DB, logger webfw.Logger, sorting content.ArticleSorting, columns, join, where, order string, args []interface{}, paging ...int) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	sql := dbo.SQL("get_article_columns")
	if columns != "" {
		sql += ", " + columns
	}

	sql += dbo.SQL("get_article_tables")
	if join != "" {
		sql += " " + join
	}

	sql += dbo.SQL("get_article_joins")

	args = append([]interface{}{u.Data().Login}, args...)
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
	case data.SortById:
		fields = append(fields, "a.id")
	case data.SortByDate:
		fields = append(fields, "a.date")
	}
	if len(fields) > 0 {
		sql += " ORDER BY "

		sql += strings.Join(fields, ",")

		if sortingOrder == data.DescendingOrder {
			sql += " DESC"
		}
	}

	if len(paging) > 0 {
		limit, offset := pagingLimit(paging)

		sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	var data []data.Article
	logger.Debugf("Articles SQL:\n%s\nArgs:%q\n", sql, args)
	if err := dbo.Select(&data, sql, args...); err != nil {
		u.Err(err)
		return
	}

	ua = make([]content.UserArticle, len(data))
	for i := range data {
		ua[i] = u.Repo().UserArticle(u)
		ua[i].Data(data[i])
	}

	return
}
