package base

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type ArticleSorting struct {
	SortingField info.SortingField
	sortingOrder info.Order
}

func (s *ArticleSorting) DefaultSorting() content.ArticleSorting {
	s.SortingField = info.DefaultSort

	return s
}

func (s *ArticleSorting) SortingById() content.ArticleSorting {
	s.SortingField = info.SortById

	return s
}

func (s *ArticleSorting) SortingByDate() content.ArticleSorting {
	s.SortingField = info.SortByDate

	return s
}

func (s *ArticleSorting) Reverse() content.ArticleSorting {
	switch s.sortingOrder {
	case info.AscendingOrder:
		s.sortingOrder = info.DescendingOrder
	case info.DescendingOrder:
		s.sortingOrder = info.AscendingOrder
	}

	return s
}

func (s ArticleSorting) Field() info.SortingField {
	return s.SortingField
}

func (s ArticleSorting) Order() info.Order {
	return s.sortingOrder
}
