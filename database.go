package dbx

import (
	"context"
	"database/sql"
)

type defaultDatabase struct {
	db *sql.DB
}

func NewDatabase(db *sql.DB) Database {
	return &defaultDatabase{db}
}

func (d *defaultDatabase) Context(ctx context.Context) Context {
	return NewContext(ctx, d)
}

func (d *defaultDatabase) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

func (d *defaultDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

func (d *defaultDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *defaultDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *defaultDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *defaultDatabase) ExecContext(dbContext context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(dbContext, query, args...)
}

func (d *defaultDatabase) QueryContext(dbContext context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(dbContext, query, args...)
}

func (d *defaultDatabase) QueryRowContext(dbContext context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(dbContext, query, args...)
}
