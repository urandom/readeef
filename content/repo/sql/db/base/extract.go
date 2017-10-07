package base

func init() {
	sqlStmts.Extract.Get = getArticleExtract
	sqlStmts.Extract.Create = createArticleExtract
	sqlStmts.Extract.Update = updateArticleExtract
}

const (
	getArticleExtract = `
SELECT ae.title, ae.content, ae.top_image, ae.language
FROM articles_extracts ae
WHERE ae.article_id = :article_id
`
	createArticleExtract = `
INSERT INTO articles_extracts(article_id, title, content, top_image, language)
	SELECT :article_id, :title, :content, :top_image, :language EXCEPT SELECT article_id, title, content, top_image, language FROM articles_extracts WHERE article_id = :article_id
`
	updateArticleExtract = `
UPDATE articles_extracts SET title = :title, content = :content, top_image = :top_image, language = :language WHERE article_id = :article_id`
)
