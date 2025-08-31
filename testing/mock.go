// Package testing provides mocks and testing utilities for dbx-based applications.
//
// This package offers mock implementations of dbx interfaces to facilitate
// unit testing without requiring actual database connections or complex
// sqlmock setups. The mocks are designed to be simple to use and integrate
// well with popular testing frameworks.
//
// Basic Usage:
//
//	// Create a mock database
//	mockDB := testing.NewMockDatabase()
//
//	// Configure expectations
//	mockDB.On("Exec", "INSERT INTO users (name) VALUES (?)", "John").
//		Return(testing.NewResult(1, 1), nil)
//
//	// Use in your code
//	ctx := mockDB.Context(context.Background())
//	result, err := ctx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
//
// Transaction Testing:
//
//	mockDB := testing.NewMockDatabase()
//	mockTx := testing.NewMockTransactor()
//
//	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
//	mockTx.On("Exec", "INSERT INTO users (name) VALUES (?)", "John").
//		Return(testing.NewResult(1, 1), nil)
//	mockTx.On("Commit").Return(nil)
//
//	err := dbx.Transaction(context.Background(), mockDB, func(ctx dbx.Context) error {
//		_, err := ctx.Executor().Exec("INSERT INTO users (name) VALUES (?)", "John")
//		return err
//	})
package testing

import (
	"context"
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/ziflex/dbx"
)

// MockExecutor is a mock implementation of dbx.Executor interface.
// It uses testify/mock for expectation setup and verification.
type MockExecutor struct {
	mock.Mock
}

// Exec mocks the Exec method.
func (m *MockExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	arguments := []interface{}{query}
	arguments = append(arguments, args...)
	called := m.Called(arguments...)
	return called.Get(0).(sql.Result), called.Error(1)
}

// Query mocks the Query method.
func (m *MockExecutor) Query(query string, args ...interface{}) (*sql.Rows, error) {
	arguments := []interface{}{query}
	arguments = append(arguments, args...)
	called := m.Called(arguments...)
	return called.Get(0).(*sql.Rows), called.Error(1)
}

// QueryRow mocks the QueryRow method.
func (m *MockExecutor) QueryRow(query string, args ...interface{}) *sql.Row {
	arguments := []interface{}{query}
	arguments = append(arguments, args...)
	called := m.Called(arguments...)
	return called.Get(0).(*sql.Row)
}

// ExecContext mocks the ExecContext method.
func (m *MockExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	arguments := []interface{}{ctx, query}
	arguments = append(arguments, args...)
	called := m.Called(arguments...)
	return called.Get(0).(sql.Result), called.Error(1)
}

// QueryContext mocks the QueryContext method.
func (m *MockExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	arguments := []interface{}{ctx, query}
	arguments = append(arguments, args...)
	called := m.Called(arguments...)
	return called.Get(0).(*sql.Rows), called.Error(1)
}

// QueryRowContext mocks the QueryRowContext method.
func (m *MockExecutor) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	arguments := []interface{}{ctx, query}
	arguments = append(arguments, args...)
	called := m.Called(arguments...)
	return called.Get(0).(*sql.Row)
}

// MockTransactor is a mock implementation of dbx.Transactor interface.
// It embeds MockExecutor and adds transaction-specific methods.
type MockTransactor struct {
	MockExecutor
}

// Commit mocks the Commit method.
func (m *MockTransactor) Commit() error {
	called := m.Called()
	return called.Error(0)
}

// Rollback mocks the Rollback method.
func (m *MockTransactor) Rollback() error {
	called := m.Called()
	return called.Error(0)
}

// MockDatabase is a mock implementation of dbx.DatabaseWithContext interface.
// It provides all database operations and context creation capabilities.
type MockDatabase struct {
	MockExecutor
}

// Close mocks the Close method.
func (m *MockDatabase) Close() error {
	called := m.Called()
	return called.Error(0)
}

// Begin mocks the Begin method.
func (m *MockDatabase) Begin() (*sql.Tx, error) {
	called := m.Called()
	return called.Get(0).(*sql.Tx), called.Error(1)
}

// BeginTx mocks the BeginTx method.
func (m *MockDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	called := m.Called(ctx, opts)
	return called.Get(0).(*sql.Tx), called.Error(1)
}

// Context creates a mock context using this database as executor.
func (m *MockDatabase) Context(ctx context.Context) dbx.Context {
	return NewMockContext(ctx, m)
}

// MockContext is a mock implementation of dbx.Context interface.
// It wraps a standard Go context and provides access to a mock executor.
type MockContext struct {
	parent   context.Context
	executor dbx.Executor
}

// NewMockContext creates a new MockContext with the given parent context and executor.
func NewMockContext(parent context.Context, executor dbx.Executor) *MockContext {
	return &MockContext{
		parent:   parent,
		executor: executor,
	}
}

// Deadline delegates to the parent context.
func (m *MockContext) Deadline() (deadline time.Time, ok bool) {
	return m.parent.Deadline()
}

// Done delegates to the parent context.
func (m *MockContext) Done() <-chan struct{} {
	return m.parent.Done()
}

// Err delegates to the parent context.
func (m *MockContext) Err() error {
	return m.parent.Err()
}

// Value delegates to the parent context.
func (m *MockContext) Value(key interface{}) interface{} {
	return m.parent.Value(key)
}

// Executor returns the mock executor.
func (m *MockContext) Executor() dbx.Executor {
	return m.executor
}

// MockResult is a mock implementation of sql.Result interface.
type MockResult struct {
	lastInsertID int64
	rowsAffected int64
	insertIDErr  error
	affectedErr  error
}

// LastInsertId returns the mock last insert ID.
func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastInsertID, m.insertIDErr
}

// RowsAffected returns the mock rows affected count.
func (m *MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, m.affectedErr
}
