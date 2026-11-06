package handler_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/handler"
	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/model/comment"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserID = "user-123"

// errFakeService is returned when a fake method has no Fn configured.
var errFakeService = errors.New("fake service: not implemented")

// newHandlerContext builds an echo.Context backed by httptest, wiring the
// user_id value that middleware.GetUserID reads. Logger falls back to
// zerolog.Nop via middleware.GetLogger, so no middleware chain is needed.
// Direct handler calls bypass GlobalErrorHandler, so tests assert on the
// returned error via errors.As into *errs.HTTPError.
func newHandlerContext(method, target string, body io.Reader, contentType string, params map[string]string, userID string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, body)
	if body != nil {
		ct := contentType
		if ct == "" {
			ct = echo.MIMEApplicationJSON
		}
		req.Header.Set(echo.HeaderContentType, ct)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if len(params) > 0 {
		names := make([]string, 0, len(params))
		for k := range params {
			names = append(names, k)
		}
		sort.Strings(names)
		values := make([]string, 0, len(names))
		for _, k := range names {
			values = append(values, params[k])
		}
		c.SetParamNames(names...)
		c.SetParamValues(values...)
	}
	c.Set(middleware.UserIDKey, userID)
	return c, rec
}

// newMultipartRequest builds a multipart POST request with files stored under
// a single form field.
func newMultipartRequest(t *testing.T, target, fieldName string, files map[string]string) *http.Request {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fw, err := w.CreateFormFile(fieldName, name)
		require.NoError(t, err)
		_, err = io.WriteString(fw, files[name])
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	req := httptest.NewRequest(http.MethodPost, target, &b)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	return req
}

// requireHTTPError asserts err carries an *errs.HTTPError with the expected status.
func requireHTTPError(t *testing.T, err error, status int) *errs.HTTPError {
	t.Helper()
	require.Error(t, err)
	var httpErr *errs.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, status, httpErr.Status)
	return httpErr
}

// ---------------------------------------------------------------------------
// FakeTodoService — hand fake for handler.TodoServiceContract
// ---------------------------------------------------------------------------

type FakeTodoService struct {
	CreateTodoFn                func(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error)
	GetTodoByIDFn               func(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error)
	GetTodosFn                  func(ctx echo.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error)
	UpdateTodoFn                func(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error)
	DeleteTodoFn                func(ctx echo.Context, userID string, todoID uuid.UUID) error
	BulkUpdateTodosFn           func(ctx echo.Context, userID string, payload *todo.BulkUpdateTodosPayload) (*todo.BulkUpdateTodosResult, error)
	BulkDeleteTodosFn           func(ctx echo.Context, userID string, payload *todo.BulkDeleteTodosPayload) (*todo.BulkDeleteTodosResult, error)
	ReorderTodosFn              func(ctx echo.Context, userID string, payload *todo.ReorderTodosPayload) ([]todo.Todo, error)
	GetTodoStatsFn              func(ctx echo.Context, userID string, payload *todo.GetTodoStatsPayload) (*todo.TodoStats, error)
	UploadTodoAttachmentFn      func(ctx echo.Context, userID string, todoID uuid.UUID, file *multipart.FileHeader) (*todo.TodoAttachment, error)
	DeleteTodoAttachmentFn      func(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) error
	GetAttachmentPresignedURLFn func(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) (string, error)
}

var _ handler.TodoServiceContract = (*FakeTodoService)(nil)

func (f *FakeTodoService) CreateTodo(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	if f.CreateTodoFn == nil {
		return nil, errFakeService
	}
	return f.CreateTodoFn(ctx, userID, payload)
}

func (f *FakeTodoService) GetTodoByID(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
	if f.GetTodoByIDFn == nil {
		return nil, errFakeService
	}
	return f.GetTodoByIDFn(ctx, userID, todoID)
}

func (f *FakeTodoService) GetTodos(ctx echo.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
	if f.GetTodosFn == nil {
		return nil, errFakeService
	}
	return f.GetTodosFn(ctx, userID, query)
}

func (f *FakeTodoService) UpdateTodo(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	if f.UpdateTodoFn == nil {
		return nil, errFakeService
	}
	return f.UpdateTodoFn(ctx, userID, payload)
}

func (f *FakeTodoService) DeleteTodo(ctx echo.Context, userID string, todoID uuid.UUID) error {
	if f.DeleteTodoFn == nil {
		return errFakeService
	}
	return f.DeleteTodoFn(ctx, userID, todoID)
}

func (f *FakeTodoService) BulkUpdateTodos(ctx echo.Context, userID string, payload *todo.BulkUpdateTodosPayload) (*todo.BulkUpdateTodosResult, error) {
	if f.BulkUpdateTodosFn == nil {
		return nil, errFakeService
	}
	return f.BulkUpdateTodosFn(ctx, userID, payload)
}

func (f *FakeTodoService) BulkDeleteTodos(ctx echo.Context, userID string, payload *todo.BulkDeleteTodosPayload) (*todo.BulkDeleteTodosResult, error) {
	if f.BulkDeleteTodosFn == nil {
		return nil, errFakeService
	}
	return f.BulkDeleteTodosFn(ctx, userID, payload)
}

