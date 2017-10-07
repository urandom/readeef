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
WHERE asco.article_id = :article_id
`
	createArticleScores = `
INSERT INTO articles_scores(article_id, score, score1, score2, score3, score4, score5)
	SELECT :article_id, :score, :score1, :score2, :score3, :score4, :score5 EXCEPT SELECT article_id, score, score1, score2, score3, score4, score5 FROM articles_scores WHERE article_id = :article_id`
	updateArticleScores = `UPDATE articles_scores SET score = :score, score1 = :score1, score2 = :score2, score3 = :score3, score4 = :score4, score5 = :score5 WHERE article_id = :article_id`
)
