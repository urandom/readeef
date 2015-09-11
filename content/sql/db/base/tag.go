package base

func init() {
	sql["get_user_tag_feeds"] = getUserTagFeeds
	sql["create_missing_user_article_state_by_tag_date"] = createMissingUserArticleStateByTagDate
	sql["update_all_user_article_state_by_tag_date"] = updateAllUserArticleReadStateByTagDate
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
	createMissingUserArticleStateByTagDate = `
INSERT INTO users_articles_states (user_login, article_id)
SELECT uf.user_login, a.id
FROM users_feeds uf INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id AND uft.user_login = uf.user_login
		AND uft.user_login = $1 AND uft.tag = $2
INNER JOIN articles a
	ON uf.feed_id = a.feed_id
	AND a.id IN (
		SELECT id FROM articles where date IS NULL OR date < $3
	)
EXCEPT SELECT uas.user_login, uas.article_id
FROM articles a INNER JOIN users_feeds_tags uft
	ON a.feed_id = uft.feed_id
INNER JOIN users_articles_states uas
	ON a.id = uas.article_id
WHERE uas.user_login = $1 AND uft.tag = $2
`
	updateAllUserArticleReadStateByTagDate = `
UPDATE users_articles_states SET read = $1 WHERE user_login = $2 AND article_id IN (
	SELECT a.id
	FROM articles a INNER JOIN users_feeds_tags uft
		ON a.feed_id = uft.feed_id
	WHERE uft.tag = $3 AND (date IS NULL OR date < $4)
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
LEFT OUTER JOIN users_articles_states uas
	ON a.id = uas.article_id AND uf.user_login = uas.user_login
WHERE uas.article_id IS NULL OR NOT uas.read
`
)
