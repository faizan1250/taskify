package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/model/comment"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// errFake is the default sentinel for unconfigured fake calls.
var errFake = errors.New("fake: not implemented")

// newEchoCtx builds a bare echo.Context. service.GetLogger falls back to
// zerolog.Nop when no logger is set, so no middleware wiring is needed.
func newEchoCtx() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

// ---------------------------------------------------------------------------
// FakeTodoRepository — hand fake for repository.TodoRepositoryInterface
// ---------------------------------------------------------------------------

type FakeTodoRepository struct {
	CreateTodoFn                          func(ctx context.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error)
	GetTodoByIDFn                         func(ctx context.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error)
	CheckTodoExistsFn                     func(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error)
	GetTodosFn                            func(ctx context.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error)
	UpdateTodoFn                          func(ctx context.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error)
	DeleteTodoFn                          func(ctx context.Context, userID string, todoID uuid.UUID) error
	BulkUpdateTodosFn                     func(ctx context.Context, userID string, ids []uuid.UUID, patch *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error)
	BulkDeleteTodosFn                     func(ctx context.Context, userID string, ids []uuid.UUID) (*todo.BulkDeleteTodosResult, error)
	GetTodoStatsFn                        func(ctx context.Context, userID string) (*todo.TodoStats, error)
	GetTodoAttachmentFn                   func(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) (*todo.TodoAttachment, error)
	GetTodoAttachmentsFn                  func(ctx context.Context, todoID uuid.UUID) ([]todo.TodoAttachment, error)
	DeleteTodoAttachmentFn                func(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) error
	GetAttachmentKeysForTodoAndSubtasksFn func(ctx context.Context, userID string, todoID uuid.UUID) ([]string, error)
	UploadTodoAttachmentFn                func(ctx context.Context, todoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*todo.TodoAttachment, error)
	GetDueRecurringTodosFn                func(ctx context.Context, now time.Time, limit int) ([]todo.Todo, error)
	MaterializeRecurringTodosFn           func(ctx context.Context, now time.Time, limit int) (int, error)
}

var _ repository.TodoRepositoryInterface = (*FakeTodoRepository)(nil)

func (f *FakeTodoRepository) CreateTodo(ctx context.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	if f.CreateTodoFn == nil {
		return nil, errFake
	}
	return f.CreateTodoFn(ctx, userID, payload)
}

func (f *FakeTodoRepository) GetTodoByID(ctx context.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
	if f.GetTodoByIDFn == nil {
		return nil, errFake
	}
	return f.GetTodoByIDFn(ctx, userID, todoID)
}

func (f *FakeTodoRepository) CheckTodoExists(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Todo, error) {
	if f.CheckTodoExistsFn == nil {
		return nil, errFake
	}
	return f.CheckTodoExistsFn(ctx, userID, todoID)
}

func (f *FakeTodoRepository) GetTodos(ctx context.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
	if f.GetTodosFn == nil {
		return nil, errFake
	}
	return f.GetTodosFn(ctx, userID, query)
}

func (f *FakeTodoRepository) UpdateTodo(ctx context.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	if f.UpdateTodoFn == nil {
		return nil, errFake
	}
	return f.UpdateTodoFn(ctx, userID, payload)
}

func (f *FakeTodoRepository) DeleteTodo(ctx context.Context, userID string, todoID uuid.UUID) error {
	if f.DeleteTodoFn == nil {
		return errFake
	}
	return f.DeleteTodoFn(ctx, userID, todoID)
}

func (f *FakeTodoRepository) BulkUpdateTodos(ctx context.Context, userID string, ids []uuid.UUID, patch *todo.BulkTodoPatch) (*todo.BulkUpdateTodosResult, error) {
	if f.BulkUpdateTodosFn == nil {
		return nil, errFake
	}
	return f.BulkUpdateTodosFn(ctx, userID, ids, patch)
}

func (f *FakeTodoRepository) BulkDeleteTodos(ctx context.Context, userID string, ids []uuid.UUID) (*todo.BulkDeleteTodosResult, error) {
	if f.BulkDeleteTodosFn == nil {
		return nil, errFake
	}
	return f.BulkDeleteTodosFn(ctx, userID, ids)
}

