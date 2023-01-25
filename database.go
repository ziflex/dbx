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

func (d *DefaultDatabase) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *DefaultDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

func (d *DefaultDatabase) Context(ctx context.Context) Context {
	return FromDB(ctx, d.db)
}
