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
WHERE ae.article_id = $1
`
	createArticleExtract = `
INSERT INTO articles_extracts(article_id, title, content, top_image, language)
	SELECT $1, $2, $3, $4, $5 EXCEPT SELECT article_id, title, content, top_image, language FROM articles_extracts WHERE article_id = $1
`
	updateArticleExtract = `
UPDATE articles_extracts SET title = $1, content = $2, top_image = $3, language = $4 WHERE article_id = $5`
)
