# dbx/testing

This package provides mocks and testing utilities for dbx-based applications. It offers mock implementations of dbx interfaces to facilitate unit testing without requiring actual database connections or complex sqlmock setups.

## Features

- **Mock Implementations**: Complete mock implementations of all dbx interfaces
- **testify/mock Integration**: Uses the popular testify/mock framework for expectations
- **Easy Setup**: Simple helper functions to create mocks quickly
- **Interface Compatible**: Works with all dbx interfaces (Executor, Database, Context, etc.)

## Quick Start

### Basic Usage

```go
import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/ziflex/dbx"
    dbxtesting "github.com/ziflex/dbx/testing"
)

func TestUserService_CreateUser(t *testing.T) {
    // Create mock database
    mockDB := dbxtesting.NewMockDatabase()
    
    // Set up expectation
    expectedResult := dbxtesting.NewResult(123, 1) // lastInsertID=123, rowsAffected=1
    mockDB.On("Exec", "INSERT INTO users (name) VALUES (?)", "John").
        Return(expectedResult, nil)
    
    // Test your service
    service := NewUserService(mockDB)
    userID, err := service.CreateUser(context.Background(), "John")
    
    // Verify results
    assert.NoError(t, err)
    assert.Equal(t, int64(123), userID)
    mockDB.AssertExpectations(t)
}
```

### Testing Errors

```go
func TestUserService_CreateUser_Error(t *testing.T) {
    mockDB := dbxtesting.NewMockDatabase()
    
    // Set up expectation to return an error
    mockDB.On("Exec", "INSERT INTO users (name) VALUES (?)", "Invalid").
        Return(dbxtesting.NewResult(0, 0), errors.New("constraint violation"))
    
    service := NewUserService(mockDB)
    userID, err := service.CreateUser(context.Background(), "Invalid")
    
    assert.Error(t, err)
    assert.Equal(t, "constraint violation", err.Error())
    assert.Equal(t, int64(0), userID)
    mockDB.AssertExpectations(t)
}
```

### Testing Result Errors

```go
func TestUserService_ResultError(t *testing.T) {
    mockDB := dbxtesting.NewMockDatabase()
    
    // Create a result that returns an error for LastInsertId
    resultWithError := dbxtesting.NewResultWithError(0, 1, errors.New("no insert id"), nil)
    mockDB.On("Exec", "INSERT INTO users (name) VALUES (?)", "Test").
        Return(resultWithError, nil)
    
    service := NewUserService(mockDB)
    userID, err := service.CreateUser(context.Background(), "Test")
    
    assert.Error(t, err)
    assert.Equal(t, "no insert id", err.Error())
    mockDB.AssertExpectations(t)
}
```

## Available Mocks

### MockExecutor

Mocks the `dbx.Executor` interface with all database operation methods.

```go
mockExec := dbxtesting.NewMockExecutor()
mockExec.On("Exec", "DELETE FROM users WHERE id = ?", 1).
    Return(dbxtesting.NewResult(0, 1), nil)
```

### MockTransactor  

Mocks the `dbx.Transactor` interface, extending MockExecutor with transaction control.

```go
mockTx := dbxtesting.NewMockTransactor()
mockTx.On("Commit").Return(nil)
mockTx.On("Rollback").Return(errors.New("rollback failed"))
```

### MockDatabase

Mocks the `dbx.DatabaseWithContext` interface, providing full database functionality.

```go
mockDB := dbxtesting.NewMockDatabase()
mockDB.On("Close").Return(nil)
mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(&sql.Tx{}, nil)

// Automatically creates mock contexts
ctx := mockDB.Context(context.Background())
```

### MockContext

Mocks the `dbx.Context` interface, wrapping a parent context with mock executor access.

```go
parentCtx := context.Background()
mockExec := dbxtesting.NewMockExecutor()
mockCtx := dbxtesting.NewMockContext(parentCtx, mockExec)

// Use like any dbx.Context
executor := mockCtx.Executor()
```

## Helper Functions

### NewResult

Creates a simple mock result with specified values:

```go
result := dbxtesting.NewResult(lastInsertID, rowsAffected)
```

### NewResultWithError

Creates a mock result that can return errors:

```go
result := dbxtesting.NewResultWithError(
    lastInsertID, rowsAffected,
    lastInsertIDError, rowsAffectedError,
)
```

## Integration with testify/mock

All mocks use [testify/mock](https://github.com/stretchr/testify#mock-package) for expectations and verification:

```go
// Set up expectations
mock.On("Method", args...).Return(result, error)

// Verify all expectations were met
mock.AssertExpectations(t)

// Check if specific method was called
mock.AssertCalled(t, "Method", args...)

// Check if method was not called
mock.AssertNotCalled(t, "Method")
```

## Limitations

### QueryRow and Scan

The `QueryRow` method returns `*sql.Row`, which is difficult to mock effectively for `Scan` operations. For testing code that uses `QueryRow` with `Scan`, consider:

1. Using `go-sqlmock` directly for `QueryRow` operations
2. Testing at a higher integration level 
3. Restructuring code to use `Query` instead of `QueryRow` when possible

### Transaction Integration

While you can mock transaction methods, testing actual `dbx.Transaction` function calls requires careful setup of `BeginTx` expectations to return proper `*sql.Tx` instances. Consider using `go-sqlmock` for full transaction testing.

## Best Practices

1. **Use Interface Dependencies**: Design your services to depend on dbx interfaces rather than concrete types
2. **Test Business Logic**: Focus on testing your business logic rather than database operations
3. **Combine with Integration Tests**: Use mocks for unit tests and real databases for integration tests
4. **Verify Expectations**: Always call `AssertExpectations(t)` to ensure all expected calls were made

## Examples

See `example_test.go` for comprehensive examples of using the testing module in various scenarios.