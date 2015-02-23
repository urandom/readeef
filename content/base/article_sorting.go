package base

import "github.com/urandom/readeef/content"

type ArticleSorting struct {
	SortingField   SortingField
	ReverseSorting bool
}

func (s ArticleSorting) DefaultSorting() content.ArticleSorting {
	s.SortingField = DefaultSort

	return s
}

func (s ArticleSorting) SortingById() content.ArticleSorting {
	s.SortingField = SortById

	return s
}

func (s ArticleSorting) SortingByDate() content.ArticleSorting {
	s.SortingField = SortByDate

	return s
}

func (s ArticleSorting) Reverse() content.ArticleSorting {
	s.ReverseSorting = !s.ReverseSorting

	return s
}
