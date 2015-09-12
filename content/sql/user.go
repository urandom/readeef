package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"text/template"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/util"
)

var (
	getArticlesTemplate     *template.Template
	readStateInsertTemplate *template.Template
	readStateUpdateTemplate *template.Template
)

type getArticlesData struct {
	Columns string
	Join    string
	Where   string
	Order   string
	Limit   string
}

type markReadInsertData struct {
	InnerWhere          string
	InsertJoin          string
	InsertJoinPredicate string
	ExceptJoin          string
	ExceptWhere         string
}

type markReadUpdateData struct {
	InnerJoin  string
	InnerWhere string
}

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

	articles := getArticles(u, u.db, u.logger, data.ArticleQueryOptions{}, u,
		"", "a.id = $2", []interface{}{id})

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

	articles := getArticles(u, u.db, u.logger, data.ArticleQueryOptions{}, u, "", where, args)
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

func (u *User) Articles(o ...data.ArticleQueryOptions) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	var opts data.ArticleQueryOptions
	if len(o) > 0 {
		opts = o[0]
	}

	login := u.Data().Login
	u.logger.Infof("Getting articles for user %s with options: %#v\n", login, opts)

	articles := getArticles(u, u.db, u.logger, opts, u, "", "", nil)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) Count(o ...data.ArticleCountOptions) (count int64) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	var opts data.ArticleCountOptions
	if len(o) > 0 {
		opts = o[0]
	}

	login := u.Data().Login
	u.logger.Infof("Getting user %s article count using options: %#v\n", login, opts)

	if opts.UnreadOnly {
		if err := u.db.Get(&count, u.db.SQL("get_user_article_unread_count"), login); err != nil {
			u.Err(err)
			return
		}
	} else {
		if err := u.db.Get(&count, u.db.SQL("get_user_article_count"), login); err != nil && err != sql.ErrNoRows {
			u.Err(err)
			return
		}

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

func (u *User) ReadState(read bool, o ...data.ArticleUpdateStateOptions) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	var opts data.ArticleUpdateStateOptions
	if len(o) > 0 {
		opts = o[0]
	}

	login := u.Data().Login
	u.logger.Infof("Getting articles for user %s with options: %#v\n", login, opts)

	readState(u, u.db, u.logger, opts, read, "", "", "", "", "", "", nil, nil)
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

func getArticles(u content.User, dbo *db.DB, logger webfw.Logger, opts data.ArticleQueryOptions, sorting content.ArticleSorting, join, where string, args []interface{}) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	var err error
	if getArticlesTemplate == nil {
		getArticlesTemplate, err = template.New("read-state-update-sql").
			Parse(dbo.SQL("get_articles_template"))

		if err != nil {
			u.Err(fmt.Errorf("Error generating get-articles-update template: %v", err))
			return
		}
	}

	renderData := getArticlesData{}
	if opts.IncludeScores {
		renderData.Columns += ", asco.score"
		renderData.Join += " INNER JOIN articles_scores asco ON a.id = asco.article_id"
	}

	if join != "" {
		renderData.Join += " " + join
	}

	args = append([]interface{}{u.Data().Login}, args...)

	whereSlice := []string{}

	if opts.UnreadOnly {
		whereSlice = append(whereSlice, "uas.article_id IS NULL OR NOT uas.read")
	}

	if where != "" {
		whereSlice = append(whereSlice, where)
	}

	if opts.BeforeId > 0 {
		whereSlice = append(whereSlice, fmt.Sprintf("a.id < $%d", len(args)+1))
		args = append(args, opts.BeforeId)
	}
	if opts.AfterId > 0 {
		whereSlice = append(whereSlice, fmt.Sprintf("a.id > $%d", len(args)+1))
		args = append(args, opts.AfterId)
	}

	if opts.FavoriteOnly {
		whereSlice = append(whereSlice, "uas.favorite")
	}

	if !opts.BeforeDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf(" AND a.date <= $%d", len(args)+1))
		args = append(args, opts.BeforeDate)
	}

	if !opts.AfterDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf(" AND a.date > $%d", len(args)+1))
		args = append(args, opts.AfterDate)
	}

	if len(whereSlice) > 0 {
		renderData.Where = " AND " + strings.Join(whereSlice, " AND ")
	}

	sortingField := sorting.Field()
	sortingOrder := sorting.Order()

	fields := []string{}

	if opts.UnreadFirst {
		fields = append(fields, "read")
	}

	if opts.IncludeScores {
		field := "asco.score"
		if sortingOrder == data.DescendingOrder {
			field += " DESC"
		}
		fields = append(fields, field)
	}

	switch sortingField {
	case data.SortById:
		fields = append(fields, "a.id")
	case data.SortByDate:
		fields = append(fields, "a.date")
	}
	if len(fields) > 0 {
		renderData.Order = " ORDER BY " + strings.Join(fields, ", ")

		if sortingOrder == data.DescendingOrder {
			renderData.Order += " DESC"
		}
	}

	if opts.Limit > 0 {
		renderData.Limit = fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, opts.Limit, opts.Offset)
	}

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	if err := getArticlesTemplate.Execute(buf, renderData); err != nil {
		u.Err(fmt.Errorf("Error executing get-articles template: %v", err))
		return
	}

	sql := buf.String()
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