func (f *FakeTodoRepository) GetTodoStats(ctx context.Context, userID string) (*todo.TodoStats, error) {
	if f.GetTodoStatsFn == nil {
		return nil, errFake
	}
	return f.GetTodoStatsFn(ctx, userID)
}

func (f *FakeTodoRepository) GetTodoAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) (*todo.TodoAttachment, error) {
	if f.GetTodoAttachmentFn == nil {
		return nil, errFake
	}
	return f.GetTodoAttachmentFn(ctx, todoID, attachmentID)
}

func (f *FakeTodoRepository) GetTodoAttachments(ctx context.Context, todoID uuid.UUID) ([]todo.TodoAttachment, error) {
	if f.GetTodoAttachmentsFn == nil {
		return nil, errFake
	}
	return f.GetTodoAttachmentsFn(ctx, todoID)
}

func (f *FakeTodoRepository) DeleteTodoAttachment(ctx context.Context, todoID uuid.UUID, attachmentID uuid.UUID) error {
	if f.DeleteTodoAttachmentFn == nil {
		return errFake
	}
	return f.DeleteTodoAttachmentFn(ctx, todoID, attachmentID)
}

func (f *FakeTodoRepository) GetAttachmentKeysForTodoAndSubtasks(ctx context.Context, userID string, todoID uuid.UUID) ([]string, error) {
	if f.GetAttachmentKeysForTodoAndSubtasksFn == nil {
		return nil, errFake
	}
	return f.GetAttachmentKeysForTodoAndSubtasksFn(ctx, userID, todoID)
}

func (f *FakeTodoRepository) UploadTodoAttachment(ctx context.Context, todoID uuid.UUID, userID string, s3Key string, fileName string, fileSize int64, mimeType string) (*todo.TodoAttachment, error) {
	if f.UploadTodoAttachmentFn == nil {
		return nil, errFake
	}
	return f.UploadTodoAttachmentFn(ctx, todoID, userID, s3Key, fileName, fileSize, mimeType)
}

func (f *FakeTodoRepository) GetTodosDueInHours(ctx context.Context, hours int, limit int) ([]todo.Todo, error) {
	return nil, errFake
}

func (f *FakeTodoRepository) GetOverdueTodos(ctx context.Context, limit int) ([]todo.Todo, error) {
	return nil, errFake
}

func (f *FakeTodoRepository) GetCompletedTodosOlderThan(ctx context.Context, cutoffDate time.Time, limit int) ([]todo.Todo, error) {
	return nil, errFake
}

func (f *FakeTodoRepository) ArchiveTodos(ctx context.Context, todoIDs []uuid.UUID) error {
	return errFake
}

func (f *FakeTodoRepository) GetWeeklyStatsForUsers(ctx context.Context, startDate, endDate time.Time) ([]todo.UserWeeklyStats, error) {
	return nil, errFake
}

func (f *FakeTodoRepository) GetCompletedTodosForUser(ctx context.Context, userID string, startDate, endDate time.Time) ([]todo.PopulatedTodo, error) {
	return nil, errFake
}

func (f *FakeTodoRepository) GetOverdueTodosForUser(ctx context.Context, userID string) ([]todo.PopulatedTodo, error) {
	return nil, errFake
}

func (f *FakeTodoRepository) MarkDueDateReminderSent(ctx context.Context, todoID uuid.UUID) error {
	return errFake
}

func (f *FakeTodoRepository) MarkOverdueNotificationSent(ctx context.Context, todoID uuid.UUID) error {
	return errFake
}

func (f *FakeTodoRepository) GetTodoStatus(ctx context.Context, id uuid.UUID) (string, error) {
	return "", errFake
}

func (f *FakeTodoRepository) GetDueRecurringTodos(ctx context.Context, now time.Time, limit int) ([]todo.Todo, error) {
	if f.GetDueRecurringTodosFn == nil {
		return nil, errFake
	}
	return f.GetDueRecurringTodosFn(ctx, now, limit)
}

