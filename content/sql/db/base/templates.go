package base

func init() {
	sqlStmts.User.GetArticlesTemplate = getArticlesTemplate
	sqlStmts.User.GetArticlesScoreJoin = getArticlesScoreJoin
	sqlStmts.User.GetArticlesUntaggedJoin = getArticlesUntaggedJoin

	sqlStmts.User.GetArticleIdsTemplate = getArticleIdsTemplate
	sqlStmts.User.GetArticleIdsUserFeedsJoin = articleCountUserFeedsJoin
	sqlStmts.User.GetArticleIdsUnreadJoin = articleCountUnreadJoin
	sqlStmts.User.GetArticleIdsFavoriteJoin = articleCountFavoriteJoin
	sqlStmts.User.GetArticleIdsUntaggedJoin = articleCountUntaggedJoin

	sqlStmts.User.ArticleCountTemplate = articleCountTemplate
	sqlStmts.User.ArticleCountUserFeedsJoin = articleCountUserFeedsJoin
	sqlStmts.User.ArticleCountUnreadJoin = articleCountUnreadJoin
	sqlStmts.User.ArticleCountFavoriteJoin = articleCountFavoriteJoin
	sqlStmts.User.ArticleCountUntaggedJoin = articleCountUntaggedJoin

	sqlStmts.User.ReadStateInsertTemplate = readStateInsertTemplate
	sqlStmts.User.ReadStateInsertFavoriteJoin = readStateInsertFavoriteJoin
	sqlStmts.User.ReadStateInsertUntaggedJoin = getArticlesUntaggedJoin

	sqlStmts.User.ReadStateDeleteTemplate = readStateDeleteTemplate
	sqlStmts.User.ReadStateDeleteFavoriteJoin = readStateInsertFavoriteJoin
	sqlStmts.User.ReadStateDeleteUntaggedJoin = getArticlesUntaggedJoin
}

const (
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

	getArticleIdsTemplate = `
SELECT a.id
FROM articles a
{{ .Join }}
{{ .Where }}
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
	readStateDeleteTemplate = `
DELETE FROM users_articles_unread WHERE user_login = $1 AND article_id IN (
	SELECT a.id FROM articles a {{ .Join }}
	{{ .Where }}
)
`
)
