package db

import "github.com/jmoiron/sqlx"

type Tx struct {
	*sqlx.Tx
	frozen bool
}

func (tx *Tx) Commit() error {
	if tx.frozen {
		return nil
	}

	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	if tx.frozen {
		return nil
	}

	return tx.Tx.Rollback()
}
