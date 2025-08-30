# dbx

A lightweight, context-aware abstraction layer for Go's `database/sql` package that simplifies database operations and transaction management.

[![API Documentation](https://godoc.org/github.com/ziflex/dbx?status.svg)](https://godoc.org/github.com/ziflex/dbx)

## Table of Contents
- [Why dbx?](#why-dbx)
- [Design Philosophy](#design-philosophy)
- [Installation](#installation)
- [Key Concepts](#key-concepts)
- [Quick Start](#quick-start)
- [Working with Contexts](#working-with-contexts)
- [Transaction Management](#transaction-management)
- [Advanced Usage](#advanced-usage)
- [Testing](#testing)
- [API Reference](#api-reference)

## Why dbx?

The standard `database/sql` package is powerful but requires boilerplate code for common patterns. `dbx` addresses several pain points:

- **Context Management**: Eliminates the need to pass both `context.Context` and database connections separately
- **Transaction Handling**: Automatic transaction lifecycle management with support for nested transactions
- **Unified Interface**: Same API for both direct database operations and transactions
- **Testing**: Easier to mock and test database operations
- **Clean Architecture**: Promotes separation of concerns between business logic and data access

## Design Philosophy

`dbx` follows these core principles:

1. **Context-Driven**: Database connections and transactions are embedded within Go contexts
2. **Interface-Based**: Uses interfaces for maximum flexibility and testability
3. **Zero Magic**: Predictable behavior with no hidden surprises
4. **Minimal Overhead**: Thin layer that doesn't compromise performance
5. **Standard Library Compatible**: Works seamlessly with existing `database/sql` code

## Installation

```bash
go get github.com/ziflex/dbx@latest
```

## Key Concepts

### Database Interface
The `Database` interface wraps a `*sql.DB` and provides context creation:
```go
type Database interface {
    io.Closer
    ContextCreator  // Creates dbx.Context
    Beginner       // Begins transactions
    Executor       // Executes queries directly
}
```

### Context Interface
The `Context` interface extends Go's `context.Context` with database execution capabilities:
```go
type Context interface {
    context.Context
    Executor() Executor  // Returns sql.DB or sql.Tx depending on transaction state
}
```

### Executor Interface
The `Executor` interface abstracts both `*sql.DB` and `*sql.Tx` operations:
```go
type Executor interface {
    Exec(query string, args ...interface{}) (sql.Result, error)
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    // ... context variants
}
```

This design allows your functions to work with both direct database connections and transactions without modification.

## Quick Start

Here's a complete example showing basic database operations:

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
	
	_ "github.com/lib/pq"
    "github.com/ziflex/dbx"
)

// User represents a user record
type User struct {
    ID   int
    Name string
}

// getUserNames demonstrates querying with dbx.Context
func getUserNames(ctx dbx.Context) ([]User, error) {
    executor := ctx.Executor()
    
    rows, err := executor.Query("SELECT id, name FROM users ORDER BY name")
    if err != nil {
        return nil, fmt.Errorf("failed to query users: %w", err)
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name); err != nil {
            return nil, fmt.Errorf("failed to scan user: %w", err)
        }
        users = append(users, user)
    }
    
    return users, rows.Err()
}

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Wrap with dbx
    dbxDB := dbx.New(db)
    
    // Create dbx context and query
    users, err := getUserNames(dbxDB.Context(context.Background()))
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d users\n", len(users))
    for _, user := range users {
        fmt.Printf("- %s (ID: %d)\n", user.Name, user.ID)
    }
}
```

### Key Benefits Demonstrated:
- **Single Parameter**: Functions only need `dbx.Context` instead of separate context and database parameters
- **Consistent Interface**: Same API works for both direct DB operations and transactions
- **Better Error Handling**: Proper error wrapping and handling patterns

## Working with Contexts

`dbx` provides multiple ways to work with contexts, allowing flexibility in your application architecture.

### Direct Context Creation
Create a dbx context directly from a Database:

```go
func directExample() {
    db := dbx.New(sqlDB)
    ctx := db.Context(context.Background())
    
    // Use ctx for database operations
    result, err := getUserCount(ctx)
}

func getUserCount(ctx dbx.Context) (int, error) {
    var count int
    err := ctx.Executor().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    return count, err
}
```

### Context Extraction Pattern
Extract dbx context from standard Go context for cleaner service layers:

```go
func serviceLayerExample(ctx context.Context) {
    // Extract dbx context from regular context
    dbxCtx := dbx.FromContext(ctx)
    if dbxCtx == nil {
        log.Fatal("database context not found")
    }
    
    users, err := getUserNames(dbxCtx)
    // ... handle results
}

func main() {
    db := dbx.New(sqlDB)
    ctx := context.Background()
    
    // Embed dbx context into regular context
    ctx = dbx.WithContext(ctx, db.Context(ctx))
    
    serviceLayerExample(ctx)
}
```

### Context Helper Functions
- `dbx.Is(ctx)` - Check if context contains dbx context
- `dbx.As(ctx)` - Extract dbx context with ok flag
- `dbx.FromContext(ctx)` - Extract dbx context (returns nil if not found)
- `dbx.WithContext(ctx, dbxCtx)` - Embed dbx context into regular context

## Transaction Management

`dbx` provides powerful transaction management with automatic lifecycle handling and support for nested operations.

### Basic Transactions

```go
func createUserWithProfile(ctx context.Context, db dbx.Database, userName, email string) error {
    return dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
        // Insert user
        result, err := txCtx.Executor().Exec(
            "INSERT INTO users (name) VALUES ($1) RETURNING id", userName)
        if err != nil {
            return fmt.Errorf("failed to insert user: %w", err)
        }
        
        var userID int64
        userID, err = result.LastInsertId()
        if err != nil {
            return fmt.Errorf("failed to get user ID: %w", err)
        }
        
        // Insert profile
        _, err = txCtx.Executor().Exec(
            "INSERT INTO profiles (user_id, email) VALUES ($1, $2)", userID, email)
        if err != nil {
            return fmt.Errorf("failed to insert profile: %w", err)
        }
        
        return nil
    })
}
```

### Transaction Reuse (Default Behavior)

By default, `dbx.Transaction` reuses existing transactions. This prevents unnecessary nesting:

```go
func processOrder(ctx dbx.Context, orderID int) error {
    // This function works both in and outside transactions
    return dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
        if err := updateInventory(txCtx, orderID); err != nil {
            return err
        }
        
        return updateOrderStatus(txCtx, orderID, "processed")
    })
}

func updateInventory(ctx dbx.Context, orderID int) error {
    // This also uses Transaction, but will reuse the existing one
    return dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
        // Inventory updates here
        return nil
    })
}
```

### Transactions with Return Values

Use `TransactionWithResult` when you need to return values from transactions:

```go
func createUserAndGetID(ctx context.Context, db dbx.Database, name string) (int64, error) {
    return dbx.TransactionWithResult(ctx, db, func(txCtx dbx.Context) (int64, error) {
        result, err := txCtx.Executor().Exec(
            "INSERT INTO users (name) VALUES ($1)", name)
        if err != nil {
            return 0, err
        }
        
        return result.LastInsertId()
    })
}
```

## Advanced Usage

### Transaction Options

Control transaction behavior with options:

```go
// Read-only transaction
err := dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
    // Only SELECT operations allowed
    return generateReport(txCtx)
}, dbx.WithReadOnly(true))

// Custom isolation level
err := dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
    return performCriticalOperation(txCtx)
}, dbx.WithIsolationLevel(sql.LevelSerializable))

