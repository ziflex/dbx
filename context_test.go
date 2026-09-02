package dbx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ziflex/dbx"
)

func TestContextHelpers(t *testing.T) {
	t.Run("Is should return true for dbx context", func(t *testing.T) {
		mockDB, _ := newSQLMock(t)

		db := dbx.New(mockDB)
		ctx := db.Context(context.Background())

		assert.True(t, dbx.Is(ctx))
	})

	t.Run("Is should return false for regular context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, dbx.Is(ctx))
	})

	t.Run("As should return dbx context when present", func(t *testing.T) {
		mockDB, _ := newSQLMock(t)

		db := dbx.New(mockDB)
		originalCtx := db.Context(context.Background())

		extractedCtx, ok := dbx.As(originalCtx)
		assert.True(t, ok)
		assert.Equal(t, originalCtx, extractedCtx)
	})

	t.Run("As should return false for regular context", func(t *testing.T) {
		ctx := context.Background()

		extractedCtx, ok := dbx.As(ctx)
		assert.False(t, ok)
		assert.Nil(t, extractedCtx)
	})

	t.Run("WithContext and FromContext should work together", func(t *testing.T) {
		mockDB, _ := newSQLMock(t)

		db := dbx.New(mockDB)
		dbCtx := db.Context(context.Background())

		// Embed dbx context into regular context
		regularCtx := context.Background()
		embeddedCtx := dbx.WithContext(regularCtx, dbCtx)

		// Extract it back
		extractedCtx := dbx.FromContext(embeddedCtx)
		assert.Equal(t, dbCtx, extractedCtx)
		assert.False(t, dbx.Is(embeddedCtx))
		assert.NotEqual(t, dbCtx, embeddedCtx)
		_, ok := dbx.As(embeddedCtx)
		assert.False(t, ok)
	})

	t.Run("FromContext should return nil for regular context", func(t *testing.T) {
		ctx := context.Background()
		extracted := dbx.FromContext(ctx)
		assert.Nil(t, extracted)
	})

	t.Run("FromContext should work with direct dbx context", func(t *testing.T) {
		mockDB, _ := newSQLMock(t)

		db := dbx.New(mockDB)
		dbCtx := db.Context(context.Background())

		extracted := dbx.FromContext(dbCtx)
		assert.Equal(t, dbCtx, extracted)
	})
}

func TestNewContextFrom(t *testing.T) {
	t.Run("returns an existing context", func(t *testing.T) {
		database, _ := newSQLMock(t)
		existing := dbx.NewContext(context.Background(), dbx.New(database))

		actual := dbx.NewContextFrom(existing, struct{}{})

		assert.Same(t, existing, actual)
	})

	t.Run("creates a context with a ContextCreator", func(t *testing.T) {
		database, _ := newSQLMock(t)
		creator := dbx.New(database)

		actual := dbx.NewContextFrom(context.Background(), creator)

		assert.Equal(t, creator, actual.Executor())
	})

	t.Run("creates a context with a Database", func(t *testing.T) {
		database, _ := newSQLMock(t)
		input := struct{ dbx.Database }{Database: dbx.New(database)}

		actual := dbx.NewContextFrom(context.Background(), input)

		assert.Equal(t, input, actual.Executor())
	})

	t.Run("creates a context with a Transactor", func(t *testing.T) {
		database, mock := newSQLMock(t)
		mock.ExpectBegin()
		mock.ExpectRollback()

		tx, err := database.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		actual := dbx.NewContextFrom(context.Background(), tx)
		assert.Equal(t, tx, actual.Executor())
		require.NoError(t, tx.Rollback())
	})

	t.Run("panics for unsupported input", func(t *testing.T) {
		assert.PanicsWithValue(t, "input must implement ContextCreator, Database, or Transactor", func() {
			dbx.NewContextFrom(context.Background(), struct{}{})
		})
	})
}

func TestDefaultContext(t *testing.T) {
	t.Run("context methods should delegate to parent", func(t *testing.T) {
		mockDB, _ := newSQLMock(t)

		// Create a context with deadline
		deadline := time.Now().Add(5 * time.Second)
		parentCtx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		db := dbx.New(mockDB)
		dbCtx := dbx.NewContext(parentCtx, db)

		// Test Deadline
		ctxDeadline, ok := dbCtx.Deadline()
		assert.True(t, ok)
		assert.Equal(t, deadline, ctxDeadline)

		// Test Done channel
		select {
		case <-dbCtx.Done():
			t.Error("Done channel should not be closed yet")
		default:
			// Expected
		}

		// Test Err (should be nil since not canceled/timed out)
		assert.NoError(t, dbCtx.Err())

		// Test Value
		key := "test-key"
		value := "test-value"
		parentWithValue := context.WithValue(parentCtx, key, value)
		dbCtxWithValue := dbx.NewContext(parentWithValue, db)

		assert.Equal(t, value, dbCtxWithValue.Value(key))
		assert.Nil(t, dbCtxWithValue.Value("nonexistent-key"))
	})

	t.Run("canceled context should propagate cancellation", func(t *testing.T) {
		mockDB, _ := newSQLMock(t)

		parentCtx, cancel := context.WithCancel(context.Background())
		db := dbx.New(mockDB)
		dbCtx := dbx.NewContext(parentCtx, db)

		// Cancel parent context
		cancel()

		// Check that cancellation is propagated
		assert.Error(t, dbCtx.Err())
		assert.Equal(t, context.Canceled, dbCtx.Err())

		// Check Done channel is closed
		select {
		case <-dbCtx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("Done channel should be closed after parent cancellation")
		}
	})

	t.Run("Executor should return the provided executor", func(t *testing.T) {
		mockDB, _ := newSQLMock(t)

		db := dbx.New(mockDB)
		ctx := dbx.NewContext(context.Background(), db)

		executor := ctx.Executor()
		assert.Equal(t, db, executor)
	})
}