func (f *FakeTodoRepository) MaterializeRecurringTodos(ctx context.Context, now time.Time, limit int) (int, error) {
	if f.MaterializeRecurringTodosFn == nil {
		return 0, errFake
	}
	return f.MaterializeRecurringTodosFn(ctx, now, limit)
}

// ---------------------------------------------------------------------------
// FakeCategoryRepository — hand fake for repository.CategoryRepositoryInterface
// ---------------------------------------------------------------------------

type FakeCategoryRepository struct {
	CreateCategoryFn  func(ctx context.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error)
	GetCategoryByIDFn func(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error)
	GetCategoriesFn   func(ctx context.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error)
	UpdateCategoryFn  func(ctx context.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error)
	DeleteCategoryFn  func(ctx context.Context, userID string, categoryID uuid.UUID) error
}

var _ repository.CategoryRepositoryInterface = (*FakeCategoryRepository)(nil)

func (f *FakeCategoryRepository) CreateCategory(ctx context.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
	if f.CreateCategoryFn == nil {
		return nil, errFake
	}
	return f.CreateCategoryFn(ctx, userID, payload)
}

func (f *FakeCategoryRepository) GetCategoryByID(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error) {
	if f.GetCategoryByIDFn == nil {
		return nil, errFake
	}
	return f.GetCategoryByIDFn(ctx, userID, categoryID)
}

func (f *FakeCategoryRepository) GetCategories(ctx context.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
	if f.GetCategoriesFn == nil {
		return nil, errFake
	}
	return f.GetCategoriesFn(ctx, userID, query)
}

func (f *FakeCategoryRepository) UpdateCategory(ctx context.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error) {
	if f.UpdateCategoryFn == nil {
		return nil, errFake
	}
	return f.UpdateCategoryFn(ctx, userID, categoryID, payload)
}

func (f *FakeCategoryRepository) DeleteCategory(ctx context.Context, userID string, categoryID uuid.UUID) error {
	if f.DeleteCategoryFn == nil {
		return errFake
	}
	return f.DeleteCategoryFn(ctx, userID, categoryID)
}

// ---------------------------------------------------------------------------
// FakeCommentRepository — hand fake for repository.CommentRepositoryInterface
// ---------------------------------------------------------------------------

type FakeCommentRepository struct {
	AddCommentFn          func(ctx context.Context, userID string, todoID uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error)
	GetCommentsByTodoIDFn func(ctx context.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error)
	GetCommentByIDFn      func(ctx context.Context, userID string, commentID uuid.UUID) (*comment.Comment, error)
	UpdateCommentFn       func(ctx context.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error)
	DeleteCommentFn       func(ctx context.Context, userID string, commentID uuid.UUID) error
}

var _ repository.CommentRepositoryInterface = (*FakeCommentRepository)(nil)

func (f *FakeCommentRepository) AddComment(ctx context.Context, userID string, todoID uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error) {
	if f.AddCommentFn == nil {
		return nil, errFake
	}
	return f.AddCommentFn(ctx, userID, todoID, payload)
}

func (f *FakeCommentRepository) GetCommentsByTodoID(ctx context.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error) {
	if f.GetCommentsByTodoIDFn == nil {
		return nil, errFake
	}
	return f.GetCommentsByTodoIDFn(ctx, userID, todoID)
}

func (f *FakeCommentRepository) GetCommentByID(ctx context.Context, userID string, commentID uuid.UUID) (*comment.Comment, error) {
	if f.GetCommentByIDFn == nil {
		return nil, errFake
	}
	return f.GetCommentByIDFn(ctx, userID, commentID)
}

func (f *FakeCommentRepository) UpdateComment(ctx context.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error) {
	if f.UpdateCommentFn == nil {
		return nil, errFake
	}
	return f.UpdateCommentFn(ctx, userID, commentID, content)
}

func (f *FakeCommentRepository) DeleteComment(ctx context.Context, userID string, commentID uuid.UUID) error {
	if f.DeleteCommentFn == nil {
		return errFake
	}
	return f.DeleteCommentFn(ctx, userID, commentID)
}
