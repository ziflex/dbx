# go-dbcontext
Custom context for DB operations

``go-dbcontext`` provides a thin layer of abstraction for communication between domain services and data repositories.

[API](https://godoc.org/github.com/ziflex/go-dbcontext)

## Examples


### User repository

```go
package users

import (
    "githib.com/ziflex/go-dbcontext"
)

type (
    User struct {
        ID int
        Name string
    }
    
    Repository struct {}
)

func (ur *UserRepository) Get(ctx dbcontext.Context, id int) (User, error) {
    
}
```