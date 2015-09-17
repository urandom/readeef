package base

func init() {
	sqlStmts.Tag.GetUserFeeds = getUserTagFeeds

	sqlStmts.Tag.GetArticlesJoin = getArticlesTagJoin
	sqlStmts.Tag.ReadStateInsertJoin = readStateInsertTagJoin
	sqlStmts.Tag.ReadStateDeleteJoin = readStateDeleteTagJoin
	sqlStmts.Tag.ArticleCountJoin = articleCountTagJoin
}

const (
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY LOWER(f.title)
`
	getArticlesTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
	AND uft.tag = $2
`

	readStateInsertTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
	AND uft.tag = $2
`
	readStateDeleteTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = a.feed_id AND uft.user_login = $1 AND uft.tag = $2
`

	articleCountTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id
	AND uft.user_login = uf.user_login
	AND uft.tag = $2
`
)
