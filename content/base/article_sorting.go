package base

type ArticleSorting struct {
	SortingField   SortingField
	ReverseSorting bool
}

func (s ArticleSorting) DefaultSorting() ArticleSorting {
	s.SortingField = DefaultSort

	return s
}

func (s ArticleSorting) SortingById() ArticleSorting {
	s.SortingField = SortById

	return s
}

func (s ArticleSorting) SortingByDate() ArticleSorting {
	s.SortingField = SortByDate

	return s
}

func (s ArticleSorting) Reverse() ArticleSorting {
	s.ReverseSorting = !s.ReverseSorting

	return s
}
