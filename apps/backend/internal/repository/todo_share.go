package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateShareToken mints a bearer token for todoID. Idempotent: an existing
// active token for the same todo is returned instead of minting a second one,
// so share retries never leak multiple live capabilities.
func (r *TodoRepository) CreateShareToken(ctx context.Context, userID string, todoID uuid.UUID) (*todo.TodoShareToken, error) {
	existingStmt := `
		SELECT
			*
		FROM
			todo_share_tokens
		WHERE
			todo_id = @todo_id
			AND revoked_at IS NULL
		ORDER BY
			created_at ASC
		LIMIT 1
	`
	rows, err := r.server.DB.Pool.Query(ctx, existingStmt, pgx.NamedArgs{
		"todo_id": todoID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check existing share token for todo_id=%s: %w", todoID.String(), err)
	}
	existing, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.TodoShareToken])
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to collect existing share token for todo_id=%s: %w", todoID.String(), err)
	}

	token, err := todo.GenerateShareToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate share token for todo_id=%s: %w", todoID.String(), err)
	}

	insertStmt := `
		INSERT INTO
			todo_share_tokens (token, todo_id, created_by)
		VALUES
			(@token, @todo_id, @created_by)
		RETURNING
			*
	`
	insertRows, err := r.server.DB.Pool.Query(ctx, insertStmt, pgx.NamedArgs{
		"token":      token,
		"todo_id":    todoID,
		"created_by": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create share token query for todo_id=%s: %w", todoID.String(), err)
	}

	share, err := pgx.CollectOneRow(insertRows, pgx.RowToStructByName[todo.TodoShareToken])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:todo_share_tokens for todo_id=%s: %w", todoID.String(), err)
	}

	return &share, nil
}

// GetSharedTodo resolves a bearer token to its todo. Public read: no userID
// scoping — possession of the unguessable token IS the authorization. Revoked
// tokens read as not-found so revocation is indistinguishable from a bad
// token (no revocation oracle).
func (r *TodoRepository) GetSharedTodo(ctx context.Context, token string) (*todo.PopulatedTodo, error) {
	stmt := `
	SELECT
		t.*,
		CASE
			WHEN c.id IS NOT NULL THEN to_jsonb(camel(c))
			ELSE NULL
		END AS category,
		COALESCE(
			(
				SELECT jsonb_agg(
					to_jsonb(camel(child))
					ORDER BY child.sort_order ASC, child.created_at ASC
				)
				FROM todos child
				WHERE child.parent_todo_id = t.id AND child.user_id = t.user_id
			),
			'[]'::jsonb
		) AS children,
		COALESCE(
			(
				SELECT jsonb_agg(
					to_jsonb(camel(com))
					ORDER BY com.created_at ASC
				)
				FROM todo_comments com
				WHERE com.todo_id = t.id AND com.user_id = t.user_id
			),
			'[]'::jsonb
		) AS comments,
		COALESCE(
			(
				SELECT jsonb_agg(
					to_jsonb(camel(att))
					ORDER BY att.created_at DESC
				)
				FROM todo_attachments att
				WHERE att.todo_id = t.id
			),
			'[]'::jsonb
		) AS attachments
	FROM
		todos t
		JOIN todo_share_tokens s ON s.todo_id = t.id
		LEFT JOIN todo_categories c ON c.id = t.category_id AND c.user_id = t.user_id
	WHERE
		s.token = @token
		AND s.revoked_at IS NULL
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"token": token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get shared todo query: %w", err)
	}

	shared, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.PopulatedTodo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			code := "SHARE_NOT_FOUND"
			return nil, errs.NewNotFoundError("shared todo not found", false, &code)
		}
		return nil, fmt.Errorf("failed to collect shared todo row: %w", err)
	}

	return &shared, nil
}

// RevokeShareTokensForTodo deletes every token for todoID, but only when the
// todo belongs to userID — the JOIN against todos enforces object-level
// authz so one user cannot unshare another user's todo.
func (r *TodoRepository) RevokeShareTokensForTodo(ctx context.Context, userID string, todoID uuid.UUID) error {
	stmt := `
		DELETE FROM todo_share_tokens s
		USING todos t
		WHERE
			s.todo_id = t.id
			AND s.todo_id = @todo_id
			AND t.user_id = @user_id
	`

	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"todo_id": todoID,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke share tokens for todo_id=%s: %w", todoID.String(), err)
	}

	if result.RowsAffected() == 0 {
		code := "SHARE_NOT_FOUND"
		return errs.NewNotFoundError("no active share found for todo", false, &code)
	}

	return nil
}