func (f *FakeTodoService) ReorderTodos(ctx echo.Context, userID string, payload *todo.ReorderTodosPayload) ([]todo.Todo, error) {
	if f.ReorderTodosFn == nil {
		return nil, errFakeService
	}
	return f.ReorderTodosFn(ctx, userID, payload)
}

func (f *FakeTodoService) GetTodoStats(ctx echo.Context, userID string, payload *todo.GetTodoStatsPayload) (*todo.TodoStats, error) {
	if f.GetTodoStatsFn == nil {
		return nil, errFakeService
	}
	return f.GetTodoStatsFn(ctx, userID, payload)
}

func (f *FakeTodoService) UploadTodoAttachment(ctx echo.Context, userID string, todoID uuid.UUID, file *multipart.FileHeader) (*todo.TodoAttachment, error) {
	if f.UploadTodoAttachmentFn == nil {
		return nil, errFakeService
	}
	return f.UploadTodoAttachmentFn(ctx, userID, todoID, file)
}

func (f *FakeTodoService) DeleteTodoAttachment(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) error {
	if f.DeleteTodoAttachmentFn == nil {
		return errFakeService
	}
	return f.DeleteTodoAttachmentFn(ctx, userID, todoID, attachmentID)
}

func (f *FakeTodoService) GetAttachmentPresignedURL(ctx echo.Context, userID string, todoID uuid.UUID, attachmentID uuid.UUID) (string, error) {
	if f.GetAttachmentPresignedURLFn == nil {
		return "", errFakeService
	}
	return f.GetAttachmentPresignedURLFn(ctx, userID, todoID, attachmentID)
}

// ---------------------------------------------------------------------------
// FakeCategoryService — hand fake for handler.CategoryServiceContract
// ---------------------------------------------------------------------------

type FakeCategoryService struct {
	CreateCategoryFn func(ctx echo.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error)
	GetCategoriesFn  func(ctx echo.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error)
	UpdateCategoryFn func(ctx echo.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error)
	DeleteCategoryFn func(ctx echo.Context, userID string, categoryID uuid.UUID) error
}

var _ handler.CategoryServiceContract = (*FakeCategoryService)(nil)

func (f *FakeCategoryService) CreateCategory(ctx echo.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
	if f.CreateCategoryFn == nil {
		return nil, errFakeService
	}
	return f.CreateCategoryFn(ctx, userID, payload)
}

func (f *FakeCategoryService) GetCategories(ctx echo.Context, userID string, query *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
	if f.GetCategoriesFn == nil {
		return nil, errFakeService
	}
	return f.GetCategoriesFn(ctx, userID, query)
}

func (f *FakeCategoryService) UpdateCategory(ctx echo.Context, userID string, categoryID uuid.UUID, payload *category.UpdateCategoryPayload) (*category.Category, error) {
	if f.UpdateCategoryFn == nil {
		return nil, errFakeService
	}
	return f.UpdateCategoryFn(ctx, userID, categoryID, payload)
}

func (f *FakeCategoryService) DeleteCategory(ctx echo.Context, userID string, categoryID uuid.UUID) error {
	if f.DeleteCategoryFn == nil {
		return errFakeService
	}
	return f.DeleteCategoryFn(ctx, userID, categoryID)
}

// ---------------------------------------------------------------------------
// FakeCommentService — hand fake for handler.CommentServiceContract
// ---------------------------------------------------------------------------

type FakeCommentService struct {
	AddCommentFn          func(ctx echo.Context, userID string, todoID uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error)
	GetCommentsByTodoIDFn func(ctx echo.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error)
	UpdateCommentFn       func(ctx echo.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error)
	DeleteCommentFn       func(ctx echo.Context, userID string, commentID uuid.UUID) error
}

var _ handler.CommentServiceContract = (*FakeCommentService)(nil)

func (f *FakeCommentService) AddComment(ctx echo.Context, userID string, todoID uuid.UUID, payload *comment.AddCommentPayload) (*comment.Comment, error) {
	if f.AddCommentFn == nil {
		return nil, errFakeService
	}
	return f.AddCommentFn(ctx, userID, todoID, payload)
}

func (f *FakeCommentService) GetCommentsByTodoID(ctx echo.Context, userID string, todoID uuid.UUID) ([]comment.Comment, error) {
	if f.GetCommentsByTodoIDFn == nil {
		return nil, errFakeService
	}
	return f.GetCommentsByTodoIDFn(ctx, userID, todoID)
}

func (f *FakeCommentService) UpdateComment(ctx echo.Context, userID string, commentID uuid.UUID, content string) (*comment.Comment, error) {
	if f.UpdateCommentFn == nil {
		return nil, errFakeService
	}
	return f.UpdateCommentFn(ctx, userID, commentID, content)
}

func (f *FakeCommentService) DeleteComment(ctx echo.Context, userID string, commentID uuid.UUID) error {
	if f.DeleteCommentFn == nil {
		return errFakeService
	}
	return f.DeleteCommentFn(ctx, userID, commentID)
}
