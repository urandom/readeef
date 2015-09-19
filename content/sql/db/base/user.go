package base

func init() {
	sqlStmts.User.Create = createUser
	sqlStmts.User.Update = updateUser
	sqlStmts.User.Delete = deleteUser
	sqlStmts.User.GetFeed = getUserFeed
	sqlStmts.User.CreateFeed = createUserFeed
	sqlStmts.User.GetFeeds = getUserFeeds
	sqlStmts.User.GetFeedIdsTags = getUserFeedIdsTags
	sqlStmts.User.GetTags = getUserTags
	sqlStmts.User.GetTag = getTag
	sqlStmts.User.GetTagByValue = GetTagByValue
	sqlStmts.User.GetAllUnreadArticleIds = getAllUnreadUserArticleIds
	sqlStmts.User.GetAllFavoriteArticleIds = getAllFavoriteUserArticleIds

	sqlStmts.User.GetArticlesTemplate = getArticlesTemplate
	sqlStmts.User.GetArticlesScoreJoin = getArticlesScoreJoin
	sqlStmts.User.GetArticlesUntaggedJoin = getArticlesUntaggedJoin

	sqlStmts.User.ReadStateInsertTemplate = readStateInsertTemplate
	sqlStmts.User.ReadStateInsertFavoriteJoin = readStateInsertFavoriteJoin
	sqlStmts.User.ReadStateInsertUntaggedJoin = readStateInsertUntaggedJoin

	sqlStmts.User.ReadStateDeleteTemplate = readStateDeleteTemplate
	sqlStmts.User.ReadStateDeleteFavoriteJoin = readStateInsertFavoriteJoin
	sqlStmts.User.ReadStateDeleteUntaggedJoin = readStateInsertUntaggedJoin

	sqlStmts.User.ArticleCountTemplate = articleCountTemplate
	sqlStmts.User.ArticleCountUserFeedsJoin = articleCountUserFeedsJoin
	sqlStmts.User.ArticleCountUnreadJoin = articleCountUnreadJoin
	sqlStmts.User.ArticleCountFavoriteJoin = articleCountFavoriteJoin
	sqlStmts.User.ArticleCountUntaggedJoin = articleCountUntaggedJoin
}

const (
	createUser = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 EXCEPT
	SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	updateUser = `
UPDATE users SET first_name = $1, last_name = $2, email = $3, admin = $4, active = $5, profile_data = $6, hash_type = $7, salt = $8, hash = $9, md5_api = $10
	WHERE login = $11`
	deleteUser  = `DELETE FROM users WHERE login = $1`
	getUserFeed = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND f.id = $1 AND uf.user_login = $2
`
	createUserFeed = `
INSERT INTO users_feeds(user_login, feed_id)
	SELECT $1, $2 EXCEPT SELECT user_login, feed_id FROM users_feeds WHERE user_login = $1 AND feed_id = $2`
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY LOWER(f.title)
`
	getUserFeedIdsTags = `
SELECT uft.feed_id, t.id, t.value
FROM users_feeds_tags uft INNER JOIN tags t
	ON t.id = uft.tag_id
WHERE uft.user_login = $1 ORDER BY uft.feed_id
`
	getUserTags = `
SELECT DISTINCT t.id, t.value
FROM tags t LEFT OUTER JOIN users_feeds_tags uft
	ON t.id = uft.tag_id
WHERE uft.user_login = $1
`
	getTag        = `SELECT value FROM tags WHERE id = $1`
	GetTagByValue = `SELECT id FROM tags WHERE value = $1`

	getAllUnreadUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_unread au
	ON a.id = au.article_id AND uf.user_login = au.user_login
WHERE au.article_id IS NOT NULL
`
	getAllFavoriteUserArticleIds = `
SELECT a.id
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id AND uf.user_login = $1
LEFT OUTER JOIN users_articles_favorite af
	ON a.id = af.article_id AND uf.user_login = af.user_login
WHERE af.article_id IS NOT NULL
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
	getArticlesScoreJoin = `
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
	{{ .JoinPredicate }}
{{ .Join }}
{{ .Where }}
EXCEPT SELECT au.user_login, au.article_id
FROM users_articles_unread au
WHERE au.user_login = $1
`
	readStateInsertFavoriteJoin = `
LEFT OUTER JOIN users_articles_favorite af
	ON a.id = af.article_id AND af.user_login = uf.user_login
`
	readStateInsertUntaggedJoin = `
LEFT OUTER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id
	AND uft.user_login = uf.user_login
`

	readStateDeleteTemplate = `
DELETE FROM users_articles_unread WHERE user_login = $1 AND article_id IN (
	SELECT a.id FROM articles a {{ .Join }}
	{{ .Where }}
)
`

	articleCountTemplate = `
SELECT count(a.id)
FROM articles a
{{ .Join }}
{{ .Where }}
`
	articleCountUserFeedsJoin = `
INNER JOIN users_feeds uf
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
`
	articleCountUnreadJoin = `
LEFT OUTER JOIN users_articles_unread au
	ON a.id = au.article_id
`
	articleCountFavoriteJoin = `
LEFT OUTER JOIN users_articles_favorite af
	ON a.id = af.article_id
`
	articleCountUntaggedJoin = `
LEFT OUTER JOIN users_feeds_tags uft
	ON uft.feed_id = a.feed_id
`
)
