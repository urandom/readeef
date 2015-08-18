package base

func init() {
	sql["create_feed_article"] = createFeedArticle
	sql["update_feed_article"] = updateFeedArticle
	sql["create_user_article_read"] = createUserArticleRead
	sql["delete_user_article_read"] = deleteUserArticleRead
	sql["create_user_article_favorite"] = createUserArticleFavorite
	sql["delete_user_article_favorite"] = deleteUserArticleFavorite
	sql["get_article_scores"] = getArticleScores
	sql["create_article_scores"] = createArticleScores
	sql["update_article_scores"] = updateArticleScores
	sql["get_article_thumbnail"] = getArticleThumbnail
	sql["create_article_thumbnail"] = createArticleThumbnail
	sql["update_article_thumbnail"] = updateArticleThumbnail
}

const (
	createFeedArticle = `
INSERT INTO articles(feed_id, link, guid, title, description, date)
	SELECT $1, $2, $3, $4, $5, $6 EXCEPT
		SELECT feed_id, link, guid, title, description, date
		FROM articles WHERE feed_id = $1 AND link = $2
`

	updateFeedArticle = `
UPDATE articles SET title = $1, description = $2, date = $3, guid = $4, link = $5
	WHERE feed_id = $6 AND (guid = $4 OR link = $5)
`

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

	getArticleThumbnail = `
SELECT at.thumbnail, at.link, at.mime_type, at.processed
FROM articles_thumbnails at
WHERE at.article_id = $1
`
	createArticleThumbnail = `
INSERT INTO articles_thumbnails(article_id, thumbnail, link, mime_type, processed)
	SELECT $1, $2, $3, $4, $5 EXCEPT SELECT article_id, thumbnail, link, mime_type, processed FROM articles_thumbnails WHERE article_id = $1
`
	updateArticleThumbnail = `
UPDATE articles_thumbnails SET thumbnail = $1, link = $2, mime_type = $3, processed = $4 WHERE article_id = $5`
)
