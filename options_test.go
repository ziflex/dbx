package dbx_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziflex/dbx"
)

func TestTransactionOptions(t *testing.T) {
	t.Run("WithIsolationLevel should set isolation level", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectBegin().WillReturnError(nil) // Default expectation
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		db := dbx.New(mockDB)
		ctx := context.Background()

		err = dbx.Transaction(ctx, db, func(c dbx.Context) error {
			_, e := c.Executor().Exec("SELECT 1")
			return e
		}, dbx.WithIsolationLevel(sql.LevelReadCommitted))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("WithReadOnly should set read-only flag", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
		mock.ExpectCommit()

		db := dbx.New(mockDB)
		ctx := context.Background()

		err = dbx.Transaction(ctx, db, func(c dbx.Context) error {
			_, e := c.Executor().Query("SELECT * FROM users")
			return e
		}, dbx.WithReadOnly(true))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("WithNewTransaction should create new transaction even if one exists", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		// Expect two separate transactions
		mock.ExpectBegin() // First transaction
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectBegin() // Second transaction (new one)
		mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit() // Second transaction commit
		mock.ExpectCommit() // First transaction commit

		db := dbx.New(mockDB)
		ctx := context.Background()

		err = dbx.Transaction(ctx, db, func(c1 dbx.Context) error {
			c1.Executor().Exec("SELECT 1")

			// This should create a NEW transaction instead of reusing
			return dbx.Transaction(c1, db, func(c2 dbx.Context) error {
				c2.Executor().Exec("SELECT 2")
				return nil
			}, dbx.WithNewTransaction())
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("multiple options should work together", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectCommit()

		db := dbx.New(mockDB)
		ctx := context.Background()

		err = dbx.Transaction(ctx, db, func(c dbx.Context) error {
			_, e := c.Executor().Query("SELECT * FROM users")
			return e
		}, dbx.WithIsolationLevel(sql.LevelSerializable), dbx.WithReadOnly(true))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTransactionWithResult(t *testing.T) {
	t.Run("should return result on success", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnRows(rows)
		mock.ExpectCommit()

		db := dbx.New(mockDB)
		ctx := context.Background()

		count, err := dbx.TransactionWithResult(ctx, db, func(c dbx.Context) (int, error) {
			var count int
			err := c.Executor().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
			return count, err
		})

		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return zero value on error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		testErr := assert.AnError
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnError(testErr)
		mock.ExpectRollback()

		db := dbx.New(mockDB)
		ctx := context.Background()

		count, err := dbx.TransactionWithResult(ctx, db, func(c dbx.Context) (int, error) {
			var count int
			err := c.Executor().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
			return count, err
		})

		assert.Error(t, err)
		assert.Equal(t, 0, count) // zero value for int
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should work with custom types", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}

		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)
		mock.ExpectCommit()

		db := dbx.New(mockDB)
		ctx := context.Background()

		user, err := dbx.TransactionWithResult(ctx, db, func(c dbx.Context) (User, error) {
			var user User
			err := c.Executor().QueryRow("SELECT id, name FROM users WHERE id = ?", 1).
				Scan(&user.ID, &user.Name)
			return user, err
		})

		assert.NoError(t, err)
		assert.Equal(t, User{ID: 1, Name: "Alice"}, user)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle begin error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		testErr := assert.AnError
		mock.ExpectBegin().WillReturnError(testErr)

		db := dbx.New(mockDB)
		ctx := context.Background()

		result, err := dbx.TransactionWithResult(ctx, db, func(c dbx.Context) (string, error) {
			return "should not reach here", nil
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
		assert.Equal(t, "", result) // zero value for string
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should handle commit error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		testErr := assert.AnError
		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(testErr)

		db := dbx.New(mockDB)
		ctx := context.Background()

		result, err := dbx.TransactionWithResult(ctx, db, func(c dbx.Context) (string, error) {
			c.Executor().Exec("SELECT 1")
			return "success", nil
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
		assert.Equal(t, "", result) // zero value for string
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
