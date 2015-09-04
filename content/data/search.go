package data

type IndexOperation int

const (
	BatchAdd IndexOperation = iota + 1
	BatchDelete
)
