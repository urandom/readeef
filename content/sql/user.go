package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"text/template"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/base/processor"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
	"github.com/urandom/webfw/util"
)

var (
	getArticlesTemplate     *template.Template
	getArticleIdsTemplate   *template.Template
	articleCountTemplate    *template.Template
	readStateInsertTemplate *template.Template
	readStateDeleteTemplate *template.Template
)

type getArticlesData struct {
	Columns string
	Join    string
	Where   string
	Order   string
	Limit   string
}

type articleIdsData struct {
	Join  string
	Where string
}

type articleCountData struct {
	Join  string
	Where string
}

type readStateInsertData struct {
	Join          string
	JoinPredicate string
	Where         string
}

type readStateDeleteData struct {
	Join  string
	Where string
}

type User struct {
	base.User
	logger webfw.Logger

	db *db.DB
}

type feedTagTuple struct {
	data.Tag
	FeedId data.FeedId `db:"feed_id"`
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

	s := u.db.SQL()
	stmt, err := tx.Preparex(s.User.Update)
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

	stmt, err = tx.Preparex(s.User.Create)
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

	stmt, err := tx.Preparex(u.db.SQL().User.Delete)
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
	if err := u.db.Get(&i, u.db.SQL().User.GetFeed, id, login); err != nil {
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

	stmt, err := tx.Preparex(u.db.SQL().User.CreateFeed)
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
	if err := u.db.Select(&data, u.db.SQL().User.GetFeeds, login); err != nil {
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

	var tuples []feedTagTuple

	if err := u.db.Select(&tuples, u.db.SQL().User.GetFeedIdsTags, login); err != nil {
		u.Err(err)
		return
	}

	uf := u.AllFeeds()
	if u.HasErr() {
		return
	}

	feedMap := make(map[data.FeedId][]content.Tag)
	repo := u.Repo()

	for _, t := range tuples {
		tag := repo.Tag(u)
		tag.Data(t.Tag)
		feedMap[t.FeedId] = append(feedMap[t.FeedId], tag)
	}

	tf = make([]content.TaggedFeed, len(uf))
	for i := range uf {
		tf[i] = repo.TaggedFeed(u)
		tf[i].Data(uf[i].Data())
		tf[i].Tags(feedMap[tf[i].Data().Id])
	}

	return
}

func (u *User) ArticleById(id data.ArticleId, o ...data.ArticleQueryOptions) (ua content.UserArticle) {
	ua = u.Repo().UserArticle(u)
	if u.HasErr() {
		ua.Err(u.Err())
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
	u.logger.Infof("Getting article '%d' for user %s\n", id, login)

	articles := getArticles(u, u.db, u.logger, opts, u,
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

func (u *User) ArticlesById(ids []data.ArticleId, o ...data.ArticleQueryOptions) (ua []content.UserArticle) {
	if u.HasErr() || len(ids) == 0 {
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
	u.logger.Infof("Getting articles %q for user %s\n", ids, login)

	where := "a.id IN ("

	args := []interface{}{}
	index := 1
	for _, id := range ids {
		if index > 1 {
			where += `, `
		}

		where += fmt.Sprintf(`$%d`, index+1)
		args = append(args, id)
		index = len(args) + 1
	}
	where += ")"

	articles := getArticles(u, u.db, u.logger, opts, u, "", where, args)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
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

func (u *User) Ids(o ...data.ArticleIdQueryOptions) (ids []data.ArticleId) {
	if u.HasErr() {
		return
	}

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	var opts data.ArticleIdQueryOptions
	if len(o) > 0 {
		opts = o[0]
	}

	u.logger.Infof("Getting user %s article ids using options: %#v\n", u.Data().Login, opts)

	return articleIds(u, u.db, u.logger, opts, "", "", nil)
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

	u.logger.Infof("Getting user %s article count using options: %#v\n", u.Data().Login, opts)

	return articleCount(u, u.db, u.logger, opts, "", "", nil)
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

	readState(u, u.db, u.logger, opts, read, "", "", "", "", nil, nil)
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

	var tagData []data.Tag

	if err := u.db.Select(&tagData, u.db.SQL().User.GetTags, login); err != nil {
		u.Err(err)
		return
	}

	tags = make([]content.Tag, len(tagData))
	for i, d := range tagData {
		tag := u.Repo().Tag(u)
		tag.Data(d)

		tags[i] = tag
	}

	return
}

func (u *User) TagById(id data.TagId) (t content.Tag) {
	t = u.Repo().Tag(u)
	if u.HasErr() {
		t.Err(u.Err())
		return
	}

	u.logger.Infof("Getting tag '%d'\n", id)

	i := data.Tag{}
	if err := u.db.Get(&i, u.db.SQL().User.GetTag, id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		t.Err(err)
		return
	}

	i.Id = id
	t.Data(i)

	return
}

func (u *User) TagByValue(v data.TagValue) (t content.Tag) {
	t = u.Repo().Tag(u)
	if u.HasErr() {
		t.Err(u.Err())
		return
	}

	u.logger.Infof("Getting tag '%s' by value\n", v)

	i := data.Tag{}
	if err := u.db.Get(&i, u.db.SQL().User.GetTagByValue, v); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		t.Err(err)
		return
	}

	i.Value = v
	t.Data(i)

	return
}

func getArticles(u content.User, dbo *db.DB, logger webfw.Logger, opts data.ArticleQueryOptions, sorting content.ArticleSorting, join, where string, args []interface{}) (ua []content.UserArticle) {
	if u.HasErr() {
		return
	}

	var err error
	if getArticlesTemplate == nil {
		getArticlesTemplate, err = template.New("read-state-update-sql").
			Parse(dbo.SQL().User.GetArticlesTemplate)

		if err != nil {
			u.Err(fmt.Errorf("Error generating get-articles-update template: %v", err))
			return
		}
	}

	/* Much faster than using 'ORDER BY read'
	 * TODO: potential overall improvement for fetching pages other than the
	 * first by using the unread count and moving the offset based on it
	 */
	if opts.UnreadFirst && opts.Offset == 0 {
		originalUnreadOnly := opts.UnreadOnly

		opts.UnreadFirst = false
		opts.UnreadOnly = true

		ua = internalGetArticles(u, dbo, logger, opts, sorting, join, where, args)

		if !originalUnreadOnly && (opts.Limit == 0 || opts.Limit > len(ua)) {
			if opts.Limit > 0 {
				opts.Limit -= len(ua)
			}
			opts.UnreadOnly = false
			opts.ReadOnly = true

			readOnly := internalGetArticles(u, dbo, logger, opts, sorting, join, where, args)

			ua = append(ua, readOnly...)
		}

		return
	}

	return internalGetArticles(u, dbo, logger, opts, sorting, join, where, args)
}

func internalGetArticles(u content.User, dbo *db.DB, logger webfw.Logger, opts data.ArticleQueryOptions, sorting content.ArticleSorting, join, where string, args []interface{}) (ua []content.UserArticle) {
	renderData := getArticlesData{}
	s := dbo.SQL()
	if opts.IncludeScores {
		renderData.Columns += ", asco.score"
		renderData.Join += s.User.GetArticlesScoreJoin
	}

	if opts.UntaggedOnly {
		renderData.Join += s.User.GetArticlesUntaggedJoin
	}

	if join != "" {
		renderData.Join += " " + join
	}

	args = append([]interface{}{u.Data().Login}, args...)

	whereSlice := []string{}

	if opts.UnreadOnly {
		whereSlice = append(whereSlice, "au.article_id IS NOT NULL")
	} else if opts.ReadOnly {
		whereSlice = append(whereSlice, "au.article_id IS NULL")
	}

	if opts.UntaggedOnly {
		whereSlice = append(whereSlice, "uft.feed_id IS NULL")
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
		whereSlice = append(whereSlice, "af.article_id IS NOT NULL")
	}

	if !opts.BeforeDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("(a.date IS NULL OR a.date < $%d)", len(args)+1))
		args = append(args, opts.BeforeDate)
	}

	if !opts.AfterDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("a.date > $%d", len(args)+1))
		args = append(args, opts.AfterDate)
	}

	if len(whereSlice) > 0 {
		renderData.Where = "WHERE " + strings.Join(whereSlice, " AND ")
	}

	sortingField := sorting.Field()
	sortingOrder := sorting.Order()

	fields := []string{}

	if opts.IncludeScores && opts.HighScoredFirst {
		field := "asco.score"
		if sortingOrder == data.DescendingOrder {
			field += " DESC"
		}
		fields = append(fields, field)
	}

	if opts.UnreadFirst {
		fields = append(fields, "read")
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
	logger.Debugf("Articles SQL:\n%s\nArgs:%v\n", sql, args)
	if err := dbo.Select(&data, sql, args...); err != nil {
		u.Err(err)
		return
	}

	ua = make([]content.UserArticle, len(data))
	for i := range data {
		ua[i] = u.Repo().UserArticle(u)
		ua[i].Data(data[i])
	}

	processors := u.Repo().ArticleProcessors()
	if !opts.SkipProcessors && len(processors) > 0 {
		for _, p := range processors {
			if opts.SkipSessionProcessors {
				if _, ok := p.(processor.ProxyHTTP); ok {
					continue
				}
			}
			ua = p.ProcessArticles(ua)
		}
	}

	return
}

func articleIds(u content.User, dbo *db.DB, logger webfw.Logger, opts data.ArticleIdQueryOptions, join, where string, args []interface{}) (ids []data.ArticleId) {
	if u.HasErr() {
		return
	}

	s := dbo.SQL()
	var err error
	if getArticleIdsTemplate == nil {
		getArticleIdsTemplate, err = template.New("article-ids-sql").
			Parse(s.User.GetArticleIdsTemplate)

		if err != nil {
			u.Err(fmt.Errorf("Error generating article-ids template: %v", err))
			return
		}
	}

	renderData := articleIdsData{}
	containsUserFeeds := !opts.UnreadOnly && !opts.FavoriteOnly

	if containsUserFeeds {
		renderData.Join += s.User.GetArticleIdsUserFeedsJoin
	} else {
		if opts.UnreadOnly {
			renderData.Join += s.User.GetArticleIdsUnreadJoin
		}
		if opts.FavoriteOnly {
			renderData.Join += s.User.GetArticleIdsFavoriteJoin
		}
	}

	if opts.UntaggedOnly {
		renderData.Join += s.User.GetArticleIdsUntaggedJoin
	}

	if join != "" {
		renderData.Join += " " + join
	}

	args = append([]interface{}{u.Data().Login}, args...)

	whereSlice := []string{}

	if opts.UnreadOnly {
		whereSlice = append(whereSlice, "au.article_id IS NOT NULL AND au.user_login = $1")
	}
	if opts.FavoriteOnly {
		whereSlice = append(whereSlice, "af.article_id IS NOT NULL AND af.user_login = $1")
	}
	if opts.UntaggedOnly {
		whereSlice = append(whereSlice, "uft.feed_id IS NULL")
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

	if !opts.BeforeDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("(a.date IS NULL OR a.date < $%d)", len(args)+1))
		args = append(args, opts.BeforeDate)
	}

	if !opts.AfterDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("a.date > $%d", len(args)+1))
		args = append(args, opts.AfterDate)
	}

	if len(whereSlice) > 0 {
		renderData.Where = "WHERE " + strings.Join(whereSlice, " AND ")
	}

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	if err := getArticleIdsTemplate.Execute(buf, renderData); err != nil {
		u.Err(fmt.Errorf("Error executing article-ids template: %v", err))
		return
	}

	sql := buf.String()

	logger.Debugf("Article ids SQL:\n%s\nArgs:%v\n", sql, args)
	if err := dbo.Select(&ids, sql, args...); err != nil {
		u.Err(err)
		return
	}

	return
}

