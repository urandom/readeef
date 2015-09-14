package base

func init() {
	sql["get_user_tag_feeds"] = getUserTagFeeds

	sql["article_count_tag_join"] = articleCountTagJoin
}

const (
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY LOWER(f.title)
`

	articleCountTagJoin = `
INNER JOIN users_feeds_tags uft
	ON uft.feed_id = uf.feed_id
	AND uft.user_login = uf.user_login
	AND uft.tag = $2
`
)
