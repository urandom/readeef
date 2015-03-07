package base

func init() {
	sql["create_user_article_read"] = createUserArticleRead
	sql["delete_user_article_read"] = deleteUserArticleRead
	sql["create_user_article_favorite"] = createUserArticleFavorite
	sql["delete_user_article_favorite"] = deleteUserArticleFavorite
	sql["get_article_scores"] = getArticleScores
	sql["create_article_scores"] = createArticleScores
	sql["update_article_scores"] = updateArticleScores
}

const (
	createUserArticleRead = `
INSERT INTO users_articles_read(user_login, article_id)
	SELECT $1, $2 EXCEPT
		SELECT user_login, article_id
		FROM users_articles_read WHERE user_login = $1 AND article_id = $2
`
	deleteUserArticleRead = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id = $2`
	createUserArticleFavorite = `
INSERT INTO users_articles_fav(user_login, article_id)
	SELECT $1, $2 EXCEPT
		SELECT user_login, article_id
		FROM users_articles_fav WHERE user_login = $1 AND article_id = $2
`
	deleteUserArticleFavorite = `
DELETE FROM users_articles_fav WHERE  user_login = $1 AND article_id = $2
`
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
