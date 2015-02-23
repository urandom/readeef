package base

type ArticleSearch struct {
	HighlightStyle string
}

func (s ArticleSearch) Highlight(highlight string) {
	s.HighlightStyle = highlight
}

func (s ArticleSearch) Query(query string) (ua []UserArticle) {
	// TODO: move search_index.go here and in sql/article_search.go
	return
}
