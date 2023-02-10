package dbx

import "database/sql"

type (
	options struct {
		*sql.TxOptions
		AlwaysCreate bool
	}

	Option func(opts *options)
)

func newOptions(setters []Option) *options {
	opts := &options{
		TxOptions: &sql.TxOptions{},
	}

	for _, setter := range setters {
		setter(opts)
	}

	return opts
}

// WithIsolationLevel sets the isolation level for the transaction.
func WithIsolationLevel(level sql.IsolationLevel) Option {
	return func(opts *options) {
		opts.Isolation = level
	}
}

// WithReadOnly sets the read-only flag for the transaction.
func WithReadOnly(readOnly bool) Option {
	return func(opts *options) {
		opts.ReadOnly = readOnly
	}
}

// WithNewTransaction creates a new transaction even if there is an existing transaction in the context.
func WithNewTransaction() Option {
	return func(opts *options) {
		opts.AlwaysCreate = true
	}
}
