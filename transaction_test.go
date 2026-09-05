package dbx_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ziflex/dbx"
)

func TestTransaction(t *testing.T) {
	t.Run("commits a successful transaction", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)

		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := dbx.Transaction(context.Background(), db, func(ctx dbx.Context) error {
			_, err := ctx.Executor().Exec("SELECT 1")

			return err
		})

		require.NoError(t, err)
	})

	t.Run("accepts a minimal Beginner implementation", func(t *testing.T) {
		database, mock := newSQLMock(t)
		beginner := &beginnerOnly{db: database}

		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := dbx.Transaction(context.Background(), beginner, func(ctx dbx.Context) error {
			_, err := ctx.Executor().Exec("SELECT 1")

			return err
		})

		require.NoError(t, err)
		assert.Zero(t, beginner.beginCalls)
		assert.Equal(t, 1, beginner.beginTxCalls)
	})

	t.Run("returns a begin error without running the operation", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		beginErr := errors.New("begin transaction")
		operationCalled := false

		mock.ExpectBegin().WillReturnError(beginErr)

		err := dbx.Transaction(context.Background(), db, func(dbx.Context) error {
			operationCalled = true

			return nil
		})

		require.ErrorIs(t, err, beginErr)
		assert.False(t, operationCalled)
	})

	t.Run("returns a commit error", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		commitErr := errors.New("commit transaction")

		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(commitErr)

		err := dbx.Transaction(context.Background(), db, func(ctx dbx.Context) error {
			_, err := ctx.Executor().Exec("SELECT 1")

			return err
		})

		require.ErrorIs(t, err, commitErr)
	})

	t.Run("rolls back on an operation error", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		operationErr := errors.New("run operation")

		mock.ExpectBegin()
		mock.ExpectRollback()

		err := dbx.Transaction(context.Background(), db, func(dbx.Context) error {
			return operationErr
		})

		require.ErrorIs(t, err, operationErr)
		assert.Equal(t, operationErr, err)
	})

	t.Run("joins operation and rollback errors", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		operationErr := errors.New("run operation")
		rollbackErr := errors.New("rollback transaction")

		mock.ExpectBegin()
		mock.ExpectRollback().WillReturnError(rollbackErr)

		err := dbx.Transaction(context.Background(), db, func(dbx.Context) error {
			return operationErr
		})

		require.ErrorIs(t, err, operationErr)
		require.ErrorIs(t, err, rollbackErr)
	})

	t.Run("ignores ErrTxDone while returning the operation error", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		operationErr := errors.New("run operation")

		mock.ExpectBegin()
		mock.ExpectRollback().WillReturnError(sql.ErrTxDone)

		err := dbx.Transaction(context.Background(), db, func(dbx.Context) error {
			return operationErr
		})

		require.ErrorIs(t, err, operationErr)
		assert.Equal(t, operationErr, err)
		assert.NotErrorIs(t, err, sql.ErrTxDone)
	})

	t.Run("rolls back and preserves a panic", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)

		mock.ExpectBegin()
		mock.ExpectRollback()

		assert.PanicsWithValue(t, "operation panic", func() {
			_ = dbx.Transaction(context.Background(), db, func(dbx.Context) error { //nolint:errcheck // The operation must panic before returning.
				panic("operation panic")
			})
		})
	})

	t.Run("reuses an existing transaction", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		unusedBeginner := &beginnerOnly{db: database}

		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := dbx.Transaction(context.Background(), db, func(outer dbx.Context) error {
			if _, err := outer.Executor().Exec("SELECT 1"); err != nil {
				return err
			}

			return dbx.Transaction(outer, unusedBeginner, func(inner dbx.Context) error {
				assert.Same(t, outer, inner)
				assert.Equal(t, outer.Executor(), inner.Executor())

				_, err := inner.Executor().Exec("SELECT 2")

				return err
			}, dbx.WithReadOnly(true))
		})

		require.NoError(t, err)
		assert.Zero(t, unusedBeginner.beginTxCalls)
	})

	t.Run("leaves a reused transaction usable after a panic", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		unusedBeginner := &beginnerOnly{db: database}
		panicValue := errors.New("nested operation panic")

		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := dbx.Transaction(context.Background(), db, func(outer dbx.Context) error {
			assert.PanicsWithValue(t, panicValue, func() {
				_ = dbx.Transaction(outer, unusedBeginner, func(inner dbx.Context) error { //nolint:errcheck // The operation must panic before returning.
					assert.Same(t, outer, inner)
					panic(panicValue)
				})
			})

			_, err := outer.Executor().ExecContext(outer, "SELECT 1")

			return err
		})

		require.NoError(t, err)
		assert.Zero(t, unusedBeginner.beginTxCalls)
	})
}
