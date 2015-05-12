package ql

// FIXME: Needs alternative to left outer joins for getArticle*

const (
	createUser = `
INSERT INTO users(login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	getUserFeed = `
SELECT id(f), f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE id(f) = uf.feed_id
	AND id(f) = $1 AND uf.user_login = $2
`
	createUserFeed = `
INSERT INTO users_feeds(user_login, feed_id, _login_id)
	VALUES ($1, $2, $1 + ":" + formatInt($2))`
	getUserFeeds = `
SELECT id(f), f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE id(f) = uf.feed_id
	AND uf.user_login = $1
ORDER BY LOWER(f.title)
`
	getUserArticleCount = `
SELECT count(id(a))
FROM users_feeds uf, articles a
WHERE uf.feed_id = a.feed_id AND uf.user_login = $1
`
	createAllUserArticlesReadByDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, id(a), uf.user_login + ":" + formatInt(id(a))
	FROM users_feeds uf, articles a
	WHERE uf.feed_id = a.feed_id AND uf.user_login = $1
		AND id(a) IN (SELECT id() FROM articles WHERE date IS NULL OR date < $2)
`
	deleteAllUserArticlesReadByDate = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id() FROM articles WHERE date IS NULL OR date < $2
)
`
	createNewerUserArticlesReadByDate = `
INSERT INTO users_articles_read
	SELECT uf.user_login, id(a), uf.user_login + ":" + formatInt(id(a))
	FROM users_feeds uf, articles a
	WHERE uf.feed_id = a.feed_id AND uf.user_login = $1
		AND id(a) IN (SELECT id() FROM articles WHERE date > $2)
`
	deleteNewerUserArticlesReadByDate = `
DELETE FROM users_articles_read WHERE user_login = $1 AND article_id IN (
	SELECT id(a) FROM articles WHERE date > $2
)
`
)
