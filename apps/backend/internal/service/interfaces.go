package service

import (
	"context"
	"mime/multipart"

	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/model/comment"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TodoServiceInterface interface {
	CreateTodo(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error)
	GetTodoByID(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error)
	GetTodos(ctx echo.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error)
	UpdateTodo(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error)
	DeleteTodo(ctx echo.Context, userID string, todoID uuid.UUID) error
	ReorderTodos(ctx echo.Context, userID string, payload *todo.ReorderTodosPayload) ([]todo.Todo, error)
	CleanupS3AttachmentsAsync(reqCtx context.Context, keys []string)
	GetTodoStats(ctx echo.Context, userID string) (*todo.TodoStats, error)
	UploadTodoAttachment(ctx echo.Context, userID string, todoID uuid.UUID, file *multipart.FileHeader) (*todo.TodoAttachment, error)
	DeleteTodoAttachment(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) error
	GetAttachmentPresignedURL(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) (string, error)
}

type CategoryServiceInterface interface {
	CreateCategory(ctx echo.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error)
	GetCategories(ctx echo.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error)
	GetCategoryByID(ctx echo.Context, userID string, categoryID uuid.UUID) (*category.Category, error)
	UpdateCategory(ctx echo.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error)
	DeleteCategory(ctx echo.Context, userID string, categoryID uuid.UUID) error
}

type CommentServiceInterface interface {
	AddComment(ctx echo.Context, userID string, todoID uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error)
	GetCommentsByTodoID(ctx echo.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error)
	UpdateComment(ctx echo.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error)
	DeleteComment(ctx echo.Context, userID string, commentID uuid.UUID) error
}

var _ TodoServiceInterface = (*TodoService)(nil)
var _ CategoryServiceInterface = (*CategoryService)(nil)
var _ CommentServiceInterface = (*CommentService)(nil)
