package dbx_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziflex/dbx"
)

func TestContextHelpers(t *testing.T) {
	t.Run("Is should return true for dbx context", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		db := dbx.New(mockDB)
		ctx := db.Context(context.Background())

		assert.True(t, dbx.Is(ctx))
	})

	t.Run("Is should return false for regular context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, dbx.Is(ctx))
	})

	t.Run("As should return dbx context when present", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

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
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		db := dbx.New(mockDB)
		dbCtx := db.Context(context.Background())

		// Embed dbx context into regular context
		regularCtx := context.Background()
		embeddedCtx := dbx.WithContext(regularCtx, dbCtx)

		// Extract it back
		extractedCtx := dbx.FromContext(embeddedCtx)
		assert.Equal(t, dbCtx, extractedCtx)
	})

	t.Run("FromContext should return nil for regular context", func(t *testing.T) {
		ctx := context.Background()
		extracted := dbx.FromContext(ctx)
		assert.Nil(t, extracted)
	})

	t.Run("FromContext should work with direct dbx context", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		db := dbx.New(mockDB)
		dbCtx := db.Context(context.Background())

		extracted := dbx.FromContext(dbCtx)
		assert.Equal(t, dbCtx, extracted)
	})
}

func TestDefaultContext(t *testing.T) {
	t.Run("context methods should delegate to parent", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

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

		// Test Err (should be nil since not cancelled/timed out)
		assert.NoError(t, dbCtx.Err())

		// Test Value
		key := "test-key"
		value := "test-value"
		parentWithValue := context.WithValue(parentCtx, key, value)
		dbCtxWithValue := dbx.NewContext(parentWithValue, db)

		assert.Equal(t, value, dbCtxWithValue.Value(key))
		assert.Nil(t, dbCtxWithValue.Value("nonexistent-key"))
	})

	t.Run("cancelled context should propagate cancellation", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

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
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		db := dbx.New(mockDB)
		ctx := dbx.NewContext(context.Background(), db)

		executor := ctx.Executor()
		assert.Equal(t, db, executor)
	})
}
