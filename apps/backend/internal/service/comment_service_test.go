package service_test

import (
	"context"
	"testing"

	"github.com/faizan1250/taskify/internal/model/comment"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// passTodoCheck wires CheckTodoExists to succeed with a top-level todo.
func passTodoCheck(tr *FakeTodoRepository) {
	tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
		return &todo.Todo{}, nil
	}
}

func TestCommentService_AddComment(t *testing.T) {
	todoID := uuid.New()
	tests := []struct {
		name       string
		setupTodo  func(tr *FakeTodoRepository)
		setupCmt   func(cr *FakeCommentRepository)
		wantErr    string
		wantConent string
	}{
		{
			name:      "happy path",
			setupTodo: passTodoCheck,
			setupCmt: func(cr *FakeCommentRepository) {
				cr.AddCommentFn = func(_ context.Context, _ string, _ uuid.UUID, _ *comment.AddCommentPayload) (*comment.Comment, error) {
					return &comment.Comment{Content: "nice"}, nil
				}
			},
			wantConent: "nice",
		},
		{
			name: "todo validation error propagates",
			setupTodo: func(tr *FakeTodoRepository) {
				tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return nil, errRepoBoom
				}
			},
			setupCmt: func(cr *FakeCommentRepository) {},
			wantErr:  "repo boom",
		},
		{
			name:      "add error propagates",
			setupTodo: passTodoCheck,
			setupCmt: func(cr *FakeCommentRepository) {
				cr.AddCommentFn = func(_ context.Context, _ string, _ uuid.UUID, _ *comment.AddCommentPayload) (*comment.Comment, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &FakeTodoRepository{}
			cr := &FakeCommentRepository{}
			tt.setupTodo(tr)
			tt.setupCmt(cr)
			svc := service.NewCommentService(nil, cr, tr)
			got, err := svc.AddComment(newEchoCtx(), "user-1", todoID, &comment.AddCommentPayload{Content: "nice"})
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantConent, got.Content)
		})
	}
}

func TestCommentService_GetCommentsByTodoID(t *testing.T) {
	todoID := uuid.New()
	tests := []struct {
		name      string
		setupTodo func(tr *FakeTodoRepository)
		setupCmt  func(cr *FakeCommentRepository)
		wantErr   string
		wantLen   int
	}{
		{
			name:      "happy path",
			setupTodo: passTodoCheck,
			setupCmt: func(cr *FakeCommentRepository) {
				cr.GetCommentsByTodoIDFn = func(_ context.Context, _ string, _ uuid.UUID) ([]comment.Comment, error) {
					return []comment.Comment{{Content: "a"}, {Content: "b"}}, nil
				}
			},
			wantLen: 2,
		},
		{
			name: "todo validation error propagates",
			setupTodo: func(tr *FakeTodoRepository) {
				tr.CheckTodoExistsFn = func(_ context.Context, _ string, _ uuid.UUID) (*todo.Todo, error) {
					return nil, errRepoBoom
				}
			},
			setupCmt: func(cr *FakeCommentRepository) {},
			wantErr:  "repo boom",
		},
		{
			name:      "fetch error propagates",
			setupTodo: passTodoCheck,
			setupCmt: func(cr *FakeCommentRepository) {
				cr.GetCommentsByTodoIDFn = func(_ context.Context, _ string, _ uuid.UUID) ([]comment.Comment, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &FakeTodoRepository{}
			cr := &FakeCommentRepository{}
			tt.setupTodo(tr)
			tt.setupCmt(cr)
			svc := service.NewCommentService(nil, cr, tr)
			got, err := svc.GetCommentsByTodoID(newEchoCtx(), "user-1", todoID)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestCommentService_UpdateComment(t *testing.T) {
	commentID := uuid.New()
	tests := []struct {
		name     string
		setup    func(cr *FakeCommentRepository)
		wantErr  string
		wantBody string
	}{
		{
			name: "happy path",
			setup: func(cr *FakeCommentRepository) {
				cr.GetCommentByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*comment.Comment, error) {
					return &comment.Comment{Content: "old"}, nil
				}
				cr.UpdateCommentFn = func(_ context.Context, _ string, _ uuid.UUID, content string) (*comment.Comment, error) {
					return &comment.Comment{Content: content}, nil
				}
			},
			wantBody: "new body",
		},
		{
			name: "validation lookup error propagates",
			setup: func(cr *FakeCommentRepository) {
				cr.GetCommentByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*comment.Comment, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
		{
			name: "update error propagates",
			setup: func(cr *FakeCommentRepository) {
				cr.GetCommentByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*comment.Comment, error) {
					return &comment.Comment{}, nil
				}
				cr.UpdateCommentFn = func(_ context.Context, _ string, _ uuid.UUID, _ string) (*comment.Comment, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &FakeCommentRepository{}
			tt.setup(cr)
			svc := service.NewCommentService(nil, cr, &FakeTodoRepository{})
			got, err := svc.UpdateComment(newEchoCtx(), "user-1", commentID, "new body")
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, got.Content)
		})
	}
}

func TestCommentService_DeleteComment(t *testing.T) {
	commentID := uuid.New()
	tests := []struct {
		name    string
		setup   func(cr *FakeCommentRepository)
		wantErr string
	}{
		{
			name: "happy path",
			setup: func(cr *FakeCommentRepository) {
				cr.GetCommentByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*comment.Comment, error) {
					return &comment.Comment{}, nil
				}
				cr.DeleteCommentFn = func(_ context.Context, _ string, _ uuid.UUID) error { return nil }
			},
		},
		{
			name: "validation lookup error propagates",
			setup: func(cr *FakeCommentRepository) {
				cr.GetCommentByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*comment.Comment, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
		{
			name: "delete error propagates",
			setup: func(cr *FakeCommentRepository) {
				cr.GetCommentByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*comment.Comment, error) {
					return &comment.Comment{}, nil
				}
				cr.DeleteCommentFn = func(_ context.Context, _ string, _ uuid.UUID) error { return errRepoBoom }
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &FakeCommentRepository{}
			tt.setup(cr)
			svc := service.NewCommentService(nil, cr, &FakeTodoRepository{})
			err := svc.DeleteComment(newEchoCtx(), "user-1", commentID)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
