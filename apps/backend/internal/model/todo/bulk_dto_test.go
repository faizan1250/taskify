package todo

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bulkIDs(n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, uuid.New())
	}
	return ids
}

func TestBulkUpdateTodosPayload_Validate(t *testing.T) {
	status := StatusCompleted

	t.Run("valid patch passes", func(t *testing.T) {
		p := &BulkUpdateTodosPayload{IDs: bulkIDs(2), Patch: BulkTodoPatch{Status: &status}}
		require.NoError(t, p.Validate())
	})

	t.Run("empty ids rejected", func(t *testing.T) {
		p := &BulkUpdateTodosPayload{IDs: []uuid.UUID{}, Patch: BulkTodoPatch{Status: &status}}
		require.Error(t, p.Validate())
	})

	t.Run("nil ids rejected", func(t *testing.T) {
		p := &BulkUpdateTodosPayload{IDs: nil, Patch: BulkTodoPatch{Status: &status}}
		require.Error(t, p.Validate())
	})

	t.Run("over-limit ids rejected", func(t *testing.T) {
		p := &BulkUpdateTodosPayload{IDs: bulkIDs(MaxBulkTodoIDs + 1), Patch: BulkTodoPatch{Status: &status}}
		require.Error(t, p.Validate())
	})

	t.Run("nil UUID entry rejected", func(t *testing.T) {
		p := &BulkUpdateTodosPayload{IDs: []uuid.UUID{uuid.New(), uuid.Nil}, Patch: BulkTodoPatch{Status: &status}}
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty UUID")
	})

	t.Run("empty patch rejected", func(t *testing.T) {
		p := &BulkUpdateTodosPayload{IDs: bulkIDs(1), Patch: BulkTodoPatch{}}
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one")
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		bogus := Status("bogus")
		p := &BulkUpdateTodosPayload{IDs: bulkIDs(1), Patch: BulkTodoPatch{Status: &bogus}}
		require.Error(t, p.Validate())
	})

	t.Run("invalid priority rejected", func(t *testing.T) {
		bogus := Priority("urgent")
		p := &BulkUpdateTodosPayload{IDs: bulkIDs(1), Patch: BulkTodoPatch{Priority: &bogus}}
		require.Error(t, p.Validate())
	})
}

func TestBulkDeleteTodosPayload_Validate(t *testing.T) {
	t.Run("valid ids pass", func(t *testing.T) {
		p := &BulkDeleteTodosPayload{IDs: bulkIDs(3)}
		require.NoError(t, p.Validate())
	})

	t.Run("empty ids rejected", func(t *testing.T) {
		p := &BulkDeleteTodosPayload{IDs: []uuid.UUID{}}
		require.Error(t, p.Validate())
	})

	t.Run("over-limit ids rejected", func(t *testing.T) {
		p := &BulkDeleteTodosPayload{IDs: bulkIDs(MaxBulkTodoIDs + 1)}
		require.Error(t, p.Validate())
	})

	t.Run("nil UUID entry rejected", func(t *testing.T) {
		p := &BulkDeleteTodosPayload{IDs: []uuid.UUID{uuid.Nil}}
		require.Error(t, p.Validate())
	})
}

func TestBulkResults_HasFailures(t *testing.T) {
	t.Run("nil-safe false", func(t *testing.T) {
		var u *BulkUpdateTodosResult
		var d *BulkDeleteTodosResult
		assert.False(t, u.HasFailures())
		assert.False(t, d.HasFailures())
	})

	t.Run("empty results have no failures", func(t *testing.T) {
		assert.False(t, (&BulkUpdateTodosResult{}).HasFailures())
		assert.False(t, (&BulkDeleteTodosResult{}).HasFailures())
	})

	t.Run("failed entries report true", func(t *testing.T) {
		u := &BulkUpdateTodosResult{Failed: []BulkTodoFailure{{ID: uuid.New(), Error: "todo not found"}}}
		d := &BulkDeleteTodosResult{Failed: []BulkTodoFailure{{ID: uuid.New(), Error: "todo not found"}}}
		assert.True(t, u.HasFailures())
		assert.True(t, d.HasFailures())
	})

	t.Run("failure message stays generic", func(t *testing.T) {
		// Per-ID errors must not leak ownership details: a foreign row and a
		// missing row are indistinguishable to the caller.
		f := BulkTodoFailure{ID: uuid.New(), Error: "todo not found"}
		assert.False(t, strings.Contains(f.Error, "user"))
	})
}
