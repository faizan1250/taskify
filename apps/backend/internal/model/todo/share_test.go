package todo_test

import (
	"testing"

	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateShareToken(t *testing.T) {
	t.Run("produces 64-char hex and unique values", func(t *testing.T) {
		a, err := todo.GenerateShareToken()
		require.NoError(t, err)
		b, err := todo.GenerateShareToken()
		require.NoError(t, err)

		assert.Len(t, a, 2*todo.ShareTokenLength)
		assert.Len(t, b, 2*todo.ShareTokenLength)
		assert.NotEqual(t, a, b)
		assert.Regexp(t, `^[0-9a-f]+$`, a)
	})
}

func TestShareTodoPayload_Validate(t *testing.T) {
	t.Run("valid UUID passes", func(t *testing.T) {
		p := &todo.ShareTodoPayload{ID: uuid.New()}
		require.NoError(t, p.Validate())
	})

	t.Run("nil UUID fails", func(t *testing.T) {
		p := &todo.ShareTodoPayload{}
		require.Error(t, p.Validate())
	})
}

func TestUnshareTodoPayload_Validate(t *testing.T) {
	t.Run("valid UUID passes", func(t *testing.T) {
		p := &todo.UnshareTodoPayload{ID: uuid.New()}
		require.NoError(t, p.Validate())
	})

	t.Run("nil UUID fails", func(t *testing.T) {
		p := &todo.UnshareTodoPayload{}
		require.Error(t, p.Validate())
	})
}

func TestGetSharedTodoPayload_Validate(t *testing.T) {
	t.Run("valid token passes", func(t *testing.T) {
		token, err := todo.GenerateShareToken()
		require.NoError(t, err)
		p := &todo.GetSharedTodoPayload{Token: token}
		require.NoError(t, p.Validate())
	})

	t.Run("empty token fails", func(t *testing.T) {
		p := &todo.GetSharedTodoPayload{}
		require.Error(t, p.Validate())
	})

	t.Run("short token fails", func(t *testing.T) {
		p := &todo.GetSharedTodoPayload{Token: "abc"}
		require.Error(t, p.Validate())
	})
}