func readState(u content.User, dbo *db.DB, logger webfw.Logger, opts data.ArticleUpdateStateOptions, read bool, insertJoin, insertJoinPredicate, exceptJoin, exceptWhere, updateInnerJoin, updateInnerWhere string, insertArgs, updateArgs []interface{}) {
	if u.HasErr() {
		return
	}

	var err error
	if readStateInsertTemplate == nil {
		readStateInsertTemplate, err = template.New("read-state-insert-sql").
			Parse(dbo.SQL("read_state_insert_template"))

		if err != nil {
			u.Err(fmt.Errorf("Error generating read-state-insert template: %v", err))
			return
		}
	}
	if readStateUpdateTemplate == nil {
		readStateUpdateTemplate, err = template.New("read-state-update-sql").
			Parse(dbo.SQL("read_state_update_template"))

		if err != nil {
			u.Err(fmt.Errorf("Error generating read-state-update template: %v", err))
			return
		}
	}

	tx, err := dbo.Beginx()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	if read {
		args := append([]interface{}{u.Data().Login}, insertArgs...)

		buf := util.BufferPool.GetBuffer()
		defer util.BufferPool.Put(buf)

		data := markReadInsertData{}

		if insertJoin != "" {
			data.InsertJoin = insertJoinPredicate
		}

		if insertJoinPredicate != "" {
			data.InsertJoinPredicate = " AND " + insertJoinPredicate
		}

		innerWhere := []string{}

		if !opts.BeforeDate.IsZero() {
			innerWhere = append(innerWhere, fmt.Sprintf("(date IS NULL OR date < $%d)", len(args)+1))
			args = append(args, opts.BeforeDate)
		}
		if !opts.AfterDate.IsZero() {
			innerWhere = append(innerWhere, fmt.Sprintf("date > $%d", len(args)+1))
			args = append(args, opts.AfterDate)
		}

		if opts.BeforeId > 0 {
			innerWhere = append(innerWhere, fmt.Sprintf("id < $%d", len(args)+1))
			args = append(args, opts.BeforeId)
		}
		if opts.AfterId > 0 {
			innerWhere = append(innerWhere, fmt.Sprintf("id > $%d", len(args)+1))
			args = append(args, opts.AfterId)
		}

		if len(innerWhere) > 0 {
			data.InnerWhere = " WHERE " + strings.Join(innerWhere, " AND ")
		}

		if exceptJoin != "" {
			data.ExceptJoin = exceptJoin
		}

		if exceptWhere != "" {
			data.ExceptWhere = " AND " + exceptWhere
		}

		if err := readStateInsertTemplate.Execute(buf, data); err != nil {
			u.Err(fmt.Errorf("Error executing read-state-insert template: %v", err))
			return
		}

		sql := buf.String()
		logger.Debugf("Read state insert SQL:\n%s\nArgs:%q\n", sql, args)

		stmt, err := tx.Preparex(sql)

		if err != nil {
			u.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(args...)
		if err != nil {
			u.Err(err)
			return
		}
	}

	args := append([]interface{}{read, u.Data().Login}, updateArgs...)

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	data := markReadUpdateData{}

	if updateInnerJoin != "" {
		data.InnerJoin = updateInnerJoin
	}

	where := []string{}

	if updateInnerWhere != "" {
		where = append(where, updateInnerWhere)
	}

	if !opts.BeforeDate.IsZero() {
		where = append(where, fmt.Sprintf("(date IS NULL OR date < $%d)", len(args)+1))
		args = append(args, opts.BeforeDate)
	}
	if !opts.AfterDate.IsZero() {
		where = append(where, fmt.Sprintf("date > $%d", len(args)+1))
		args = append(args, opts.AfterDate)
	}

	if opts.BeforeId > 0 {
		where = append(where, fmt.Sprintf("id < $%d", len(args)+1))
		args = append(args, opts.BeforeId)
	}
	if opts.AfterId > 0 {
		where = append(where, fmt.Sprintf("id > $%d", len(args)+1))
		args = append(args, opts.AfterId)
	}

	if len(where) > 0 {
		data.InnerWhere = " WHERE " + strings.Join(where, " AND ")
	}

	if err := readStateUpdateTemplate.Execute(buf, data); err != nil {
		u.Err(fmt.Errorf("Error executing read-state-update template: %v", err))
		return
	}

	sql := buf.String()
	logger.Debugf("Read state update SQL:\n%s\nArgs:%q\n", sql, args)

	stmt, err := tx.Preparex(sql)

	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		u.Err(err)
		return
	}

	if err = tx.Commit(); err != nil {
		u.Err(err)
	}
}
