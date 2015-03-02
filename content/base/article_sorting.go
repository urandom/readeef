package base

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

type ArticleSorting struct {
	sortingField data.SortingField
	sortingOrder data.Order
}

func (s *ArticleSorting) DefaultSorting() content.ArticleSorting {
	s.sortingField = data.DefaultSort

	return s
}

func (s *ArticleSorting) SortingById() content.ArticleSorting {
	s.sortingField = data.SortById

	return s
}

func (s *ArticleSorting) SortingByDate() content.ArticleSorting {
	s.sortingField = data.SortByDate

	return s
}

func (s *ArticleSorting) Reverse() content.ArticleSorting {
	switch s.sortingOrder {
	case data.AscendingOrder:
		s.sortingOrder = data.DescendingOrder
	case data.DescendingOrder:
		s.sortingOrder = data.AscendingOrder
	}

	return s
}

func (s *ArticleSorting) Field(f ...data.SortingField) data.SortingField {
	if len(f) > 0 {
		s.sortingField = f[0]
	}
	return s.sortingField
}

func (s *ArticleSorting) Order(o ...data.Order) data.Order {
	if len(o) > 0 {
		s.sortingOrder = o[0]
	}
	return s.sortingOrder
}
