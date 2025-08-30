package dbx

import "database/sql"

type (
	// options holds configuration settings for transaction creation and behavior.
	// It embeds sql.TxOptions to provide standard transaction configuration
	// while adding dbx-specific settings.
	options struct {
		*sql.TxOptions      // Standard SQL transaction options (isolation, read-only)
		AlwaysCreate   bool // Forces creation of new transaction even if one exists in context
	}

	// Option is a functional option type for configuring transaction behavior.
	// Options are applied when creating transactions through Transaction or
	// TransactionWithResult functions.
	//
	// Example:
	//	err := dbx.Transaction(ctx, db, operation,
	//	    dbx.WithIsolationLevel(sql.LevelSerializable),
	//	    dbx.WithReadOnly(true),
	//	)
	Option func(opts *options)
)

// newOptions creates a new options instance with default values and applies
// the provided option setters. This function is used internally to process
// functional options passed to transaction functions.
//
// Parameters:
//   - setters: Slice of Option functions to apply to the options
//
// Returns:
//   - *options: Configured options instance with all setters applied
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
// This option configures how the transaction isolates its operations
// from other concurrent transactions.
//
// Parameters:
//   - level: The SQL isolation level (e.g., sql.LevelReadCommitted, sql.LevelSerializable)
//
// Returns:
//   - Option: A functional option that can be passed to Transaction functions
//
// Example:
//
//	err := dbx.Transaction(ctx, db, operation,
//	    dbx.WithIsolationLevel(sql.LevelSerializable),
//	)
func WithIsolationLevel(level sql.IsolationLevel) Option {
	return func(opts *options) {
		opts.Isolation = level
	}
}

// WithReadOnly sets the read-only flag for the transaction.
// Read-only transactions can provide performance benefits and prevent
// accidental data modifications.
//
// Parameters:
//   - readOnly: true to make the transaction read-only, false for read-write
//
// Returns:
//   - Option: A functional option that can be passed to Transaction functions
//
// Example:
//
//	// Create a read-only transaction for safe data reading
//	users, err := dbx.TransactionWithResult(ctx, db, getUsersOperation,
//	    dbx.WithReadOnly(true),
//	)
func WithReadOnly(readOnly bool) Option {
	return func(opts *options) {
		opts.ReadOnly = readOnly
	}
}

// WithNewTransaction forces the creation of a new transaction even if there
// is an existing transaction in the context. This is useful when you need
// a separate transaction scope that can be committed or rolled back
// independently of the outer transaction.
//
// By default, dbx reuses existing transactions found in the context to avoid
// nested transaction issues. Use this option when you explicitly need a new
// transaction boundary.
//
// Returns:
//   - Option: A functional option that can be passed to Transaction functions
//
// Example:
//
//	// Force a new transaction even if we're already in a transaction
//	err := dbx.Transaction(ctx, db, independentOperation,
//	    dbx.WithNewTransaction(),
//	)
func WithNewTransaction() Option {
	return func(opts *options) {
		opts.AlwaysCreate = true
	}
}
