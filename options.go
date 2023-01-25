package dbx

import "database/sql"

type Option func(opts *sql.TxOptions)

func WithIsolationLevel(level sql.IsolationLevel) Option {
	return func(opts *sql.TxOptions) {
		opts.Isolation = level
	}
}

func WithReadOnly(readOnly bool) Option {
	return func(opts *sql.TxOptions) {
		opts.ReadOnly = readOnly
	}
}
