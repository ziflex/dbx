package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ziflex/dbx"
)

func TestTransaction(test *testing.T) {
	test.Run("should handle single transaction", func(t *testing.T) {
		dbMock, dmock, _ := sqlmock.New()
		defer dbMock.Close()

		ctx := context.Background()

		db := dbx.NewDatabase(dbMock)
		dmock.ExpectBegin()
		dmock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 3").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectCommit()

		err := dbx.Transaction(ctx, db, func(c dbx.Context) error {
			executor := c.Executor()

			if _, e := executor.Exec("SELECT 1"); e != nil {
				return e
			}

			if _, e := executor.Exec("SELECT 2"); e != nil {
				return e
			}

			if _, e := executor.Exec("SELECT 3"); e != nil {
				return e
			}

			return nil
		})

		assert.NoError(t, err)
	})

	test.Run("should handle tx begin errors", func(t *testing.T) {
		dbMock, dmock, _ := sqlmock.New()
		defer dbMock.Close()

		ctx := context.Background()

		testErr := errors.New("test error")
		db := dbx.NewDatabase(dbMock)
		dmock.ExpectBegin().WillReturnError(testErr)

		err := dbx.Transaction(ctx, db, func(c dbx.Context) error {
			executor := c.Executor()
			executor.Exec("SELECT 1")

			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	test.Run("should handle tx commit errors", func(t *testing.T) {
		dbMock, dmock, _ := sqlmock.New()
		defer dbMock.Close()

		ctx := context.Background()

		testErr := errors.New("test error")
		db := dbx.NewDatabase(dbMock)
		dmock.ExpectBegin().WillReturnError(testErr)
		dmock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectCommit().WillReturnError(testErr)

		err := dbx.Transaction(ctx, db, func(c dbx.Context) error {
			executor := c.Executor()
			executor.Exec("SELECT 1")

			return nil
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	test.Run("should handle single transaction and rollback on errors", func(t *testing.T) {
		dbMock, dmock, _ := sqlmock.New()
		defer dbMock.Close()

		ctx := context.Background()

		testErr := errors.New("test error")
		db := dbx.NewDatabase(dbMock)
		dmock.ExpectBegin()
		dmock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 3").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectRollback()

		err := dbx.Transaction(ctx, db, func(c dbx.Context) error {
			executor := c.Executor()
			executor.Exec("SELECT 1")
			executor.Exec("SELECT 2")
			executor.Exec("SELECT 3")

			return testErr
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	test.Run("should reuse nested transaction", func(t *testing.T) {
		dbMock, dmock, _ := sqlmock.New()
		defer dbMock.Close()

		ctx := context.Background()

		db := dbx.NewDatabase(dbMock)
		dmock.ExpectBegin()
		dmock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 3").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 4").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 5").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectExec("SELECT 6").WillReturnResult(sqlmock.NewResult(1, 1))
		dmock.ExpectCommit()

		err := dbx.Transaction(ctx, db, func(c1 dbx.Context) error {
			executor := c1.Executor()
			executor.Exec("SELECT 1")
			executor.Exec("SELECT 2")
			executor.Exec("SELECT 3")

			return dbx.Transaction(c1, db, func(c2 dbx.Context) error {
				executor2 := c2.Executor()
				executor2.Exec("SELECT 4")

				assert.Equal(t, c1, c2)
				assert.Equal(t, executor, executor2)

				return dbx.Transaction(c2, db, func(c3 dbx.Context) error {
					executor3 := c3.Executor()
					executor3.Exec("SELECT 5")

					assert.Equal(t, c2, c3)
					assert.Equal(t, executor2, executor3)

					return dbx.Transaction(c3, db, func(c4 dbx.Context) error {
						executor4 := c4.Executor()
						executor4.Exec("SELECT 6")

						assert.Equal(t, c3, c4)
						assert.Equal(t, executor3, executor4)

						return nil
					})
				})
			})
		})

		assert.NoError(t, err)
	})
}
