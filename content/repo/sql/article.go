package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
	"github.com/urandom/readeef/pool"
)

var (
	getArticlesUserlessTemplate *template.Template
	getArticlesTemplate         *template.Template
	getArticleIDsTemplate       *template.Template
	articleCountTemplate        *template.Template
	readStateInsertTemplate     *template.Template
	readStateDeleteTemplate     *template.Template
	favoriteStateInsertTemplate *template.Template
	favoriteStateDeleteTemplate *template.Template
)

type articleRepo struct {
	db *db.DB

	log log.Log
}

type getArticlesData struct {
	Columns string
	Join    string
	Where   string
	Order   string
	Limit   string
}

const (
	userLogin    = "user_login"
	beforeID     = "before_id"
	afterID      = "after_id"
	beforeDate   = "before_date"
	afterDate    = "after_date"
	idPrefix     = "id"
	feedIDPRefix = "feed_id"
	limit        = "limit"
	offset       = "offset"
)

// ForUser returns all user articles restricted by the QueryOptions
func (r articleRepo) ForUser(user content.User, opts ...content.QueryOpt) ([]content.Article, error) {
	articles := []content.Article{}

	if err := user.Validate(); err != nil {
		return articles, errors.WithMessage(err, "validating user")
	}

	o := content.QueryOptions{}
	o.Apply(opts)

	r.log.Infof("Getting articles for user %s", user)

	articles, err := getArticles(user.Login, r.db, r.log, o)
	if err != nil {
		err = errors.WithMessage(err, fmt.Sprintf("getting articles for user %s", user))
	}

	return articles, err
}

func (r articleRepo) All(opts ...content.QueryOpt) ([]content.Article, error) {
	o := content.QueryOptions{}
	o.Apply(opts)

	r.log.Infof("Getting all articles")

	var err error
	if getArticlesUserlessTemplate == nil {
		getArticlesUserlessTemplate, err = template.New("userless-articles").
			Parse(r.db.SQL().Article.GetUserlessTemplate)

		if err != nil {
			return []content.Article{}, errors.Wrap(err, "generating get-articles-userless template")
		}
	}

	renderData := getArticlesData{}
	if o.IncludeScores {
		renderData.Columns += ", asco.score"
	}

	var args map[string]interface{}
	renderData.Join, renderData.Where, renderData.Order, renderData.Limit, args = constructSQLQueryOptions("", o, r.db)

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err = getArticlesUserlessTemplate.Execute(buf, renderData); err != nil {
		return []content.Article{}, errors.Wrap(err, "executing get-articles-userless template")
	}

	sql := buf.String()
	var articles []content.Article

	r.log.Debugf("Articles SQL:\n%s\nArgs:%v\n", sql, args)

	if err = r.db.WithNamedStmt(sql, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&articles, args)
	}); err != nil {
		err = errors.WithMessage(err, "getting articles")
	}

	return articles, err
}

func (r articleRepo) Count(user content.User, opts ...content.QueryOpt) (int64, error) {
	if err := user.Validate(); err != nil {
		return 0, errors.WithMessage(err, "validating user")
	}

	o := content.QueryOptions{}
	o.Apply(opts)

	r.log.Infof("Getting articles count")

	s := r.db.SQL()

	var err error
	if articleCountTemplate == nil {
		articleCountTemplate, err = template.New("article-count-sql").
			Parse(s.Article.CountTemplate)

		if err != nil {
			return 0, errors.Wrap(err, "generating article-count template")
		}
	}

	var login content.Login
	var renderData getArticlesData

	if user.Login != "" {
		renderData.Join += s.Article.CountUserFeedsJoin
		login = user.Login
	}

	if o.FavoriteOnly {
		renderData.Join += s.Article.StateFavoriteJoin
	}
	if o.ReadOnly || o.UnreadOnly {
		renderData.Join += s.Article.StateUnreadJoin
	}

	o.IncludeScores = false
	o.UnreadFirst = false

	join, where, _, _, args := constructSQLQueryOptions(login, o, r.db)

	renderData.Join += join
	renderData.Where = where

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := articleCountTemplate.Execute(buf, renderData); err != nil {
		return 0, errors.Wrap(err, "executing article-count template")
	}

	var count int64
	r.log.Debugf("Article count SQL:\n%s\nArgs:%v\n", buf.String(), args)

	if err = r.db.WithNamedStmt(buf.String(), nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&count, args)
	}); err != nil {
		return 0, errors.WithMessage(err, "getting article count")
	}

	return count, nil
}

