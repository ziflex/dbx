package dbx_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziflex/dbx"
)

func TestDatabase(t *testing.T) {
	t.Run("New should create a database wrapper", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		db := dbx.New(mockDB)
		assert.NotNil(t, db)
	})

	t.Run("Close should close the underlying database", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)

		mock.ExpectClose()

		db := dbx.New(mockDB)
		err = db.Close()

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Context should create a dbx context", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		db := dbx.New(mockDB)
		ctx := db.Context(context.Background())

		assert.NotNil(t, ctx)
		assert.Equal(t, db, ctx.Executor())
	})

	t.Run("Begin should create a transaction", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectBegin()

		db := dbx.New(mockDB)
		tx, err := db.Begin()

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Exec should execute query", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectExec("INSERT INTO users").
			WithArgs("Alice").
			WillReturnResult(sqlmock.NewResult(1, 1))

		db := dbx.New(mockDB)
		result, err := db.Exec("INSERT INTO users (name) VALUES (?)", "Alice")

		assert.NoError(t, err)
		assert.NotNil(t, result)

		id, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), id)

		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Query should execute query and return rows", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")
		mock.ExpectQuery("SELECT id, name FROM users").WillReturnRows(rows)

		db := dbx.New(mockDB)
		result, err := db.Query("SELECT id, name FROM users")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		defer result.Close()

		var users []struct {
			ID   int
			Name string
		}

		for result.Next() {
			var user struct {
				ID   int
				Name string
			}
			err := result.Scan(&user.ID, &user.Name)
			assert.NoError(t, err)
			users = append(users, user)
		}

		assert.Len(t, users, 2)
		assert.Equal(t, "Alice", users[0].Name)
		assert.Equal(t, "Bob", users[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("QueryRow should execute query and return single row", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		rows := sqlmock.NewRows([]string{"count"}).AddRow(42)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").WillReturnRows(rows)

		db := dbx.New(mockDB)
		row := db.QueryRow("SELECT COUNT(*) FROM users")

		var count int
		err = row.Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 42, count)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ExecContext should execute query with context", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectExec("UPDATE users SET name").
			WithArgs("Updated Alice", 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		db := dbx.New(mockDB)
		ctx := context.Background()
		result, err := db.ExecContext(ctx, "UPDATE users SET name = ? WHERE id = ?", "Updated Alice", 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("QueryContext should execute query with context", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
		mock.ExpectQuery("SELECT id, name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)

		db := dbx.New(mockDB)
		ctx := context.Background()
		result, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE id = ?", 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		defer result.Close()

		assert.True(t, result.Next())
		var id int
		var name string
		err = result.Scan(&id, &name)
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
		assert.Equal(t, "Alice", name)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("QueryRowContext should execute query with context", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		rows := sqlmock.NewRows([]string{"name"}).AddRow("Alice")
		mock.ExpectQuery("SELECT name FROM users WHERE id").
			WithArgs(1).
			WillReturnRows(rows)

		db := dbx.New(mockDB)
		ctx := context.Background()
		row := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", 1)

		var name string
		err = row.Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "Alice", name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
