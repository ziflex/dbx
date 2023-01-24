package dbx

import (
	"context"
	"database/sql"
	"time"
)

type DefaultContext struct {
	parent context.Context
	db     *sql.DB
	tx     *sql.Tx
}

// New returns a new context with a given DB
func New(parent context.Context, db *sql.DB) Context {
	return &DefaultContext{
		parent: parent,
		db:     db,
	}
}

// WithTx returns a new context with a given TX
func WithTx(parent context.Context, tx *sql.Tx) Context {
	return &DefaultContext{
		parent: parent,
		tx:     tx,
	}
}

func (c *DefaultContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

func (c *DefaultContext) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *DefaultContext) Err() error {
	return c.parent.Err()
}

func (c *DefaultContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

func (c *DefaultContext) Executor() Executor {
	if c.tx == nil {
		return c.db
	}

	return c.tx
}
