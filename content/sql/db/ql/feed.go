package ql

const (
	createFeed = `
INSERT INTO feeds(link, title, description, hub_link, site_link, update_error, subscribe_error)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	updateFeed        = `UPDATE feeds SET link = $1, title = $2, description = $3, hub_link = $4, site_link = $5, update_error = $6, subscribe_error = $7 WHERE id() = $8`
	deleteFeed        = `DELETE FROM feeds WHERE id() = $1`
	createFeedArticle = `
INSERT INTO articles(feed_id, link, guid, title, description, date, _feed_id_guid, _feed_id_link)
	VALUES ($1, $2, $3, $4, $5, $6, formatInt($1) + ':' + $3, formatInt($1) + ':' + $2)
`
	getAllFeedArticles = `
SELECT a.feed_id, id(a), a.title, a.description, a.link, a.guid, a.date
FROM articles a
WHERE a.feed_id = $1
`
	getLatestFeedArticles = `
SELECT a.feed_id, id(a), a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > NOW() - INTERVAL '5 days'
`
	createAllUsersArticlesReadByFeedDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, a.id, uf.user_login + ':' + formatInt(a.id)
	FROM users_feeds uf INNER JOIN articles a
		ON uf.feed_id = a.feed_id AND uf.user_login = $1 AND uf.feed_id = $2
		AND a.id IN (SELECT id FROM articles WHERE date IS NULL OR date < $3)
`
	deleteAllUsersArticlesReadByFeedDate = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id() FROM articles WHERE feed_id = $2 AND (date IS NULL OR date < $3)
)
`
	createUserFeedTag = `
INSERT INTO users_feeds_tags(user_login, feed_id, tag, _login_id_tag)
	VALUES ($1, $2, $3, $1 + ':' + formatInt($2) + ':' + $3)
`
)