func (r articleRepo) IDs(user content.User, opts ...content.QueryOpt) ([]content.ArticleID, error) {
	if err := user.Validate(); err != nil {
		return []content.ArticleID{}, errors.WithMessage(err, "validating user")
	}

	o := content.QueryOptions{}
	o.Apply(opts)

	r.log.Infof("Getting article ids")

	s := r.db.SQL()

	var err error
	if getArticleIDsTemplate == nil {
		getArticleIDsTemplate, err = template.New("article-ids-sql").
			Parse(s.Article.GetIDsTemplate)

		if err != nil {
			return []content.ArticleID{}, errors.Wrap(err, "generating article-ids template")
		}
	}

	var login content.Login
	var renderData getArticlesData

	if user.Login != "" {
		renderData.Join += s.Article.CountUserFeedsJoin
		login = user.Login
	}

	if o.FavoriteOnly {
		renderData.Join += s.Article.StateFavoriteJoin
	}
	if o.ReadOnly || o.UnreadOnly || o.UnreadFirst {
		renderData.Join += s.Article.StateUnreadJoin
	}

	if o.UnreadFirst {
		renderData.Columns += ", " + s.Article.StateReadColumn
	}

	o.IncludeScores = false

	join, where, order, limit, args := constructSQLQueryOptions(login, o, r.db)

	renderData.Join += join
	renderData.Where = where
	renderData.Order = order
	renderData.Limit = limit

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := getArticleIDsTemplate.Execute(buf, renderData); err != nil {
		return []content.ArticleID{}, errors.Wrap(err, "executing article-count template")
	}

	var ids []content.ArticleID
	r.log.Debugf("Article ids SQL:\n%s\nArgs:%v\n", buf.String(), args)

	if err = r.db.WithNamedStmt(buf.String(), nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&ids, args)
	}); err != nil {
		return []content.ArticleID{}, errors.WithMessage(err, "getting article ids")
	}

	return ids, nil
}

func (r articleRepo) Read(
	state bool,
	user content.User,
	opts ...content.QueryOpt,
) error {
	return articleStateSet(readState, state, user, r.db, r.log, opts)
}

func (r articleRepo) Favor(
	state bool,
	user content.User,
	opts ...content.QueryOpt,
) error {
	return articleStateSet(favoriteState, state, user, r.db, r.log, opts)
}

type staleArgs struct {
	InsertDate time.Time `db:"insert_date"`
}

func (r articleRepo) RemoveStaleUnreadRecords() error {
	r.log.Infof("Removing stale unread article records")

	if err := r.db.WithNamedTx(
		r.db.SQL().Article.DeleteStaleUnreadRecords,
		func(stmt *sqlx.NamedStmt) error {
			_, err := stmt.Exec(staleArgs{time.Now().AddDate(0, -1, 0)})
			return err
		},
	); err != nil {
		return errors.WithMessage(err, "removing stale unread article records")
	}

	return nil
}