func articleCount(u content.User, dbo *db.DB, logger webfw.Logger, opts data.ArticleCountOptions, join, where string, args []interface{}) (count int64) {
	if u.HasErr() {
		return
	}

	s := dbo.SQL()
	var err error
	if articleCountTemplate == nil {
		articleCountTemplate, err = template.New("article-count-sql").
			Parse(s.User.ArticleCountTemplate)

		if err != nil {
			u.Err(fmt.Errorf("Error generating article-count template: %v", err))
			return
		}
	}

	renderData := articleCountData{}
	containsUserFeeds := !opts.UnreadOnly && !opts.FavoriteOnly

	if containsUserFeeds {
		renderData.Join += s.User.ArticleCountUserFeedsJoin
	} else {
		if opts.UnreadOnly {
			renderData.Join += s.User.ArticleCountUnreadJoin
		}
		if opts.FavoriteOnly {
			renderData.Join += s.User.ArticleCountFavoriteJoin
		}
	}

	if opts.UntaggedOnly {
		renderData.Join += s.User.ArticleCountUntaggedJoin
	}

	if join != "" {
		renderData.Join += " " + join
	}

	args = append([]interface{}{u.Data().Login}, args...)

	whereSlice := []string{}

	if opts.UnreadOnly {
		whereSlice = append(whereSlice, "au.article_id IS NOT NULL AND au.user_login = $1")
	}
	if opts.FavoriteOnly {
		whereSlice = append(whereSlice, "af.article_id IS NOT NULL AND af.user_login = $1")
	}
	if opts.UntaggedOnly {
		whereSlice = append(whereSlice, "uft.feed_id IS NULL")
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

	if !opts.BeforeDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("(a.date IS NULL OR a.date < $%d)", len(args)+1))
		args = append(args, opts.BeforeDate)
	}

	if !opts.AfterDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("a.date > $%d", len(args)+1))
		args = append(args, opts.AfterDate)
	}

	if len(whereSlice) > 0 {
		renderData.Where = "WHERE " + strings.Join(whereSlice, " AND ")
	}

	buf := util.BufferPool.GetBuffer()
	defer util.BufferPool.Put(buf)

	if err := articleCountTemplate.Execute(buf, renderData); err != nil {
		u.Err(fmt.Errorf("Error executing article-count template: %v", err))
		return
	}

	sql := buf.String()

	logger.Debugf("Article count SQL:\n%s\nArgs:%v\n", sql, args)
	if err := dbo.Get(&count, sql, args...); err != nil {
		u.Err(err)
		return
	}

	return
}

