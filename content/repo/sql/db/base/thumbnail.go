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
WHERE at.article_id = :article_id
`
	createArticleThumbnail = `
INSERT INTO articles_thumbnails(article_id, thumbnail, link, processed)
	SELECT :article_id, :thumbnail, :link, :processed EXCEPT SELECT article_id, thumbnail, link, processed FROM articles_thumbnails WHERE article_id = :article_id 
`
	updateArticleThumbnail = `
UPDATE articles_thumbnails SET thumbnail = :thumbnail, link = :link, processed = :processed WHERE article_id = :article_id`
)
