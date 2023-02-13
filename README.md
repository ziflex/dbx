# dbx
This Go library provides a light abstraction on top of the sql package from the Go standard library. 
It provides convenient functions for working with SQL databases.

[API](https://godoc.org/github.com/ziflex/dbx)

## Installation
```bash
  go get github.com/ziflex/dbx@latest
```

## Quick Start
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

func getUserNames(ctx dbx.Context) []string {
	// Create a new query
	executor := ctx.Executor()
    
    // Execute the query
    rows, err := executor.Query("SELECT name FROM users")
    
	if err != nil {
        log.Fatal(err)
    }
    
    // Iterate over the rows
    var names []string
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            log.Fatal(err)
        }
        
        names = append(names, name)
    }
    
    return names
}

func main() {
	// Connect to a database
	db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")
	
	if err != nil {
		fmt.Println(err)
		return
	}

	// Wrap the *sql.DB object with dbx.NewDatabase
	dbxDB := dbx.New(db)
	
	userNames := getUserNames(dbxDB.Context(context.Background()))
	
	fmt.Println(userNames)
}
```

## Context
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

func getUserNames(ctx context.Context) []string {
    // Get dbx context from the context
    dbxContext := dbx.FromContext(ctx)
	
    if dbxContext == nil {
        log.Fatal("dbx context is not found")
    }
	
    // Create a new query
    executor := dbxContext.Executor()
    
    // Execute the query
    rows, err := executor.Query("SELECT name FROM users")
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Iterate over the rows
    var names []string
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            log.Fatal(err)
        }
        
        names = append(names, name)
    }
    
    return names
}

func main() {
    // Connect to a database
    db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")

    if err != nil {
        fmt.Println(err)
        return
    }

    // Wrap the *sql.DB object with dbx.NewDatabase
    dbxDB := dbx.New(db)
    
    // Create a new context
    ctx := context.Background()
    
    // Add dbx context to the context
    ctx = dbxDB.Context(ctx)
    
    userNames := getUserNames(ctx)
    
    fmt.Println(userNames)
}
```

## Transactions

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

func insertUser(ctx dbx.Context, name string) {
    // Create a new query
    executor := ctx.Executor()
    
    // Execute the query
    _, err := executor.Exec("INSERT INTO users (name) VALUES ($1)", name)
    
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    // Connect to a database
    db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")

    if err != nil {
        fmt.Println(err)
        return
    }

    // Wrap the *sql.DB object with dbx.NewDatabase
    dbxDB := dbx.New(db)
    
    err = dbx.Transaction(context.Background(), dbxDB, func(ctx dbx.Context) error {
        insertUser(ctx, "John")
        insertUser(ctx, "Doe")
            
        return nil
    })

    if err != nil {
        fmt.Println(err)
        return
    }
}
```

> Transactions are reusable by default. Using ``dbx.Transaction`` multiple times within the same transaction will not create a new transaction. 
> To disable this behavior, use ``dbx.WithNewTransaction`` option. 

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

func insertUser(ctx dbx.Context, name string) {
    // Create a new query
    executor := ctx.Executor()

    // Execute the query
    _, err := executor.Exec("INSERT INTO users (name) VALUES ($1)", name)
    
    if err != nil {
        log.Fatal(err)
    }
}

func insertCompany(ctx dbx.Context, name string) {
    // Create a new query
    executor := ctx.Executor()

    // Execute the query
    _, err := executor.Exec("INSERT INTO companies (name) VALUES ($1)", name)
    
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    // Connect to a database
    db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")

    if err != nil {
        fmt.Println(err)
        return
    }

    // Wrap the *sql.DB object with dbx.NewDatabase
    dbxDB := dbx.New(db)
	
    err = dbx.Transaction(context.Background(), dbxDB, func(ctx dbx.Context) error {
        insertUser(ctx, "John")
        insertUser(ctx, "Doe")
    
        return dbx.Transaction(ctx, dbxDB, func(ctx dbx.Context) error {
            insertCompany(ctx, "Google")
            insertCompany(ctx, "Apple")
            
            return nil
        }, dbx.WithNewTransaction())
    })
	
    if err != nil {
        fmt.Println(err)
        return
    }
}
```