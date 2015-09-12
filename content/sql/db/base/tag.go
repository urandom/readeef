package base

func init() {
	sql["get_user_tag_feeds"] = getUserTagFeeds
	sql["get_tag_article_count"] = getTagArticleCount
	sql["get_tag_article_unread_count"] = getTagArticleUnreadCount
}

const (
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY LOWER(f.title)
`
	getTagArticleCount = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id
	AND uft.user_login = uf.user_login
	AND uft.tag = $2
`
	getTagArticleUnreadCount = `
SELECT count(a.id)
FROM users_feeds uf INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND uf.user_login = $1
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id
	AND uft.user_login = uf.user_login
	AND uft.tag = $2
LEFT OUTER JOIN users_articles_states uas
	ON a.id = uas.article_id AND uf.user_login = uas.user_login
WHERE uas.article_id IS NULL OR NOT uas.read
`
)
