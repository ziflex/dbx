package testing_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ziflex/dbx"
	dbxtesting "github.com/ziflex/dbx/testing"
)

func TestMockExecutor(t *testing.T) {
	t.Run("Exec should work with mock expectations", func(t *testing.T) {
		mockExec := dbxtesting.NewMockExecutor()
		result := dbxtesting.NewResult(1, 1)

		mockExec.On("Exec", "INSERT INTO users (name) VALUES (?)", "John").
			Return(result, nil)

		actualResult, err := mockExec.Exec("INSERT INTO users (name) VALUES (?)", "John")

		assert.NoError(t, err)
		assert.NotNil(t, actualResult)

		lastID, _ := actualResult.LastInsertId()
		rowsAffected, _ := actualResult.RowsAffected()
		assert.Equal(t, int64(1), lastID)
		assert.Equal(t, int64(1), rowsAffected)

		mockExec.AssertExpectations(t)
	})

	t.Run("Query should work with mock expectations", func(t *testing.T) {
		mockExec := dbxtesting.NewMockExecutor()

		// Note: In real tests, you'd typically use sqlmock.NewRows() for actual sql.Rows
		// For this test, we'll just verify the mock works with nil (test the mocking mechanism)
		mockExec.On("Query", "SELECT * FROM users").
			Return((*sql.Rows)(nil), nil)

		rows, err := mockExec.Query("SELECT * FROM users")

		assert.NoError(t, err)
		assert.Nil(t, rows) // Expected nil in this test case

		mockExec.AssertExpectations(t)
	})

	t.Run("QueryRow should work with mock expectations", func(t *testing.T) {
		mockExec := dbxtesting.NewMockExecutor()

		// Note: In real tests, you'd typically create actual *sql.Row
		// For this test, we'll just verify the mock works with nil
		mockExec.On("QueryRow", "SELECT name FROM users WHERE id = ?", 1).
			Return((*sql.Row)(nil))

		row := mockExec.QueryRow("SELECT name FROM users WHERE id = ?", 1)

		assert.Nil(t, row) // Expected nil in this test case

		mockExec.AssertExpectations(t)
	})

	t.Run("ExecContext should work with mock expectations", func(t *testing.T) {
		mockExec := dbxtesting.NewMockExecutor()
		result := dbxtesting.NewResult(2, 3)
		ctx := context.Background()

		mockExec.On("ExecContext", ctx, "UPDATE users SET name = ? WHERE id = ?", "Jane", 1).
			Return(result, nil)

		actualResult, err := mockExec.ExecContext(ctx, "UPDATE users SET name = ? WHERE id = ?", "Jane", 1)

		assert.NoError(t, err)
		assert.NotNil(t, actualResult)

		lastID, _ := actualResult.LastInsertId()
		rowsAffected, _ := actualResult.RowsAffected()
		assert.Equal(t, int64(2), lastID)
		assert.Equal(t, int64(3), rowsAffected)

		mockExec.AssertExpectations(t)
	})
}

func TestMockTransactor(t *testing.T) {
	t.Run("should implement Transactor interface", func(t *testing.T) {
		mockTx := dbxtesting.NewMockTransactor()

		// Verify it implements dbx.Transactor
		var _ dbx.Transactor = mockTx

		// Test transaction methods
		mockTx.On("Commit").Return(nil)
		mockTx.On("Rollback").Return(errors.New("rollback error"))

		err := mockTx.Commit()
		assert.NoError(t, err)

		err = mockTx.Rollback()
		assert.Error(t, err)
		assert.Equal(t, "rollback error", err.Error())

		mockTx.AssertExpectations(t)
	})

	t.Run("should also work as Executor", func(t *testing.T) {
		mockTx := dbxtesting.NewMockTransactor()
		result := dbxtesting.NewResult(5, 1)

		mockTx.On("Exec", "DELETE FROM users WHERE id = ?", 5).
			Return(result, nil)

		actualResult, err := mockTx.Exec("DELETE FROM users WHERE id = ?", 5)

		assert.NoError(t, err)
		assert.NotNil(t, actualResult)

		rowsAffected, _ := actualResult.RowsAffected()
		assert.Equal(t, int64(1), rowsAffected)

		mockTx.AssertExpectations(t)
	})
}