func getArticles(login content.Login, dbo *db.DB, log log.Log, opts content.QueryOptions) ([]content.Article, error) {
	var err error
	if getArticlesTemplate == nil {
		getArticlesTemplate, err = template.New("get-articles-sql").
			Parse(dbo.SQL().Article.GetTemplate)

		if err != nil {
			return []content.Article{}, errors.Wrap(err, "generating get-articles template")
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

		articles, err := internalGetArticles(login, dbo, log, opts)
		if err != nil {
			return []content.Article{}, errors.WithMessage(err, "getting unread articles first")
		}

		if !originalUnreadOnly && (opts.Limit == 0 || opts.Limit > len(articles)) {
			if opts.Limit > 0 {
				opts.Limit -= len(articles)
			}
			opts.UnreadOnly = false
			opts.ReadOnly = true

			readOnly, err := internalGetArticles(login, dbo, log, opts)
			if err != nil {
				return []content.Article{}, errors.WithMessage(err, "getting read articles only")
			}

			articles = append(articles, readOnly...)
		}

		return articles, nil
	}

	return internalGetArticles(login, dbo, log, opts)
}

func internalGetArticles(login content.Login, dbo *db.DB, log log.Log, opts content.QueryOptions) ([]content.Article, error) {
	renderData := getArticlesData{}

	var args map[string]interface{}
	renderData.Join, renderData.Where, renderData.Order, renderData.Limit, args = constructSQLQueryOptions(login, opts, dbo)

	if opts.IncludeScores {
		renderData.Columns += ", asco.score"
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := getArticlesTemplate.Execute(buf, renderData); err != nil {
		return []content.Article{}, errors.Wrap(err, "executing get-articles template")
	}

	sql := buf.String()
	var articles []content.Article

	log.Debugf("Articles SQL:\n%s\nArgs:%v\n", sql, args)

	if err := dbo.WithNamedStmt(sql, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Select(&articles, args)
	}); err != nil {
		return []content.Article{}, errors.Wrap(err, "getting articles")
	}

	return articles, nil
}

type stateType int

const (
	readState     stateType = iota
	favoriteState stateType = iota
)

func articleStateSet(
	stateType stateType,
	state bool,
	user content.User,
	db *db.DB,
	log log.Log,
	opts []content.QueryOpt,
) error {
	if err := instantiateStateTemplates(db.SQL()); err != nil {
		return errors.WithMessage(err, "instantiating article state templates")
	}

	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	o := content.QueryOptions{}
	o.Apply(opts)

	var tmpl *template.Template

	switch stateType {
	case readState:
		log.Infof("Setting articles read state")

		if state {
			tmpl = readStateDeleteTemplate
		} else {
			tmpl = readStateInsertTemplate
		}
	case favoriteState:
		log.Infof("Setting articles favorite state")

		if state {
			tmpl = favoriteStateInsertTemplate
		} else {
			tmpl = favoriteStateDeleteTemplate
		}
	}

	s := db.SQL()
	renderData := getArticlesData{}
	var args map[string]interface{}
	renderData.Join, renderData.Where, _, _, args = constructSQLQueryOptions(user.Login, o, db)

	if o.FavoriteOnly {
		renderData.Join += s.Article.StateFavoriteJoin
	}
	if o.ReadOnly || o.UnreadOnly {
		renderData.Join += s.Article.StateUnreadJoin
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := tmpl.Execute(buf, renderData); err != nil {
		return errors.Wrap(err, "executing article state template")
	}

	log.Debugf("Articles state SQL:\n%s\nArgs:%v\n", buf.String(), args)
	if err := db.WithNamedTx(buf.String(), func(stmt *sqlx.NamedStmt) error {
		_, err := stmt.Exec(args)
		return err
	}); err != nil {
		return errors.Wrap(err, "executing article state statement")
	}

	return nil
}

func constructSQLQueryOptions(
	login content.Login,
	opts content.QueryOptions,
	db *db.DB,
) (string, string, string, string, map[string]interface{}) {
	args := map[string]interface{}{}

	hasUser := login != ""

	var join string

	s := db.SQL()
	if opts.IncludeScores {
		join += s.Article.GetScoreJoin
	}

	if hasUser {
		args[userLogin] = login
		if opts.UntaggedOnly {
			join += s.Article.GetUntaggedJoin
		}
	}

	whereSlice := []string{}

	if hasUser {
		if opts.UnreadOnly {
			whereSlice = append(whereSlice, "au.article_id IS NOT NULL")
		} else if opts.ReadOnly {
			whereSlice = append(whereSlice, "au.article_id IS NULL")
		}

		if opts.UntaggedOnly {
			whereSlice = append(whereSlice, "uft.feed_id IS NULL")
		}

		if opts.FavoriteOnly {
			whereSlice = append(whereSlice, "af.article_id IS NOT NULL")
		}
	}

	if opts.BeforeID > 0 {
		whereSlice = append(whereSlice, "a.id < :before_id")
		args[beforeID] = opts.BeforeID
	}
	if opts.AfterID > 0 {
		whereSlice = append(whereSlice, "a.id > :after_id")
		args[afterID] = opts.AfterID
	}

	if !opts.BeforeDate.IsZero() {
		whereSlice = append(whereSlice, "(a.date IS NULL OR a.date < :before_date)")
		args[beforeDate] = opts.BeforeDate
	}

	if !opts.AfterDate.IsZero() {
		whereSlice = append(whereSlice, "a.date > :after_date")
		args[afterDate] = opts.AfterDate
	}

	if len(opts.IDs) > 0 {
		whereSlice = append(whereSlice, db.WhereMultipleORs("a.id", idPrefix, len(opts.IDs)))
		for i := range opts.IDs {
			args[fmt.Sprintf("%s%d", idPrefix, i)] = opts.IDs[i]
		}
	}

	if len(opts.FeedIDs) > 0 {
		whereSlice = append(whereSlice, db.WhereMultipleORs("a.feed_id", feedIDPRefix, len(opts.FeedIDs)))
		for i := range opts.FeedIDs {
			args[fmt.Sprintf("%s%d", feedIDPRefix, i)] = opts.FeedIDs[i]
		}
	}

	var where string
	if len(whereSlice) > 0 {
		where = "WHERE " + strings.Join(whereSlice, " AND ")
	}

	sortingField := opts.SortField
	sortingOrder := opts.SortOrder

	fields := []string{}

	if opts.IncludeScores && opts.HighScoredFirst {
		field := "asco.score"
		if sortingOrder == content.DescendingOrder {
			field += " DESC"
		}
		fields = append(fields, field)
	}

	if hasUser {
		if opts.UnreadFirst {
			fields = append(fields, "read")
		}
	}

	switch sortingField {
	case content.SortByID:
		fields = append(fields, "a.id")
	case content.SortByDate:
		fields = append(fields, "a.date")
	}

	var order string
	if len(fields) > 0 {
		order = " ORDER BY " + strings.Join(fields, ", ")

		if sortingOrder == content.DescendingOrder {
			order += " DESC"
		}
	}

	var paging string
	if opts.Limit > 0 {
		paging = " LIMIT :limit OFFSET :offset"
		args[limit] = opts.Limit
		args[offset] = opts.Offset
	}

	return join, where, order, paging, args
}

func updateArticle(a content.Article, tx *sqlx.Tx, db *db.DB, log log.Log) (content.Article, error) {
	if err := a.Validate(); err != nil && a.ID != 0 {
		return content.Article{}, errors.WithMessage(err, "validating article")
	}

	log.Infof("Updating article %s\n", a)

	s := db.SQL()

	db.WithNamedStmt(s.Article.Update, tx, func(stmt *sqlx.NamedStmt) error {
		res, err := stmt.Exec(a)
		if err != nil {
			return errors.WithMessage(err, "executing article update statement")
		}

		if num, err := res.RowsAffected(); err != nil && err == sql.ErrNoRows || num == 0 {
			log.Infof("Creating article %s\n", a)

			id, err := db.CreateWithID(tx, s.Article.Create, a)

			if err != nil {
				return errors.WithMessage(err, "creating article")
			}

			a.ID = content.ArticleID(id)
			a.IsNew = true
		}

		return nil
	})

	return a, nil
}

func instantiateStateTemplates(s db.SqlStmts) error {
	var err error
	if readStateInsertTemplate == nil {
		readStateInsertTemplate, err = template.New("read-state-insert-sql").
			Parse(s.Article.ReadStateInsertTemplate)

		if err != nil {
			return errors.Wrap(err, "generating read-state-insert template")
		}
	}

	if readStateDeleteTemplate == nil {
		readStateDeleteTemplate, err = template.New("read-state-delete-sql").
			Parse(s.Article.ReadStateDeleteTemplate)

		if err != nil {
			return errors.Wrap(err, "generating read-state-delete template")
		}
	}

	if favoriteStateInsertTemplate == nil {
		favoriteStateInsertTemplate, err = template.New("favorite-state-insert-sql").
			Parse(s.Article.FavoriteStateInsertTemplate)

		if err != nil {
			return errors.Wrap(err, "generating favorite-state-insert template")
		}
	}

	if favoriteStateDeleteTemplate == nil {
		favoriteStateDeleteTemplate, err = template.New("favorite-state-delete-sql").
			Parse(s.Article.FavoriteStateDeleteTemplate)

		if err != nil {
			return errors.Wrap(err, "generating favorite-state-delete template")
		}
	}

	return nil
}
