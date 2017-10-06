package base

func init() {
	sqlStmts.Thumbnail.Get = getArticleThumbnail
	sqlStmts.Thumbnail.Create = createArticleThumbnail
	sqlStmts.Thumbnail.Update = updateArticleThumbnail
}

const (
	getArticleThumbnail = `
SELECT at.thumbnail, at.link, at.processed
FROM articles_thumbnails at
WHERE at.article_id = $1
`
	createArticleThumbnail = `
INSERT INTO articles_thumbnails(article_id, thumbnail, link, processed)
	SELECT $1, $2, $3, $4 EXCEPT SELECT article_id, thumbnail, link, processed FROM articles_thumbnails WHERE article_id = $1
`
	updateArticleThumbnail = `
UPDATE articles_thumbnails SET thumbnail = $1, link = $2, processed = $3 WHERE article_id = $4`
)
