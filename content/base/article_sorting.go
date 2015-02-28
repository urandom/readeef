package base

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/info"
)

type ArticleSorting struct {
	sortingField info.SortingField
	sortingOrder info.Order
}

func (s *ArticleSorting) DefaultSorting() content.ArticleSorting {
	s.sortingField = info.DefaultSort

	return s
}

func (s *ArticleSorting) SortingById() content.ArticleSorting {
	s.sortingField = info.SortById

	return s
}

func (s *ArticleSorting) SortingByDate() content.ArticleSorting {
	s.sortingField = info.SortByDate

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

func (s *ArticleSorting) Field(f ...info.SortingField) info.SortingField {
	if len(f) > 0 {
		s.sortingField = f[0]
	}
	return s.sortingField
}

func (s *ArticleSorting) Order(o ...info.Order) info.Order {
	if len(o) > 0 {
		s.sortingOrder = o[0]
	}
	return s.sortingOrder
}
