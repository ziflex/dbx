package dbx

import (
	"context"
	"database/sql"
)

type DefaultDatabase struct {
	db *sql.DB
}

func New(db *sql.DB) Database {
	return &DefaultDatabase{db}
}

func (d *DefaultDatabase) Context(ctx context.Context) Context {
	return NewContext(ctx, d)
}

func (d *DefaultDatabase) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *DefaultDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

func (d *DefaultDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *DefaultDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *DefaultDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *DefaultDatabase) ExecContext(dbContext context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(dbContext, query, args...)
}

func (d *DefaultDatabase) QueryContext(dbContext context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(dbContext, query, args...)
}

func (d *DefaultDatabase) QueryRowContext(dbContext context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(dbContext, query, args...)
}
