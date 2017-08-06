package base

func init() {
	sqlStmts.Article.Create = createFeedArticle
	sqlStmts.Article.Update = updateFeedArticle
	sqlStmts.Article.CreateUserUnread = createUserArticleUnread
	sqlStmts.Article.DeleteUserUnread = deleteUserArticleUnread
	sqlStmts.Article.CreateUserFavorite = createUserArticleFavorite
	sqlStmts.Article.DeleteUserFavorite = deleteUserArticleFavorite
	sqlStmts.Article.GetScores = getArticleScores
	sqlStmts.Article.CreateScores = createArticleScores
	sqlStmts.Article.UpdateScores = updateArticleScores
	sqlStmts.Article.GetThumbnail = getArticleThumbnail
	sqlStmts.Article.CreateThumbnail = createArticleThumbnail
	sqlStmts.Article.UpdateThumbnail = updateArticleThumbnail
	sqlStmts.Article.GetExtract = getArticleExtract
	sqlStmts.Article.CreateExtract = createArticleExtract
	sqlStmts.Article.UpdateExtract = updateArticleExtract
}

const (
	createFeedArticle = `
INSERT INTO articles(feed_id, link, guid, title, description, date)
	SELECT $1, $2, $3, $4, $5, $6 EXCEPT
		SELECT feed_id, link, CAST($3 AS TEXT), CAST($4 as TEXT), CAST($5 AS TEXT), CAST($6 AS TIMESTAMP WITH TIME ZONE)
		FROM articles WHERE feed_id = $1 AND link = $2
`

	updateFeedArticle = `
UPDATE articles SET title = $1, description = $2, date = $3, guid = $4, link = $5
	WHERE feed_id = $6 AND (guid = $4 OR link = $5)
`

	createUserArticleUnread = `
INSERT INTO users_articles_unread(user_login, article_id)
	SELECT $1, $2 EXCEPT
		SELECT user_login, article_id
		FROM users_articles_unread WHERE user_login = $1 AND article_id = $2
`

	deleteUserArticleUnread = `
DELETE FROM users_articles_unread WHERE user_login = $1 AND article_id = $2`

	createUserArticleFavorite = `
INSERT INTO users_articles_favorite(user_login, article_id)
	SELECT $1, $2 EXCEPT
		SELECT user_login, article_id
		FROM users_articles_favorite WHERE user_login = $1 AND article_id = $2
`

	deleteUserArticleFavorite = `
DELETE FROM users_articles_favorite WHERE user_login = $1 AND article_id = $2`

	getArticleScores = `
SELECT asco.score, asco.score1, asco.score2, asco.score3, asco.score4, asco.score5
FROM articles_scores asco
WHERE asco.article_id = $1
`
	createArticleScores = `
INSERT INTO articles_scores(article_id, score, score1, score2, score3, score4, score5)
	SELECT $1, $2, $3, $4, $5, $6, $7 EXCEPT SELECT article_id, score, score1, score2, score3, score4, score5 FROM articles_scores WHERE article_id = $1`
	updateArticleScores = `UPDATE articles_scores SET score = $1, score1 = $2, score2 = $3, score3 = $4, score4 = $5, score5 = $6 WHERE article_id = $7`

	getArticleThumbnail = `
SELECT at.thumbnail, at.link, at.processed
FROM articles_thumbnails at
WHERE at.article_id = $1
`
	createArticleThumbnail = `
INSERT INTO articles_thumbnails(article_id, thumbnail, link, processed)
	SELECT $1, $2, $3, $4 EXCEPT SELECT article_id, thumbnail, link, processed FROM articles_thumbnails WHERE article_id = $1
`
	updateArticleThumbnail = `
UPDATE articles_thumbnails SET thumbnail = $1, link = $2, processed = $3 WHERE article_id = $4`

	getArticleExtract = `
SELECT ae.title, ae.content, ae.top_image, ae.language
FROM articles_extracts ae
WHERE ae.article_id = $1
`
	createArticleExtract = `
INSERT INTO articles_extracts(article_id, title, content, top_image, language)
	SELECT $1, $2, $3, $4, $5 EXCEPT SELECT article_id, title, content, top_image, language FROM articles_extracts WHERE article_id = $1
`
	updateArticleExtract = `
UPDATE articles_extracts SET title = $1, content = $2, top_image = $3, language = $4 WHERE article_id = $5`
)
