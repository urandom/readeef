package sql

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Article struct {
	base.Article
}

type ScoredArticle struct {
	Article
	logger webfw.Logger

	db *db.DB
}

type UserArticle struct {
	base.UserArticle
	Article
	logger webfw.Logger

	db *db.DB
}

func (ua *UserArticle) Read(read bool) {
	if ua.HasErr() {
		return
	}

	d := ua.Data()
	if d.Id == 0 {
		ua.Err(content.NewValidationError(errors.New("Invalid article id")))
		return
	}

	login := ua.User().Data().Login
	ua.logger.Infof("Marking user '%s' article '%d' as read: %v\n", login, d.Id, read)

	tx, err := ua.db.Begin()
	if err != nil {
		ua.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(ua.db.SQL("delete_user_article_read"))

	if err != nil {
		ua.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, d.Id)
	if err != nil {
		ua.Err(err)
		return
	}

	d.Read = read

	if read {
		stmt, err = tx.Preparex(ua.db.SQL("create_user_article_read"))
		if err != nil {
			ua.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, d.Id)
		ua.Err(err)
	}

	tx.Commit()

	ua.Data(d)
}

func (ua *UserArticle) Favorite(favorite bool) {
	if ua.HasErr() {
		return
	}

	d := ua.Data()
	if d.Id == 0 {
		ua.Err(content.NewValidationError(errors.New("Invalid article id")))
		return
	}

	login := ua.User().Data().Login
	ua.logger.Infof("Marking user '%s' article '%d' as favorite: %v\n", login, d.Id, favorite)

	tx, err := ua.db.Begin()
	if err != nil {
		ua.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(ua.db.SQL("delete_user_article_favorite"))

	if err != nil {
		ua.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, d.Id)
	if err != nil {
		ua.Err(err)
		return
	}

	d.Favorite = favorite

	if favorite {
		stmt, err = tx.Preparex(ua.db.SQL("create_user_article_favorite"))
		if err != nil {
			ua.Err(err)
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(login, d.Id)
		ua.Err(err)
	}

	tx.Commit()

	ua.Data(d)
}

func (sa *ScoredArticle) Scores() (asc content.ArticleScores) {
	asc = sa.Repo().ArticleScores()
	if sa.HasErr() {
		asc.Err(sa.Err())
		return
	}

	id := sa.Data().Id
	if id == 0 {
		sa.Err(content.NewValidationError(errors.New("Invalid article id")))
		return
	}

	sa.logger.Infof("Getting article '%d' scores\n", id)

	var i data.ArticleScores
	if err := sa.db.Get(&i, sa.db.SQL("get_article_scores"), id); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}
		asc.Err(err)
	}

	asc.Data(i)

	return
}

func query(term, highlight string, index bleve.Index, u content.User, feedIds []data.FeedId, paging ...int) (ua []content.UserArticle, err error) {
	var query bleve.Query

	query = bleve.NewQueryStringQuery(term)

	if len(feedIds) > 0 {
		queries := make([]bleve.Query, len(feedIds))
		conjunct := make([]bleve.Query, 2)

		for i, id := range feedIds {
			q := bleve.NewTermQuery(strconv.FormatInt(int64(id), 10))
			q.SetField("FeedId")

			queries[i] = q
		}

		disjunct := bleve.NewDisjunctionQuery(queries)

		conjunct[0] = query
		conjunct[1] = disjunct

		query = bleve.NewConjunctionQuery(conjunct)
	}

	searchRequest := bleve.NewSearchRequest(query)

	if highlight != "" {
		searchRequest.Highlight = bleve.NewHighlightWithStyle(highlight)
	}

	limit, offset := pagingLimit(paging)
	searchRequest.Size = limit
	searchRequest.From = offset

	searchResult, err := index.Search(searchRequest)

	if err != nil {
		return
	}

	if len(searchResult.Hits) == 0 {
		return
	}

	articleIds := []data.ArticleId{}
	hitMap := map[data.ArticleId]*search.DocumentMatch{}

	for _, hit := range searchResult.Hits {
		if articleId, err := strconv.ParseInt(hit.ID, 10, 64); err == nil {
			id := data.ArticleId(articleId)
			articleIds = append(articleIds, id)
			hitMap[id] = hit
		}
	}

	ua = u.ArticlesById(articleIds)
	if u.HasErr() {
		return ua, u.Err()
	}

	for i := range ua {
		data := ua[i].Data()

		hit := hitMap[data.Id]

		if len(hit.Fragments) > 0 {
			data.Hit.Fragments = hit.Fragments
			ua[i].Data(data)
		}
	}
	return
}
