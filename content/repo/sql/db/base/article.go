package base

func init() {
	sqlStmts.Article.Create = createFeedArticle
	sqlStmts.Article.Update = updateFeedArticle

	sqlStmts.Article.CountTemplate = articleCountTemplate
	sqlStmts.Article.GetUserlessTemplate = getArticlesUserlessTemplate
	sqlStmts.Article.GetTemplate = getArticlesTemplate
	sqlStmts.Article.CountUserFeedsJoin = articleCountUserFeedsJoin
	sqlStmts.Article.StateReadColumn = stateReadColumn
	sqlStmts.Article.StateUnreadJoin = stateUnreadJoin
	sqlStmts.Article.StateFavoriteJoin = stateFavoriteJoin
	sqlStmts.Article.GetIDsTemplate = getArticleIDsTemplate
	sqlStmts.Article.DeleteStaleUnreadRecords = deleteStaleUnreadRecords
	sqlStmts.Article.GetScoreJoin = getArticlesScoreJoin
	sqlStmts.Article.GetUntaggedJoin = getArticlesUntaggedJoin

	sqlStmts.Article.ReadStateInsertTemplate = readStateInsertTemplate
	sqlStmts.Article.ReadStateDeleteTemplate = readStateDeleteTemplate
	sqlStmts.Article.FavoriteStateInsertTemplate = favoriteStateInsertTemplate
	sqlStmts.Article.FavoriteStateDeleteTemplate = favoriteStateDeleteTemplate
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
	articleCountTemplate = `
SELECT count(a.id)
FROM articles a
{{ .Join }}
{{ .Where }}
`
	getArticlesUserlessTemplate = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
	COALESCE(at.thumbnail, '') as thumbnail,
	COALESCE(at.link, '') as thumbnail_link
	{{ .Columns }}
FROM articles a
{{ .Join }}
LEFT OUTER JOIN articles_thumbnails at
    ON a.id = at.article_id
{{ .Where }}
{{ .Order }}
{{ .Limit }}
`
	getArticlesTemplate = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid,
	CASE WHEN au.article_id IS NULL THEN 1 ELSE 0 END AS read,
	CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite,
	COALESCE(at.thumbnail, '') as thumbnail,
	COALESCE(at.link, '') as thumbnail_link
	{{ .Columns }}
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
{{ .Join }}
LEFT OUTER JOIN users_articles_unread au
    ON a.id = au.article_id AND uf.user_login = au.user_login
LEFT OUTER JOIN users_articles_favorite af
    ON a.id = af.article_id AND uf.user_login = af.user_login
LEFT OUTER JOIN articles_thumbnails at
    ON a.id = at.article_id
{{ .Where }}
{{ .Order }}
{{ .Limit }}
`
	articleCountUserFeedsJoin = `
INNER JOIN users_feeds uf
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
`
	stateReadColumn   = ` CASE WHEN au.article_id IS NULL THEN 1 ELSE 0 END AS read `
	stateFavoriteJoin = `
LEFT OUTER JOIN users_articles_favorite af
	ON a.id = af.article_id AND af.user_login = uf.user_login
`
	stateUnreadJoin = `
LEFT OUTER JOIN users_articles_unread au
	ON a.id = au.article_id AND au.user_login = uf.user_login
`
	getArticleIDsTemplate = `
SELECT a.id FROM (
    SELECT a.id
	{{ .Columns }}
	FROM articles a
	{{ .Join }}
	{{ .Where }}
	{{ .Order }}
	{{ .Limit }}
) a
`
	deleteStaleUnreadRecords = `DELETE FROM users_articles_unread WHERE insert_date < $1`
	getArticlesScoreJoin     = `
	INNER JOIN articles_scores asco ON a.id = asco.article_id
`
	getArticlesUntaggedJoin = `
LEFT OUTER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id
	AND uft.user_login = uf.user_login
`
	readStateInsertTemplate = `
INSERT INTO users_articles_unread (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf
INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
{{ .Join }}
{{ .Where }}
EXCEPT SELECT au.user_login, au.article_id
FROM users_articles_unread au
WHERE au.user_login = $1
`
	readStateDeleteTemplate = `
DELETE FROM users_articles_unread WHERE user_login = $1 AND article_id IN (
	SELECT a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id
		AND uf.user_login = $1
	{{ .Join }}
	{{ .Where }}
)
`
	favoriteStateInsertTemplate = `
INSERT INTO users_articles_favorite (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf
INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
{{ .Join }}
{{ .Where }}
EXCEPT SELECT af.user_login, af.article_id
FROM users_articles_favorite af
WHERE af.user_login = $1
`
	favoriteStateDeleteTemplate = `
DELETE FROM users_articles_favorite WHERE user_login = $1 AND article_id IN (
	SELECT a.id
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id
		AND uf.user_login = $1
	{{ .Join }}
	{{ .Where }}
)
`
)
