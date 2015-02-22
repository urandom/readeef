package base

type SortingField int

const (
	DefaultSort SortingField = iota
	SortById
	SortByDate
)
