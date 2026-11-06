package service

import (
	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ShareTodo mints (or returns the existing) bearer token for a caller-owned
// todo. Object-level authz: the todo must belong to userID, enforced via
// CheckTodoExists before any token is created.
func (s *TodoService) ShareTodo(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.ShareTodoResult, error) {
	logger := middleware.GetLogger(ctx)

	if _, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID); err != nil {
		logger.Error().Err(err).Msg("share todo validation failed")
		return nil, err
	}

	share, err := s.todoRepo.CreateShareToken(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create share token")
		return nil, err
	}

	middleware.GetLogger(ctx).Info().
		Str("event", "todo_shared").
		Str("todo_id", todoID.String()).
		Msg("Todo shared")

	return &todo.ShareTodoResult{Token: share.Token, TodoID: share.TodoID}, nil
}

// GetSharedTodo resolves a public bearer token to its todo. No userID: token
// possession is the authorization. Never logs the token value.
func (s *TodoService) GetSharedTodo(ctx echo.Context, token string) (*todo.PopulatedTodo, error) {
	logger := middleware.GetLogger(ctx)

	shared, err := s.todoRepo.GetSharedTodo(ctx.Request().Context(), token)
	if err != nil {
		logger.Error().Msg("failed to fetch shared todo")
		return nil, err
	}

	return shared, nil
}

// UnshareTodo revokes every active token for a caller-owned todo.
// Object-level authz: the todo must belong to userID.
func (s *TodoService) UnshareTodo(ctx echo.Context, userID string, todoID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	if _, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID); err != nil {
		logger.Error().Err(err).Msg("unshare todo validation failed")
		return err
	}

	if err := s.todoRepo.RevokeShareTokensForTodo(ctx.Request().Context(), userID, todoID); err != nil {
		logger.Error().Err(err).Msg("failed to revoke share tokens")
		return err
	}

	middleware.GetLogger(ctx).Info().
		Str("event", "todo_unshared").
		Str("todo_id", todoID.String()).
		Msg("Todo unshared")

	return nil
}
