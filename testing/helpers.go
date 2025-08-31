package testing

import (
	"database/sql"
)

// NewMockExecutor creates a new MockExecutor instance.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

// NewMockTransactor creates a new MockTransactor instance.
func NewMockTransactor() *MockTransactor {
	return &MockTransactor{}
}

// NewMockDatabase creates a new MockDatabase instance.
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

// NewResult creates a MockResult with the specified last insert ID and rows affected.
// This is a convenience function for creating sql.Result mocks in tests.
//
// Parameters:
//   - lastInsertID: The last insert ID to return
//   - rowsAffected: The number of affected rows to return
//
// Returns:
//   - sql.Result: A mock result that can be used in test expectations
//
// Example:
//
//	result := testing.NewResult(1, 1) // ID=1, affected=1 row
//	mockDB.On("Exec", "INSERT INTO users (name) VALUES (?)", "John").
//		Return(result, nil)
func NewResult(lastInsertID, rowsAffected int64) sql.Result {
	return &MockResult{
		lastInsertID: lastInsertID,
		rowsAffected: rowsAffected,
	}
}

// NewResultWithError creates a MockResult that returns errors for LastInsertId or RowsAffected.
// This is useful for testing error conditions in database operations.
//
// Parameters:
//   - lastInsertID: The last insert ID to return
//   - rowsAffected: The number of affected rows to return
//   - insertIDErr: Error to return from LastInsertId() (nil for no error)
//   - affectedErr: Error to return from RowsAffected() (nil for no error)
//
// Returns:
//   - sql.Result: A mock result that can return errors
//
// Example:
//
//	result := testing.NewResultWithError(0, 0, errors.New("insert error"), nil)
//	mockDB.On("Exec", "INSERT INTO users (name) VALUES (?)", "John").
//		Return(result, nil)
func NewResultWithError(lastInsertID, rowsAffected int64, insertIDErr, affectedErr error) sql.Result {
	return &MockResult{
		lastInsertID: lastInsertID,
		rowsAffected: rowsAffected,
		insertIDErr:  insertIDErr,
		affectedErr:  affectedErr,
	}
}
