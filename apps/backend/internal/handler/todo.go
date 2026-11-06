package handler

import (
	"net/http"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/faizan1250/taskify/internal/validation"
	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	Handler
	todoService TodoServiceContract
}

func NewTodoHandler(s *server.Server, todoService TodoServiceContract) *TodoHandler {
	return &TodoHandler{
		Handler:     NewHandler(s),
		todoService: todoService,
	}
}

func (h *TodoHandler) CreateTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.CreateTodo(c, userID, payload)
		},
		http.StatusCreated,
		&todo.CreateTodoPayload{},
	)(c)
}

func (h *TodoHandler) GetTodoByID(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.GetTodoByIDPayload) (*todo.PopulatedTodo, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.GetTodoByID(c, userID, payload.ID)
		},
		http.StatusOK,
		&todo.GetTodoByIDPayload{},
	)(c)
}

func (h *TodoHandler) GetTodos(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
			userID := middleware.GetUserID(c)
			return h.todoService.GetTodos(c, userID, query)
		},
		http.StatusOK,
		&todo.GetTodosQuery{},
	)(c)
}

func (h *TodoHandler) UpdateTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.UpdateTodo(c, userID, payload)
		},
		http.StatusOK,
		&todo.UpdateTodoPayload{},
	)(c)
}

func (h *TodoHandler) DeleteTodo(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *todo.DeleteTodoPayload) error {
			userID := middleware.GetUserID(c)
			return h.todoService.DeleteTodo(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&todo.DeleteTodoPayload{},
	)(c)
}

// BulkUpdateTodos applies one patch to many todos. It returns 200 when every
// ID succeeds and 207 Multi-Status with a per-ID failed report otherwise.
// These handlers bind manually instead of Handle() because the success
// status is dynamic; validation still flows through BindAndValidate, so
// malformed payloads return 400 identically.
func (h *TodoHandler) BulkUpdateTodos(c echo.Context) error {
	payload := &todo.BulkUpdateTodosPayload{}
	if err := validation.BindAndValidate(c, payload); err != nil {
		return err
	}
	userID := middleware.GetUserID(c)
	result, err := h.todoService.BulkUpdateTodos(c, userID, payload)
	if err != nil {
		return err
	}
	status := http.StatusOK
	if result.HasFailures() {
		status = http.StatusMultiStatus
	}
	return c.JSON(status, result)
}

// BulkDeleteTodos deletes many todos. Like BulkUpdateTodos it returns 200
// on full success and 207 Multi-Status on partial failure. Echo binds JSON
// bodies on DELETE, so DELETE /todos/bulk with {ids[]} needs no
// POST workaround.
func (h *TodoHandler) BulkDeleteTodos(c echo.Context) error {
	payload := &todo.BulkDeleteTodosPayload{}
	if err := validation.BindAndValidate(c, payload); err != nil {
		return err
	}
	userID := middleware.GetUserID(c)
	result, err := h.todoService.BulkDeleteTodos(c, userID, payload)
	if err != nil {
		return err
	}
	status := http.StatusOK
	if result.HasFailures() {
		status = http.StatusMultiStatus
	}
	return c.JSON(status, result)
}

// ReorderTodos persists a drag order. Returns 200 with the reordered todos
// in the caller-supplied order. Binds manually instead of Handle() to mirror
// the bulk handlers; validation still flows through BindAndValidate, so
// malformed payloads return 400 identically.
func (h *TodoHandler) ReorderTodos(c echo.Context) error {
	payload := &todo.ReorderTodosPayload{}
	if err := validation.BindAndValidate(c, payload); err != nil {
		return err
	}
	userID := middleware.GetUserID(c)
	result, err := h.todoService.ReorderTodos(c, userID, payload)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *TodoHandler) GetTodoStats(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.GetTodoStatsPayload) (*todo.TodoStats, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.GetTodoStats(c, userID)
		},
		http.StatusOK,
		&todo.GetTodoStatsPayload{},
	)(c)
}

func (h *TodoHandler) UploadTodoAttachment(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.UploadTodoAttachmentPayload) (*todo.TodoAttachment, error) {
			userID := middleware.GetUserID(c)

			form, err := c.MultipartForm()
			if err != nil {
				return nil, errs.NewBadRequestError("multipart form not found", false, nil, nil, nil)
			}

			files := form.File["file"]
			if len(files) == 0 {
				return nil, errs.NewBadRequestError("no file found", false, nil, nil, nil)
			}

			if len(files) > 1 {
				return nil, errs.NewBadRequestError("only one file allowed per upload", false, nil, nil, nil)
			}

			return h.todoService.UploadTodoAttachment(c, userID, payload.TodoID, files[0])
		},
		http.StatusCreated,
		&todo.UploadTodoAttachmentPayload{},
	)(c)
}

func (h *TodoHandler) DeleteTodoAttachment(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *todo.DeleteTodoAttachmentPayload) error {
			userID := middleware.GetUserID(c)
			return h.todoService.DeleteTodoAttachment(c, userID, payload.TodoID, payload.AttachmentID)
		},
		http.StatusNoContent,
		&todo.DeleteTodoAttachmentPayload{},
	)(c)
}

func (h *TodoHandler) GetAttachmentPresignedURL(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.GetAttachmentPresignedURLPayload) (*struct {
			URL string `json:"url"`
		}, error,
		) {
			userID := middleware.GetUserID(c)
			url, err := h.todoService.GetAttachmentPresignedURL(c, userID, payload.TodoID, payload.AttachmentID)
			if err != nil {
				return nil, err
			}
			return &struct {
				URL string `json:"url"`
			}{URL: url}, nil
		},
		http.StatusOK,
		&todo.GetAttachmentPresignedURLPayload{},
	)(c)
}
