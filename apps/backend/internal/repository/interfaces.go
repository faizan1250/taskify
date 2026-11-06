package repository

import (
	"context"
	"time"

	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/model/comment"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
)

type TodoRepositoryInterface interface {
	CreateTodo(ctx context.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error)
	GetTodoByID(ctx context.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error)
	CheckTodoExists(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error)
	GetTodos(ctx context.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error)
	UpdateTodo(ctx context.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error)
	DeleteTodo(ctx context.Context, userID string, todoID uuid.UUID) error
	BulkUpdateTodos(ctx context.Context, userID string, ids []uuid.UUID, patch *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error)
	BulkDeleteTodos(ctx context.Context, userID string, ids []uuid.UUID) (*todo.BulkDeleteTodosResult, error)
	GetTodoStats(ctx context.Context, userID string) (*todo.TodoStats, error)
	GetTodoAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) (*todo.TodoAttachment, error)
	GetTodoAttachments(ctx context.Context, todoID uuid.UUID) ([]todo.TodoAttachment, error)
	DeleteTodoAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) error
	GetAttachmentKeysForTodoAndSubtasks(ctx context.Context, userID string, todoID uuid.UUID) ([]string, error)
	UploadTodoAttachment(ctx context.Context, todoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*todo.TodoAttachment, error)
	GetTodosDueInHours(ctx context.Context, hours int, limit int) ([]todo.Todo, error)
	GetOverdueTodos(ctx context.Context, limit int) ([]todo.Todo, error)
	GetCompletedTodosOlderThan(ctx context.Context, cutoffDate time.Time, limit int) ([]todo.Todo, error)
	ArchiveTodos(ctx context.Context, todoIDs []uuid.UUID) error
	GetWeeklyStatsForUsers(ctx context.Context, startDate, endDate time.Time) ([]todo.UserWeeklyStats, error)
	GetCompletedTodosForUser(ctx context.Context, userID string, startDate, endDate time.Time) ([]todo.PopulatedTodo, error)
	GetOverdueTodosForUser(ctx context.Context, userID string) ([]todo.PopulatedTodo, error)
	MarkDueDateReminderSent(ctx context.Context, todoID uuid.UUID) error
	MarkOverdueNotificationSent(ctx context.Context, todoID uuid.UUID) error
	GetTodoStatus(ctx context.Context, id uuid.UUID) (string, error)
}

type CategoryRepositoryInterface interface {
	CreateCategory(ctx context.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error)
	GetCategoryByID(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error)
	GetCategories(ctx context.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error)
	UpdateCategory(ctx context.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error)
	DeleteCategory(ctx context.Context, userID string, categoryID uuid.UUID) error
}

type CommentRepositoryInterface interface {
	AddComment(ctx context.Context, userID string, todoID uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error)
	GetCommentsByTodoID(ctx context.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error)
	GetCommentByID(ctx context.Context, userID string, commentID uuid.UUID) (*comment.Comment, error)
	UpdateComment(ctx context.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error)
	DeleteComment(ctx context.Context, userID string, commentID uuid.UUID) error
}

var _ TodoRepositoryInterface = (*TodoRepository)(nil)
var _ CategoryRepositoryInterface = (*CategoryRepository)(nil)
var _ CommentRepositoryInterface = (*CommentRepository)(nil)
