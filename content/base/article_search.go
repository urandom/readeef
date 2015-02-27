package base

type ArticleSearch struct {
	highlight string
}

func (s *ArticleSearch) Highlight(highlight ...string) string {
	if len(highlight) > 0 {
		s.highlight = highlight[0]
	}

	return s.highlight
}
