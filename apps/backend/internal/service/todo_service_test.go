package service_test

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errRepoBoom = errors.New("repo boom")

// validFileHeader returns a header that passes ValidateAttachment. The
// attachment precondition tests fail before file.Open, so no body needed.
func validFileHeader(_ *testing.T) *multipart.FileHeader {
	return &multipart.FileHeader{Filename: "note.txt", Size: 1024}
}

func TestTodoService_CreateTodo(t *testing.T) {
	parentID := uuid.New()
	categoryID := uuid.New()
	subtaskParentID := uuid.New() // parent that is itself a subtask

	tests := []struct {
		name      string
		payload   *todo.CreateTodoPayload
		setup     func(tr *FakeTodoRepository, cr *FakeCategoryRepository)
		wantErr   string // "" = happy path
		wantTitle string
	}{
		{
			name:      "happy path without parent or category",
			payload:   &todo.CreateTodoPayload{Title: "buy milk"},
			setup:     func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {},
			wantTitle: "buy milk",
		},
		{
			name:    "happy path with valid parent and category",
			payload: &todo.CreateTodoPayload{Title: "sub", ParentTodoID: &parentID, CategoryID: &categoryID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return &todo.Todo{}, nil
				}
			},
			wantTitle: "sub",
		},
		{
			name:    "parent validation error propagates",
			payload: &todo.CreateTodoPayload{Title: "sub", ParentTodoID: &parentID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
		{
			name:    "parent that is a subtask is rejected",
			payload: &todo.CreateTodoPayload{Title: "sub", ParentTodoID: &subtaskParentID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return &todo.Todo{ParentTodoID: &parentID}, nil
				}
			},
			wantErr: "Parent todo cannot have children",
		},
		{
			name:    "category validation error propagates",
			payload: &todo.CreateTodoPayload{Title: "t", CategoryID: &categoryID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				cr.GetCategoryByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*category.Category, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
		{
			name:    "create error propagates",
			payload: &todo.CreateTodoPayload{Title: "t"},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				tr.CreateTodoFn = func(_ context.Context, _ string, _ *todo.CreateTodoPayload) (*todo.Todo, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &FakeTodoRepository{}
			catRepo := &FakeCategoryRepository{}
			tt.setup(todoRepo, catRepo)

			// Defaults: parent check passes, category check passes, create echoes title.
			if todoRepo.CheckTodoExistsFn == nil {
				todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return &todo.Todo{}, nil
				}
			}
			if catRepo.GetCategoryByIDFn == nil {
				catRepo.GetCategoryByIDFn = func(_ context.Context, _ string, id uuid.UUID) (*category.Category, error) {
					return &category.Category{}, nil
				}
			}
			if todoRepo.CreateTodoFn == nil {
				title := tt.payload.Title
				todoRepo.CreateTodoFn = func(_ context.Context, _ string, _ *todo.CreateTodoPayload) (*todo.Todo, error) {
					return &todo.Todo{Title: title}, nil
				}
			}

			svc := service.NewTodoService(nil, todoRepo, catRepo, nil)
			got, err := svc.CreateTodo(newEchoCtx(), "user-1", tt.payload)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantTitle, got.Title)
		})
	}
}

