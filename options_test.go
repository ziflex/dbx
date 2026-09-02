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

func TestTransactionOptions(t *testing.T) {
	t.Run("passes isolation and read-only options to BeginTx", func(t *testing.T) {
		database, mock := newSQLMock(t)
		beginner := &beginnerOnly{db: database}

		mock.ExpectBegin()
		mock.ExpectCommit()

		err := dbx.Transaction(
			context.Background(),
			beginner,
			func(dbx.Context) error { return nil },
			dbx.WithIsolationLevel(sql.LevelSerializable),
			dbx.WithReadOnly(true),
		)

		require.NoError(t, err)
		assert.Equal(t, 1, beginner.beginTxCalls)
		assert.Equal(t, sql.LevelSerializable, beginner.lastOptions.Isolation)
		assert.True(t, beginner.lastOptions.ReadOnly)
	})

	t.Run("creates an independent transaction when requested", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)

		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectBegin()
		mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		mock.ExpectCommit()

		err := dbx.Transaction(context.Background(), db, func(outer dbx.Context) error {
			if _, err := outer.Executor().Exec("SELECT 1"); err != nil {
				return err
			}

			return dbx.Transaction(outer, db, func(inner dbx.Context) error {
				assert.NotEqual(t, outer.Executor(), inner.Executor())

				_, err := inner.Executor().Exec("SELECT 2")

				return err
			}, dbx.WithNewTransaction())
		})

		require.NoError(t, err)
	})
}

func TestTransactionWithResult(t *testing.T) {
	t.Run("returns the operation result on success", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		rows := sqlmock.NewRows([]string{testCountColumn}).AddRow(5)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnRows(rows)
		mock.ExpectCommit()

		count, err := dbx.TransactionWithResult(context.Background(), db, func(ctx dbx.Context) (int, error) {
			var count int
			err := ctx.Executor().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)

			return count, err
		})

		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("returns the zero value on an operation error", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		operationErr := errors.New("query users")

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnError(operationErr)
		mock.ExpectRollback()

		count, err := dbx.TransactionWithResult(context.Background(), db, func(ctx dbx.Context) (int, error) {
			var count int
			err := ctx.Executor().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)

			return count, err
		})

		require.ErrorIs(t, err, operationErr)
		assert.Zero(t, count)
	})

	t.Run("supports custom result types", func(t *testing.T) {
		type user struct {
			ID   int
			Name string
		}

		database, mock := newSQLMock(t)
		db := dbx.New(database)
		rows := sqlmock.NewRows([]string{testIDColumn, testNameColumn}).AddRow(1, "Alice")

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)
		mock.ExpectCommit()

		result, err := dbx.TransactionWithResult(context.Background(), db, func(ctx dbx.Context) (user, error) {
			var result user
			err := ctx.Executor().QueryRow("SELECT id, name FROM users WHERE id = ?", 1).
				Scan(&result.ID, &result.Name)

			return result, err
		})

		require.NoError(t, err)
		assert.Equal(t, user{ID: 1, Name: "Alice"}, result)
	})

	t.Run("returns the zero value on a commit error", func(t *testing.T) {
		database, mock := newSQLMock(t)
		db := dbx.New(database)
		commitErr := errors.New("commit transaction")

		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(commitErr)

		result, err := dbx.TransactionWithResult(context.Background(), db, func(dbx.Context) (string, error) {
			return "result", nil
		})

		require.ErrorIs(t, err, commitErr)
		assert.Empty(t, result)
	})
}
