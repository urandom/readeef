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
)
