package base

func init() {
	sqlStmts.Tag.Get = getUserTag
	sqlStmts.Tag.GetByValue = getTagByValue
	sqlStmts.Tag.GetUserFeedIDs = getUserTagFeedIDs
	sqlStmts.Tag.AllForUser = getUserTags
	sqlStmts.Tag.AllForFeed = getUserFeedTags
	sqlStmts.Tag.Create = createTag
	sqlStmts.Tag.DeleteStale = deleteStaleTags
}

const (
	getUserTag = `
SELECT t.value
FROM tags t LEFT OUTER JOIN users_feeds_tags uft
	ON t.id = uft.tag_id
WHERE id = $1 AND uft.user_login = $2
`
	getTagByValue   = `SELECT id FROM tags WHERE value = $1`
	getUserFeedTags = `
SELECT t.id, t.value
FROM users_feeds_tags uft INNER JOIN tags t
	ON uft.tag_id = t.id
WHERE uft.user_login = $1 AND uft.feed_id = $2`
	getUserTags = `
SELECT DISTINCT t.id, t.value
FROM tags t LEFT OUTER JOIN users_feeds_tags uft
	ON t.id = uft.tag_id
WHERE uft.user_login = $1
`
	getUserTagFeedIDs = `
SELECT uft.feed_id
FROM users_feeds_tags uft
WHERE uft.user_login = $1 AND uft.tag_id = $2
`

	createTag = `
INSERT INTO tags (value)
	SELECT :value EXCEPT SELECT value FROM tags WHERE value = :value
`

	deleteStaleTags = `DELETE FROM tags WHERE id NOT IN (SELECT tag_id FROM users_feeds_tags)`
)