func readState(u content.User, dbo *db.DB, logger webfw.Logger, opts data.ArticleUpdateStateOptions, read bool, join, joinPredicate, deleteJoin, deleteWhere string, insertArgs, deleteArgs []interface{}) {
	if u.HasErr() {
		return
	}

	s := dbo.SQL()

	var err error
	if readStateInsertTemplate == nil {
		readStateInsertTemplate, err = template.New("read-state-insert-sql").
			Parse(s.User.ReadStateInsertTemplate)

		if err != nil {
			u.Err(fmt.Errorf("Error generating read-state-insert template: %v", err))
			return
		}
	}
	if readStateDeleteTemplate == nil {
		readStateDeleteTemplate, err = template.New("read-state-delete-sql").
			Parse(s.User.ReadStateDeleteTemplate)

		if err != nil {
			u.Err(fmt.Errorf("Error generating read-state-delete template: %v", err))
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
		args := append([]interface{}{u.Data().Login}, deleteArgs...)

		buf := util.BufferPool.GetBuffer()
		defer util.BufferPool.Put(buf)

		data := readStateDeleteData{}

		if deleteJoin != "" {
			data.Join = deleteJoin
		}

		if opts.FavoriteOnly {
			data.Join += s.User.ReadStateDeleteFavoriteJoin
		}

		if opts.UntaggedOnly {
			data.Join += s.User.ReadStateDeleteUntaggedJoin
		}

		where := []string{}

		if deleteWhere != "" {
			where = append(where, deleteWhere)
		}

		if !opts.BeforeDate.IsZero() {
			where = append(where, fmt.Sprintf("(a.date IS NULL OR a.date < $%d)", len(args)+1))
			args = append(args, opts.BeforeDate)
		}
		if !opts.AfterDate.IsZero() {
			where = append(where, fmt.Sprintf("a.date > $%d", len(args)+1))
			args = append(args, opts.AfterDate)
		}

		if opts.BeforeId > 0 {
			where = append(where, fmt.Sprintf("a.id < $%d", len(args)+1))
			args = append(args, opts.BeforeId)
		}
		if opts.AfterId > 0 {
			where = append(where, fmt.Sprintf("a.id > $%d", len(args)+1))
			args = append(args, opts.AfterId)
		}

		if opts.FavoriteOnly {
			where = append(where, "af.article_id IS NOT NULL")
		}

		if opts.UntaggedOnly {
			where = append(where, "uft.feed_id IS NULL")
		}

		if len(where) > 0 {
			data.Where = " WHERE " + strings.Join(where, " AND ")
		}

		if err := readStateDeleteTemplate.Execute(buf, data); err != nil {
			u.Err(fmt.Errorf("Error executing read-state-delete template: %v", err))
			return
		}

		sql := buf.String()
		logger.Debugf("Read state delete SQL:\n%s\nArgs:%v\n", sql, args)

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
	} else {
		args := append([]interface{}{u.Data().Login}, insertArgs...)

		buf := util.BufferPool.GetBuffer()
		defer util.BufferPool.Put(buf)

		data := readStateInsertData{}

		if joinPredicate != "" {
			data.JoinPredicate = " AND " + joinPredicate
		}

		if opts.FavoriteOnly {
			data.Join += s.User.ReadStateInsertFavoriteJoin
		}

		if opts.UntaggedOnly {
			data.Join += s.User.ReadStateInsertUntaggedJoin
		}

		if join != "" {
			data.Join += joinPredicate
		}

		where := []string{}

		if !opts.BeforeDate.IsZero() {
			where = append(where, fmt.Sprintf("(a.date IS NULL OR a.date < $%d)", len(args)+1))
			args = append(args, opts.BeforeDate)
		}
		if !opts.AfterDate.IsZero() {
			where = append(where, fmt.Sprintf("a.date > $%d", len(args)+1))
			args = append(args, opts.AfterDate)
		}

		if opts.BeforeId > 0 {
			where = append(where, fmt.Sprintf("a.id < $%d", len(args)+1))
			args = append(args, opts.BeforeId)
		}
		if opts.AfterId > 0 {
			where = append(where, fmt.Sprintf("a.id > $%d", len(args)+1))
			args = append(args, opts.AfterId)
		}

		if opts.FavoriteOnly {
			where = append(where, "af.article_id IS NOT NULL")
		}

		if opts.UntaggedOnly {
			where = append(where, "uft.feed_id IS NULL")
		}

		if len(where) > 0 {
			data.Where = " WHERE " + strings.Join(where, " AND ")
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

	if err = tx.Commit(); err != nil {
		u.Err(err)
	}
}
