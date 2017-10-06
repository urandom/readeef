package base

func init() {
	sqlStmts.Scores.Get = getArticleScores
	sqlStmts.Scores.Create = createArticleScores
	sqlStmts.Scores.Update = updateArticleScores
}

const (
	getArticleScores = `
SELECT asco.score, asco.score1, asco.score2, asco.score3, asco.score4, asco.score5
FROM articles_scores asco
WHERE asco.article_id = $1
`
	createArticleScores = `
INSERT INTO articles_scores(article_id, score, score1, score2, score3, score4, score5)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT article_id, score, score1, score2, score3, score4, score5 FROM articles_scores WHERE article_id = $1`
	updateArticleScores = `UPDATE articles_scores SET score = $1, score1 = $2, score2 = $3, score3 = $4, score4 = $5, score5 = $6 WHERE article_id = $7`
)
