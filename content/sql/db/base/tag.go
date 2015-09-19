package base

func init() {
	sqlStmts.Tag.Create = createTag
	sqlStmts.Tag.Update = updateTag
	sqlStmts.Tag.GetUserFeeds = getUserTagFeeds
	sqlStmts.Tag.DeleteStale = deleteStaleTags

	sqlStmts.Tag.GetArticlesJoin = getArticlesTagJoin
	sqlStmts.Tag.ReadStateInsertJoin = readStateInsertTagJoin
	sqlStmts.Tag.ReadStateDeleteJoin = readStateDeleteTagJoin
	sqlStmts.Tag.ArticleCountJoin = articleCountTagJoin
}

const (
	createTag = `
INSERT INTO tags (value)
	SELECT $1 EXCEPT SELECT value FROM tags WHERE value = $1
`
	updateTag = `UPDATE tags SET value = $1 WHERE id = $2`

	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft, tags t
WHERE f.id = uft.feed_id
	AND t.id = uft.tag_id
	AND uft.user_login = $1 AND t.value = $2
ORDER BY LOWER(f.title)
`
	deleteStaleTags = `DELETE FROM tags WHERE id NOT IN (SELECT tag_id FROM users_feeds_tags)`

	getArticlesTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
INNER JOIN tags t
	ON t.id = uft.tag_id
	AND t.value = $2
`

	readStateInsertTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
INNER JOIN tags t
	ON t.id = uft.tag_id
	AND t.value = $2
`
	readStateDeleteTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = a.feed_id AND uft.user_login = $1
INNER JOIN tags t
	ON t.id = uft.tag_id
	AND t.value = $2
`

	articleCountTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = a.feed_id
	AND uft.user_login = $1
INNER JOIN tags t
	ON t.id = uft.tag_id
	AND t.value = $2
`
)
