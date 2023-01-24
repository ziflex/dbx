package dbx

import (
	"context"
	"database/sql"
)

type DefaultDatabase struct {
	db *sql.DB
}

func NewDatabase(db *sql.DB) Database {
	return &DefaultDatabase{db}
}

func (d DefaultDatabase) Transaction(ctx context.Context, receiver Receiver) error {
	return d.TransactionWith(ctx, receiver, nil)
}

func (d DefaultDatabase) TransactionWith(ctx context.Context, receiver Receiver, opts *sql.TxOptions) error {
	tx, err := d.db.BeginTx(ctx, opts)

	if err != nil {
		return err
	}

	err = receiver(WithTx(ctx, tx))

	if err != nil {
		tx.Rollback()

		return err
	}

	return tx.Commit()
}

func (d DefaultDatabase) Context(ctx context.Context) Context {
	return New(ctx, d.db)
}
