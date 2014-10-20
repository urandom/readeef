package readeef

import (
	"database/sql"
	"time"
)

const (
	get_latest_feed_articles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > NOW() - INTERVAL '5 days'
`

	get_article_scores = `
SELECT asco.score, asco.score1, asco.score2, asco.score3, asco.score4, asco.score5
FROM articles_scores asco
WHERE asco.article_id = $1
`

	create_article_scores = `
INSERT INTO articles_scores(article_id, score, score1, score2, score3, score4, score5)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT article_id, score, score1, score2, score3, score4, score5 FROM articles_scores WHERE article_id = $1`

	update_article_scores = `UPDATE articles_scores SET score = $1, score1 = $2, score2 = $3, score3 = $4, score4 = $5, score5 = $6 WHERE article_id = $7`
)

func (db DB) GetLatestFeedArticles(f Feed) ([]Article, error) {
	var articles []Article

	if err := db.Select(&articles, db.NamedSQL("get_latest_feed_articles"), f.Id); err != nil {
		return articles, err
	}

	return articles, nil
}

func (db DB) GetScoredUserArticles(u User, since time.Time, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"a.date > $2", "asco.score, a.date", []interface{}{since}, paging...)
}

func (db DB) GetScoredUserArticlesDesc(u User, since time.Time, paging ...int) ([]Article, error) {
	return db.getArticles(u, "", "INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"a.date > $2", "asco.score DESC, a.date DESC", []interface{}{since}, paging...)
}

func (db DB) GetArticleScores(a Article) (ArticleScores, error) {
	var asc ArticleScores
	if err := db.Get(&asc, db.NamedSQL("get_article_scores"), a.Id); err != nil && err != sql.ErrNoRows {
		return asc, err
	}

	asc.ArticleId = a.Id

	return asc, nil
}

func (db DB) UpdateArticleScores(asc ArticleScores) error {
	if err := asc.Validate(); err != nil {
		return err
	}

	// FIXME: Remove when the 'FOREIGN KEY constraing failed' bug is removed
	if db.driver == "sqlite3" {
		db.Query("SELECT 1")
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ustmt, err := tx.Preparex(db.NamedSQL("update_article_scores"))
	if err != nil {
		return err
	}
	defer ustmt.Close()

	res, err := ustmt.Exec(asc.Score, asc.Score1, asc.Score2, asc.Score3, asc.Score4, asc.Score5, asc.ArticleId)
	if err != nil {
		return err
	}

	if num, err := res.RowsAffected(); err != nil || num == 0 {
		cstmt, err := tx.Preparex(db.NamedSQL("create_article_scores"))
		if err != nil {
			return err
		}
		defer cstmt.Close()

		_, err = cstmt.Exec(asc.ArticleId, asc.Score, asc.Score1, asc.Score2, asc.Score3, asc.Score4, asc.Score5)
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func init() {
	sql_stmt["generic:get_latest_feed_articles"] = get_latest_feed_articles
	sql_stmt["generic:get_article_scores"] = get_article_scores
	sql_stmt["generic:create_article_scores"] = create_article_scores
	sql_stmt["generic:update_article_scores"] = update_article_scores
}
