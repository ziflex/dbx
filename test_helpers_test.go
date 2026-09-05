package dbx_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCountColumn = "count"
	testIDColumn    = "id"
	testNameColumn  = "name"
)

type beginnerOnly struct {
	db           *sql.DB
	beginCalls   int
	beginTxCalls int
	lastOptions  sql.TxOptions
}

func (b *beginnerOnly) Begin() (*sql.Tx, error) {
	b.beginCalls++

	return b.db.Begin() //nolint:noctx // Beginner includes the legacy Begin method for compatibility.
}

func (b *beginnerOnly) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	b.beginTxCalls++
	if opts != nil {
		b.lastOptions = *opts
	}

	return b.db.BeginTx(ctx, opts)
}

func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	database, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		_ = database.Close() // SQL behavior is asserted before best-effort test cleanup.
	})

	return database, mock
}
