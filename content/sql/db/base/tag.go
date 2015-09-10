package base

func init() {
	sql["get_user_tag_feeds"] = getUserTagFeeds
	sql["create_all_user_tag_articles_read_by_date"] = createAllUserTagArticlesByDate
	sql["delete_all_user_tag_articles_read_by_date"] = deleteAllUserTagArticlesByDate
	sql["get_tag_unread_count"] = getTagUnreadCount
}

const (
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY LOWER(f.title)
`
	createAllUserTagArticlesByDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id
	FROM users_feeds uf INNER JOIN users_feeds_tags uft
		ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
			AND uft.user_login = $1 AND uft.tag = $2
	INNER JOIN articles a
		ON uf.feed_id = a.feed_id
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`

	deleteAllUserTagArticlesByDate = `
DELETE FROM users_articles_read WHERE user_login = $1
	AND article_id IN (
		SELECT feed_id FROM users_feeds_tags WHERE user_login = $1 AND tag = $2
	) AND article_id IN (
		SELECT id FROM articles WHERE date IS NULL OR date < $3
	)
`
	getTagUnreadCount = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id
	AND uft.user_login = uf.user_login
	AND uft.tag = $2
LEFT OUTER JOIN users_articles_read ar
	ON a.id = ar.article_id AND uf.user_login = ar.user_login
WHERE ar.article_id IS NULL
`
)