// Force new transaction (disable reuse)
err := dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
    return independentOperation(txCtx)
}, dbx.WithNewTransaction())
```

### Error Handling Patterns

`dbx` automatically handles transaction rollback on errors:

```go
func transferFunds(ctx context.Context, db dbx.Database, fromID, toID int, amount decimal.Decimal) error {
    return dbx.Transaction(ctx, db, func(txCtx dbx.Context) error {
        // Debit source account
        result, err := txCtx.Executor().Exec(
            "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1", 
            amount, fromID)
        if err != nil {
            return fmt.Errorf("failed to debit account %d: %w", fromID, err)
        }
        
        rowsAffected, err := result.RowsAffected()
        if err != nil {
            return fmt.Errorf("failed to check debit result: %w", err)
        }
        if rowsAffected == 0 {
            return fmt.Errorf("insufficient funds in account %d", fromID)
        }
        
        // Credit destination account
        _, err = txCtx.Executor().Exec(
            "UPDATE accounts SET balance = balance + $1 WHERE id = $2", 
            amount, toID)
        if err != nil {
            return fmt.Errorf("failed to credit account %d: %w", toID, err)
        }
        
        // Any error here will automatically rollback the entire transaction
        return nil
    })
}
```

### Working with Prepared Statements

Since `dbx.Context.Executor()` returns the underlying `sql.DB` or `sql.Tx`, you can use prepared statements:

```go
func batchInsertUsers(ctx dbx.Context, users []User) error {
    executor := ctx.Executor()
    
    // Prepare statement (works with both DB and Tx)
    stmt, err := executor.Prepare("INSERT INTO users (name, email) VALUES ($1, $2)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, user := range users {
        if _, err := stmt.Exec(user.Name, user.Email); err != nil {
            return fmt.Errorf("failed to insert user %s: %w", user.Name, err)
        }
    }
    
    return nil
}
```

## Testing

`dbx` works seamlessly with testing frameworks and mocking libraries:

### Using go-sqlmock

```go
func TestGetUserNames(t *testing.T) {
    // Create mock database
    mockDB, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer mockDB.Close()
    
    // Setup expectations
    rows := sqlmock.NewRows([]string{"id", "name"}).
        AddRow(1, "Alice").
        AddRow(2, "Bob")
    mock.ExpectQuery("SELECT id, name FROM users").WillReturnRows(rows)
    
    // Test with dbx
    db := dbx.New(mockDB)
    users, err := getUserNames(db.Context(context.Background()))
    
    require.NoError(t, err)
    assert.Len(t, users, 2)
    assert.Equal(t, "Alice", users[0].Name)
    assert.Equal(t, "Bob", users[1].Name)
    
    // Verify all expectations met
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Testing Transactions

```go
func TestTransferFunds(t *testing.T) {
    mockDB, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer mockDB.Close()
    
    // Setup transaction expectations
    mock.ExpectBegin()
    mock.ExpectExec("UPDATE accounts SET balance").
        WithArgs(100, 1, 100).
        WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectExec("UPDATE accounts SET balance").
        WithArgs(100, 2).
        WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectCommit()
    
    db := dbx.New(mockDB)
    err = transferFunds(context.Background(), db, 1, 2, decimal.NewFromInt(100))
    
    require.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

## API Reference

### Core Functions

- `dbx.New(db *sql.DB) Database` - Creates a new dbx Database wrapper
- `dbx.Transaction(ctx context.Context, db Database, op Operation, opts ...Option) error` - Executes operation in transaction
- `dbx.TransactionWithResult[T](ctx context.Context, db Database, op OperationWithResult[T], opts ...Option) (T, error)` - Executes operation in transaction with return value

### Context Functions

- `dbx.FromContext(ctx context.Context) Context` - Extract dbx context from context
- `dbx.WithContext(ctx context.Context, dbxCtx Context) context.Context` - Embed dbx context
- `dbx.Is(ctx context.Context) bool` - Check if context contains dbx context
- `dbx.As(ctx context.Context) (Context, bool)` - Extract dbx context with ok flag

### Transaction Options

- `dbx.WithIsolationLevel(level sql.IsolationLevel)` - Set transaction isolation level
- `dbx.WithReadOnly(readOnly bool)` - Set read-only flag
- `dbx.WithNewTransaction()` - Force creation of new transaction (disable reuse)

For complete API documentation, see [GoDoc](https://godoc.org/github.com/ziflex/dbx).