func TestMockDatabase(t *testing.T) {
	t.Run("should implement DatabaseWithContext interface", func(t *testing.T) {
		mockDB := dbxtesting.NewMockDatabase()

		// Verify it implements dbx.DatabaseWithContext
		var _ dbx.DatabaseWithContext = mockDB
	})

	t.Run("Context should create dbx.Context", func(t *testing.T) {
		mockDB := dbxtesting.NewMockDatabase()
		ctx := context.Background()

		dbxCtx := mockDB.Context(ctx)

		assert.NotNil(t, dbxCtx)
		assert.Equal(t, mockDB, dbxCtx.Executor())

		// Verify it implements dbx.Context
		var _ dbx.Context = dbxCtx
	})

	t.Run("should work with database operations", func(t *testing.T) {
		mockDB := dbxtesting.NewMockDatabase()
		result := dbxtesting.NewResult(10, 1)

		mockDB.On("Exec", "CREATE TABLE users (id INT PRIMARY KEY)").
			Return(result, nil)

		actualResult, err := mockDB.Exec("CREATE TABLE users (id INT PRIMARY KEY)")

		assert.NoError(t, err)
		assert.NotNil(t, actualResult)

		mockDB.AssertExpectations(t)
	})

	t.Run("Close should work", func(t *testing.T) {
		mockDB := dbxtesting.NewMockDatabase()

		mockDB.On("Close").Return(nil)

		err := mockDB.Close()

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("Begin should work", func(t *testing.T) {
		mockDB := dbxtesting.NewMockDatabase()

		mockDB.On("Begin").Return((*sql.Tx)(nil), nil)

		tx, err := mockDB.Begin()

		assert.NoError(t, err)
		assert.Nil(t, tx) // Expected nil in this test case
		mockDB.AssertExpectations(t)
	})

	t.Run("BeginTx should work", func(t *testing.T) {
		mockDB := dbxtesting.NewMockDatabase()
		ctx := context.Background()
		opts := &sql.TxOptions{}

		mockDB.On("BeginTx", ctx, opts).Return((*sql.Tx)(nil), nil)

		tx, err := mockDB.BeginTx(ctx, opts)

		assert.NoError(t, err)
		assert.Nil(t, tx) // Expected nil in this test case
		mockDB.AssertExpectations(t)
	})
}

func TestMockContext(t *testing.T) {
	t.Run("should delegate to parent context", func(t *testing.T) {
		parentCtx := context.Background()
		mockExec := dbxtesting.NewMockExecutor()
		mockCtx := dbxtesting.NewMockContext(parentCtx, mockExec)

		// Test context.Context methods
		assert.Equal(t, parentCtx.Done(), mockCtx.Done())
		assert.Equal(t, parentCtx.Err(), mockCtx.Err())

		deadline, ok := mockCtx.Deadline()
		expectedDeadline, expectedOk := parentCtx.Deadline()
		assert.Equal(t, expectedDeadline, deadline)
		assert.Equal(t, expectedOk, ok)
	})

	t.Run("should provide access to executor", func(t *testing.T) {
		parentCtx := context.Background()
		mockExec := dbxtesting.NewMockExecutor()
		mockCtx := dbxtesting.NewMockContext(parentCtx, mockExec)

		assert.Equal(t, mockExec, mockCtx.Executor())
	})
}

func TestMockResult(t *testing.T) {
	t.Run("NewResult should create proper result", func(t *testing.T) {
		result := dbxtesting.NewResult(42, 5)

		lastID, err := result.LastInsertId()
		assert.NoError(t, err)
		assert.Equal(t, int64(42), lastID)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(5), rowsAffected)
	})

	t.Run("NewResultWithError should handle errors", func(t *testing.T) {
		insertErr := errors.New("insert id error")
		affectedErr := errors.New("rows affected error")

		result := dbxtesting.NewResultWithError(0, 0, insertErr, affectedErr)

		lastID, err := result.LastInsertId()
		assert.Error(t, err)
		assert.Equal(t, insertErr, err)
		assert.Equal(t, int64(0), lastID)

		rowsAffected, err := result.RowsAffected()
		assert.Error(t, err)
		assert.Equal(t, affectedErr, err)
		assert.Equal(t, int64(0), rowsAffected)
	})
}

// Integration test demonstrating how to use the mocks with dbx.Transaction
func TestIntegrationWithDbxTransaction(t *testing.T) {
	t.Run("should work with dbx.Transaction", func(t *testing.T) {
		mockDB := dbxtesting.NewMockDatabase()
		mockTx := dbxtesting.NewMockTransactor()

		// This test demonstrates the mock usage pattern.
		// In real usage, you would set up expectations based on your specific use case.

		// Verify the mocks implement the expected interfaces
		var _ dbx.Database = mockDB
		var _ dbx.Transactor = mockTx

		assert.NotNil(t, mockDB)
		assert.NotNil(t, mockTx)
	})
}
