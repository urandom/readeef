package sql

import (
	"database/sql"
	"fmt"
	"html/template"
	"strings"
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
	getArticleIdsTemplate       *template.Template
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
			Parse(r.db.SQL().User.GetArticlesUserlessTemplate)

		if err != nil {
			return []content.Article{}, errors.Wrap(err, "generating get-articles-userless template")
		}
	}

	renderData := getArticlesData{}
	if o.IncludeScores {
		renderData.Columns += ", asco.score"
	}

	var args []interface{}
	renderData.Join, renderData.Where, renderData.Order, renderData.Limit, args = constructSQLQueryOptions("", o, r.db)

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err = getArticlesUserlessTemplate.Execute(buf, renderData); err != nil {
		return []content.Article{}, errors.Wrap(err, "executing get-articles-userless template")
	}

	sql := buf.String()
	var articles []content.Article

	r.log.Debugf("Articles SQL:\n%s\nArgs:%v\n", sql, args)
	if err = r.db.Select(&articles, sql, args...); err != nil {
		err = errors.Wrap(err, "getting articles")
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
			Parse(s.User.ArticleCountTemplate)

		if err != nil {
			return 0, errors.Wrap(err, "generating article-count template")
		}
	}

	var login content.Login
	var renderData getArticlesData

	if o.UnreadOnly || o.FavoriteOnly || o.ReadOnly || o.UntaggedOnly {
		renderData.Join += s.User.ArticleCountUserFeedsJoin
		login = user.Login

		if o.FavoriteOnly {
			renderData.Join += s.User.StateFavoriteJoin
		}
		if o.ReadOnly || o.UnreadOnly {
			renderData.Join += s.User.StateUnreadJoin
		}
	}

	o.IncludeScores = false

	join, where, _, _, args := constructSQLQueryOptions(login, o, r.db)

	renderData.Join += join
	renderData.Where = where

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := articleCountTemplate.Execute(buf, renderData); err != nil {
		return 0, errors.Wrap(err, "executing article-count template")
	}

	var count int64
	if err = r.db.Get(&count, buf.String(), args...); err != nil {
		return 0, errors.Wrap(err, "getting article count")
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
	if getArticleIdsTemplate == nil {
		getArticleIdsTemplate, err = template.New("article-ids-sql").
			Parse(s.User.GetArticleIdsTemplate)

		if err != nil {
			return []content.ArticleID{}, errors.Wrap(err, "generating article-ids template")
		}
	}

	var login content.Login
	var renderData getArticlesData

	if o.UnreadOnly || o.FavoriteOnly || o.ReadOnly || o.UntaggedOnly {
		renderData.Join += s.User.ArticleCountUserFeedsJoin
		login = user.Login

		if o.FavoriteOnly {
			renderData.Join += s.User.StateFavoriteJoin
		}
		if o.ReadOnly || o.UnreadOnly {
			renderData.Join += s.User.StateUnreadJoin
		}
	}

	o.IncludeScores = false
	join, where, _, _, args := constructSQLQueryOptions(login, o, r.db)

	renderData.Join += join
	renderData.Where = where

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := articleCountTemplate.Execute(buf, renderData); err != nil {
		return []content.ArticleID{}, errors.Wrap(err, "executing article-count template")
	}

	var ids []content.ArticleID
	if err = r.db.Select(&ids, buf.String(), args...); err != nil {
		return []content.ArticleID{}, errors.Wrap(err, "getting article count")
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

func (r articleRepo) RemoveStaleUnreadRecords() error {
	r.log.Infof("Removing stale unread article records")

	_, err := r.db.Exec(r.db.SQL().Repo.DeleteStaleUnreadRecords, time.Now().AddDate(0, -1, 0))
	if err != nil {
		return errors.Wrap(err, "removing stale unread article records")
	}

	return nil
}

func getArticles(login content.Login, dbo *db.DB, log log.Log, opts content.QueryOptions) ([]content.Article, error) {
	var err error
	if getArticlesTemplate == nil {
		getArticlesTemplate, err = template.New("get-articles-sql").
			Parse(dbo.SQL().User.GetArticlesTemplate)

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

	var args []interface{}
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
	if err := dbo.Select(&articles, sql, args...); err != nil {
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
			tmpl = favoriteStateDeleteTemplate
		} else {
			tmpl = favoriteStateInsertTemplate
		}
	}

	s := db.SQL()
	renderData := getArticlesData{}
	var args []interface{}
	renderData.Join, renderData.Where, _, _, args = constructSQLQueryOptions(user.Login, o, db)

	if o.FavoriteOnly {
		renderData.Join += s.User.StateFavoriteJoin
	}
	if o.ReadOnly || o.UnreadOnly {
		renderData.Join += s.User.StateUnreadJoin
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := tmpl.Execute(buf, renderData); err != nil {
		return errors.Wrap(err, "executing article state template")
	}

	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	if _, err := tx.Exec(buf.String(), args); err != nil {
		return errors.Wrap(err, "executing article state statement")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

func constructSQLQueryOptions(
	login content.Login,
	opts content.QueryOptions,
	db *db.DB,
) (string, string, string, string, []interface{}) {

	hasUser := login != ""
	off := 0
	if hasUser {
		off = 1
	}

	var join string

	s := db.SQL()
	if opts.IncludeScores {
		join += s.User.GetArticlesScoreJoin
	}

	if hasUser {
		if opts.UntaggedOnly {
			join += s.User.GetArticlesUntaggedJoin
		}
	}

	args := make([]interface{}, 0, 6)
	if hasUser {
		args = append(args, login)
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
		whereSlice = append(whereSlice, fmt.Sprintf("a.id < $%d", len(args)+off))
		args = append(args, opts.BeforeID)
	}
	if opts.AfterID > 0 {
		whereSlice = append(whereSlice, fmt.Sprintf("a.id > $%d", len(args)+off))
		args = append(args, opts.AfterID)
	}

	if !opts.BeforeDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("(a.date IS NULL OR a.date < $%d)", len(args)+off))
		args = append(args, opts.BeforeDate)
	}

	if !opts.AfterDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("a.date > $%d", len(args)+off))
		args = append(args, opts.AfterDate)
	}

	if len(opts.IDs) > 0 {
		whereSlice = append(whereSlice, db.WhereMultipleORs("a.id", len(opts.IDs), len(args)+off))
		for i := range opts.IDs {
			args = append(args, opts.IDs[i])
		}
	}

	if len(opts.FeedIDs) > 0 {
		whereSlice = append(whereSlice, db.WhereMultipleORs("a.feed_id", len(opts.FeedIDs), len(args)+off))
		for i := range opts.FeedIDs {
			args = append(args, opts.FeedIDs[i])
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

	var limit string
	if opts.Limit > 0 {
		limit = fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+off, len(args)+off+1)
		args = append(args, opts.Limit, opts.Offset)
	}

	return join, where, order, limit, args
}

func updateArticle(a content.Article, tx *sqlx.Tx, db *db.DB, log log.Log) (content.Article, error) {
	if err := a.Validate(); err != nil {
		return content.Article{}, errors.WithMessage(err, "validating article")
	}

	log.Infof("Updating article %s\n", a)

	s := db.SQL()

	stmt, err := tx.Preparex(s.Article.Update)
	if err != nil {
		return content.Article{}, errors.Wrap(err, "preparing article update statement")
	}
	defer stmt.Close()

	res, err := stmt.Exec(a.Title, a.Description, a.Date, a.Guid, a.Link, a.FeedID)
	if err != nil {
		return content.Article{}, errors.Wrap(err, "executing article update statement")
	}

	if num, err := res.RowsAffected(); err != nil && err == sql.ErrNoRows || num == 0 {
		log.Infof("Creating article %s\n", a)

		id, err := db.CreateWithID(tx, s.Article.Create, a.FeedID, a.Link, a.Guid,
			a.Title, a.Description, a.Date)

		if err != nil {
			return content.Article{}, errors.WithMessage(err, "updating article")
		}

		a.ID = content.ArticleID(id)
		a.IsNew = true
	}

	return a, nil
}

func instantiateStateTemplates(s db.SqlStmts) error {
	var err error
	if readStateInsertTemplate == nil {
		readStateInsertTemplate, err = template.New("read-state-insert-sql").
			Parse(s.User.ReadStateInsertTemplate)

		if err != nil {
			return errors.Wrap(err, "generating read-state-insert template")
		}
	}

	if readStateDeleteTemplate == nil {
		readStateDeleteTemplate, err = template.New("read-state-delete-sql").
			Parse(s.User.ReadStateDeleteTemplate)

		if err != nil {
			return errors.Wrap(err, "generating read-state-delete template")
		}
	}

	if favoriteStateInsertTemplate == nil {
		favoriteStateInsertTemplate, err = template.New("favorite-state-insert-sql").
			Parse(s.User.FavoriteStateInsertTemplate)

		if err != nil {
			return errors.Wrap(err, "generating favorite-state-insert template")
		}
	}

	if favoriteStateDeleteTemplate == nil {
		favoriteStateDeleteTemplate, err = template.New("favorite-state-delete-sql").
			Parse(s.User.FavoriteStateDeleteTemplate)

		if err != nil {
			return errors.Wrap(err, "generating favorite-state-delete template")
		}
	}

	return nil
}
