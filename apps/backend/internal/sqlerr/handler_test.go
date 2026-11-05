package sqlerr_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/sqlerr"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fkError(table, column, constraint string) *pgconn.PgError {
	return &pgconn.PgError{
		Severity:       "ERROR",
		Code:           "23503",
		Message:        `insert or update on table "todos" violates foreign key constraint`,
		TableName:      table,
		ColumnName:     column,
		ConstraintName: constraint,
	}
}

func requireHTTPStatus(t *testing.T, err error, status int) *errs.HTTPError {
	t.Helper()
	require.Error(t, err)
	var httpErr *errs.HTTPError
	require.True(t, errors.As(err, &httpErr), "expected *errs.HTTPError, got %T", err)
	assert.Equal(t, status, httpErr.Status)
	return httpErr
}

func TestHandleError_ForeignKeyViolationMapsToNotFound(t *testing.T) {
	httpErr := requireHTTPStatus(t, sqlerr.HandleError(fkError("todos", "category_id", "todos_category_id_fkey")),
		http.StatusNotFound)
	assert.Equal(t, "The referenced Category does not exist", httpErr.Message)
	assert.Equal(t, "TODO_NOT_FOUND", httpErr.Code)
}

func TestHandleError_ForeignKeyViolationWrappedMapsToNotFound(t *testing.T) {
	wrapped := fmt.Errorf("failed to execute create todo query: %w", fkError("todos", "parent_todo_id", "todos_parent_todo_id_fkey"))

	httpErr := requireHTTPStatus(t, sqlerr.HandleError(wrapped), http.StatusNotFound)
	assert.Equal(t, "The referenced Parent Todo does not exist", httpErr.Message)
	assert.Equal(t, "TODO_NOT_FOUND", httpErr.Code)
}

func TestHandleError_ForeignKeyViolationWithoutColumnFallsBackToTable(t *testing.T) {
	httpErr := requireHTTPStatus(t, sqlerr.HandleError(fkError("todos", "", "todos_category_id_fkey")),
		http.StatusNotFound)
	assert.Equal(t, "The referenced Todo does not exist", httpErr.Message)
}

func TestHandleError_UniqueViolationStillBadRequest(t *testing.T) {
	err := sqlerr.HandleError(&pgconn.PgError{
		Severity:       "ERROR",
		Code:           "23505",
		Message:        `duplicate key value violates unique constraint`,
		TableName:      "todos",
		ConstraintName: "unique_todos_title",
	})

	requireHTTPStatus(t, err, http.StatusBadRequest)
}
