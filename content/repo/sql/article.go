package sql

import (
	"database/sql"
	"fmt"
	"html/template"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/pool"
)

var (
	getArticlesUserlessTemplate *template.Template
	getArticlesTemplate         *template.Template
	getArticleIdsTemplate       *template.Template
	articleCountTemplate        *template.Template
	readStateInsertTemplate     *template.Template
	readStateDeleteTemplate     *template.Template
)

type articleRepo struct {
	db *db.DB

	log readeef.Logger
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

	o := repo.QueryOptions{}
	o.Apply(opts)

	r.log.Infof("Getting articles for user %s", user)

	articles, err := getArticles(user.Login, r.db, r.log, o, "", "", nil)
	if err != nil {
		err = errors.WithMessage(err, fmt.Sprintf("getting articles for user %s", user))
	}

	return articles, err
}

func (r articleRepo) All(opts ...content.QueryOpt) ([]content.Article, error) {
	o := repo.QueryOptions{}
	o.Apply(opts)

	r.log.Infof("Getting all articles")

	if getArticlesUserlessTemplate == nil {
		getArticlesUserlessTemplate, err = template.New("userless-articles").
			Parse(dbo.SQL().User.GetArticlesUserlessTemplate)

		if err != nil {
			return []content.Article{}, errors.Wrap(err, "generating get-articles-userless template")
		}
	}

	renderData := getArticlesData{}
	s := dbo.SQL()
	if opts.IncludeScores {
		renderData.Columns += ", asco.score"
		renderData.Join += s.User.GetArticlesScoreJoin
	}

	renderData.Where, renderData.Order, renderData.Limit, args = constructSQLQueryOptions("", opts, args, where)

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := getArticlesUserlessTemplate.Execute(buf, renderData); err != nil {
		return []content.Article{}, errors.Wrap(err, "executing get-articles-userless template")
	}

	sql := buf.String()
	var articles []content.Article

	log.Debugf("Articles SQL:\n%s\nArgs:%v\n", sql, args)
	if err = dbo.Select(&articles, sql, args...); err != nil {
		err = errors.Wrap(err, "getting articles")
	}

	return articles, err
}

func getArticles(login content.Login, dbo *db.DB, log readeef.Logger, opts content.QueryOptions, join, where string, args []interface{}) ([]content.Article, error) {
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

		ua = internalGetArticles(login, dbo, log, opts, join, where, args)

		if !originalUnreadOnly && (opts.Limit == 0 || opts.Limit > len(ua)) {
			if opts.Limit > 0 {
				opts.Limit -= len(ua)
			}
			opts.UnreadOnly = false
			opts.ReadOnly = true

			readOnly := internalGetArticles(login, dbo, log, opts, join, where, args)

			ua = append(ua, readOnly...)
		}

		return
	}

	return internalGetArticles(login, dbo, log, opts, join, where, args)
}

func internalGetArticles(login content.Login, dbo *db.DB, log readeef.Logger, opts content.QueryOptions, join, where string, args []interface{}) ([]content.Article, error) {
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

	renderData.Where, renderData.Order, renderData.Limit, args = constructSQLQueryOptions(login, opts, args, where)

	if opts.Limit > 0 {
		renderData.Limit = fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, opts.Limit, opts.Offset)
	}

	buf := pool.Buffer.Get()
	defer pool.Buffer.Put(buf)

	if err := getArticlesTemplate.Execute(buf, renderData); err != nil {
		return []content.Article{}, errors.Wrap(err, "executing get-articles template")
	}

	sql := buf.String()
	var articles []content.Article

	log.Debugf("Articles SQL:\n%s\nArgs:%v\n", sql, args)
	if err = dbo.Select(&articles, sql, args...); err != nil {
		err = errors.Wrap(err, "getting articles")
	}

	return articles, err
}

func constructSQLQueryOptions(
	login content.Login,
	opts content.QueryOptions,
	input []interface{},
	inputWhere string,
) (string, string, string, []interface{}) {

	hasUser := login != ""
	len := len(input)
	off := 0
	if hasUser {
		len++
		off = 1
	}

	args := make([]interface{}, len, len*2)
	if hasUser {
		args[0] = login
		copy(args[1:], input)
	} else {
		copy(args, input)
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

	if inputWhere != "" {
		whereSlice = append(whereSlice, inputWhere)
	}

	if opts.BeforeId > 0 {
		whereSlice = append(whereSlice, fmt.Sprintf("a.id < $%d", len(args)+off))
		args = append(args, opts.BeforeId)
	}
	if opts.AfterId > 0 {
		whereSlice = append(whereSlice, fmt.Sprintf("a.id > $%d", len(args)+off))
		args = append(args, opts.AfterId)
	}

	if !opts.BeforeDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("(a.date IS NULL OR a.date < $%d)", len(args)+off))
		args = append(args, opts.BeforeDate)
	}

	if !opts.AfterDate.IsZero() {
		whereSlice = append(whereSlice, fmt.Sprintf("a.date > $%d", len(args)+off))
		args = append(args, opts.AfterDate)
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

	return where, order, limit, args
}

func updateArticle(a content.Article, tx *sqlx.Tx, db *db.DB, log readeef.Logger) (content.Article, error) {
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

	res, err := stmt.Exec(a.Title, a.Description, a.Date, a.Guid, a.Link, a.FeedId)
	if err != nil {
		return content.Article{}, errors.Wrap(err, "executing article update statement")
	}

	if num, err := res.RowsAffected(); err != nil && err == sql.ErrNoRows || num == 0 {
		log.Infof("Creating article %s\n", a)

		id, err := db.CreateWithID(tx, s.Article.Create, a.FeedId, a.Link, a.Guid,
			a.Title, a.Description, a.Date)

		if err != nil {
			return content.Article{}, errors.WithMessage(err, "updating article")
		}

		a.Id = content.ArticleID(id)
		a.IsNew = true
	}

	return a, nil
}