func TestTodoService_GetTodoByID(t *testing.T) {
	todoID := uuid.New()
	tests := []struct {
		name    string
		setup   func(tr *FakeTodoRepository)
		wantErr string
	}{
		{
			name: "happy path",
			setup: func(tr *FakeTodoRepository) {
				tr.GetTodoByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.PopulatedTodo, error) {
					return &todo.PopulatedTodo{}, nil
				}
			},
		},
		{
			name: "repo error propagates",
			setup: func(tr *FakeTodoRepository) {
				tr.GetTodoByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.PopulatedTodo, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &FakeTodoRepository{}
			tt.setup(todoRepo)
			svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
			got, err := svc.GetTodoByID(newEchoCtx(), "user-1", todoID)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestTodoService_GetTodos(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(tr *FakeTodoRepository)
		wantErr   string
		wantTotal int
	}{
		{
			name: "happy path",
			setup: func(tr *FakeTodoRepository) {
				tr.GetTodosFn = func(_ context.Context, _ string, _ *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
					return &model.PaginatedResponse[todo.PopulatedTodo]{Data: []todo.PopulatedTodo{{}, {}}, Total: 2}, nil
				}
			},
			wantTotal: 2,
		},
		{
			name: "repo error propagates",
			setup: func(tr *FakeTodoRepository) {
				tr.GetTodosFn = func(_ context.Context, _ string, _ *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &FakeTodoRepository{}
			tt.setup(todoRepo)
			svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
			got, err := svc.GetTodos(newEchoCtx(), "user-1", &todo.GetTodosQuery{})
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantTotal, got.Total)
		})
	}
}

func TestTodoService_UpdateTodo(t *testing.T) {
	parentID := uuid.New()
	categoryID := uuid.New()
	todoID := uuid.New()
	newTitle := "updated"

	tests := []struct {
		name    string
		payload *todo.UpdateTodoPayload
		setup   func(tr *FakeTodoRepository, cr *FakeCategoryRepository)
		wantErr string
	}{
		{
			name:    "happy path without parent or category",
			payload: &todo.UpdateTodoPayload{ID: todoID, Title: &newTitle},
			setup:   func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {},
		},
		{
			name:    "happy path with valid parent and category",
			payload: &todo.UpdateTodoPayload{ID: todoID, ParentTodoID: &parentID, CategoryID: &categoryID},
			setup:   func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {},
		},
		{
			name:    "parent check error propagates",
			payload: &todo.UpdateTodoPayload{ID: todoID, ParentTodoID: &parentID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
		{
			name:    "self parent is rejected",
			payload: &todo.UpdateTodoPayload{ID: todoID, ParentTodoID: &todoID},
			setup:   func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {},
			wantErr: "Todo cannot be its own parent",
		},
		{
			name:    "parent that is a subtask is rejected",
			payload: &todo.UpdateTodoPayload{ID: todoID, ParentTodoID: &parentID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return &todo.Todo{ParentTodoID: &todoID}, nil
				}
			},
			wantErr: "Parent todo cannot have children",
		},
		{
			name:    "category check error propagates",
			payload: &todo.UpdateTodoPayload{ID: todoID, CategoryID: &categoryID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				cr.GetCategoryByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*category.Category, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
		{
			name:    "update error propagates",
			payload: &todo.UpdateTodoPayload{ID: todoID},
			setup: func(tr *FakeTodoRepository, cr *FakeCategoryRepository) {
				tr.UpdateTodoFn = func(_ context.Context, _ string, _ *todo.UpdateTodoPayload) (*todo.Todo, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &FakeTodoRepository{}
			catRepo := &FakeCategoryRepository{}
			tt.setup(todoRepo, catRepo)

			if todoRepo.CheckTodoExistsFn == nil {
				todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, id uuid.UUID) (*todo.Todo, error) {
					// Default parent is a top-level todo with a different ID.
					other := uuid.New()
					for other == tt.payload.ID {
						other = uuid.New()
					}
					parent := &todo.Todo{}
					parent.ID = other
					return parent, nil
				}
			}
			if catRepo.GetCategoryByIDFn == nil {
				catRepo.GetCategoryByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*category.Category, error) {
					return &category.Category{}, nil
				}
			}
			if todoRepo.UpdateTodoFn == nil {
				todoRepo.UpdateTodoFn = func(_ context.Context, _ string, _ *todo.UpdateTodoPayload) (*todo.Todo, error) {
					return &todo.Todo{Title: "updated"}, nil
				}
			}

			// Self-parent case: parent lookup must return the todo itself.
			if tt.name == "self parent is rejected" {
				todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					self := &todo.Todo{}
					self.ID = todoID
					return self, nil
				}
			}

			svc := service.NewTodoService(nil, todoRepo, catRepo, nil)
			got, err := svc.UpdateTodo(newEchoCtx(), "user-1", tt.payload)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestTodoService_DeleteTodo(t *testing.T) {
	todoID := uuid.New()
	tests := []struct {
		name    string
		setup   func(tr *FakeTodoRepository)
		wantErr string
	}{
		{
			name: "happy path with no attachments",
			setup: func(tr *FakeTodoRepository) {
				tr.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
					return nil, nil
				}
				tr.DeleteTodoFn = func(_ context.Context, _ string, _ uuid.UUID) error { return nil }
			},
		},
		{
			name: "happy path with attachment keys skips S3 when client is nil",
			setup: func(tr *FakeTodoRepository) {
				tr.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
					return []string{"k1", "k2"}, nil
				}
				tr.DeleteTodoFn = func(_ context.Context, _ string, _ uuid.UUID) error { return nil }
			},
		},
		{
			name: "attachment keys error propagates",
			setup: func(tr *FakeTodoRepository) {
				tr.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
		{
			name: "delete error propagates",
			setup: func(tr *FakeTodoRepository) {
				tr.GetAttachmentKeysForTodoAndSubtasksFn = func(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
					return nil, nil
				}
				tr.DeleteTodoFn = func(_ context.Context, _ string, _ uuid.UUID) error { return errRepoBoom }
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &FakeTodoRepository{}
			tt.setup(todoRepo)
			// awsClient nil: CleanupS3AttachmentsAsync is a documented no-op.
			svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
			err := svc.DeleteTodo(newEchoCtx(), "user-1", todoID)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTodoService_GetTodoStats(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(tr *FakeTodoRepository)
		wantErr   string
		wantTotal int
	}{
		{
			name: "happy path",
			setup: func(tr *FakeTodoRepository) {
				tr.GetTodoStatsFn = func(_ context.Context, _ string) (*todo.TodoStats, error) {
					return &todo.TodoStats{Total: 5, Active: 3, Completed: 2}, nil
				}
			},
			wantTotal: 5,
		},
		{
			name: "repo error propagates",
			setup: func(tr *FakeTodoRepository) {
				tr.GetTodoStatsFn = func(_ context.Context, _ string) (*todo.TodoStats, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoRepo := &FakeTodoRepository{}
			tt.setup(todoRepo)
			svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
			got, err := svc.GetTodoStats(newEchoCtx(), "user-1")
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, got.Total)
		})
	}
}

func TestTodoService_AttachmentPreconditions(t *testing.T) {
	todoID := uuid.New()
	attachmentID := uuid.New()

	t.Run("UploadTodoAttachment propagates todo-check error before touching S3", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.UploadTodoAttachment(newEchoCtx(), "user-1", todoID, validFileHeader(t))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})

	t.Run("DeleteTodoAttachment propagates todo-check error", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		err := svc.DeleteTodoAttachment(newEchoCtx(), "user-1", todoID, attachmentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})

	t.Run("DeleteTodoAttachment propagates get-attachment error", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return &todo.Todo{}, nil
		}
		todoRepo.GetTodoAttachmentFn = func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*todo.TodoAttachment, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		err := svc.DeleteTodoAttachment(newEchoCtx(), "user-1", todoID, attachmentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})

	t.Run("GetAttachmentPresignedURL propagates todo-check error", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.GetAttachmentPresignedURL(newEchoCtx(), "user-1", todoID, attachmentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})

	t.Run("GetAttachmentPresignedURL propagates get-attachment error", func(t *testing.T) {
		todoRepo := &FakeTodoRepository{}
		todoRepo.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
			return &todo.Todo{}, nil
		}
		todoRepo.GetTodoAttachmentFn = func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*todo.TodoAttachment, error) {
			return nil, errRepoBoom
		}
		svc := service.NewTodoService(nil, todoRepo, &FakeCategoryRepository{}, nil)
		_, err := svc.GetAttachmentPresignedURL(newEchoCtx(), "user-1", todoID, attachmentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo boom")
	})
}
