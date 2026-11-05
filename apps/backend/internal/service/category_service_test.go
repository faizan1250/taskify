package service_test

import (
	"context"
	"testing"

	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryService_CreateCategory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(cr *FakeCategoryRepository)
		wantErr string
	}{
		{
			name: "happy path",
			setup: func(cr *FakeCategoryRepository) {
				cr.CreateCategoryFn = func(_ context.Context, _ string, _ *category.CreateCategoryPayload) (*category.Category, error) {
					return &category.Category{Name: "work", Color: "#ff0000"}, nil
				}
			},
		},
		{
			name: "repo error propagates",
			setup: func(cr *FakeCategoryRepository) {
				cr.CreateCategoryFn = func(_ context.Context, _ string, _ *category.CreateCategoryPayload) (*category.Category, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &FakeCategoryRepository{}
			tt.setup(cr)
			svc := service.NewCategoryService(nil, cr)
			got, err := svc.CreateCategory(newEchoCtx(), "user-1", &category.CreateCategoryPayload{Name: "work", Color: "#ff0000"})
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "work", got.Name)
		})
	}
}

func TestCategoryService_GetCategories(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(cr *FakeCategoryRepository)
		wantErr   string
		wantTotal int
	}{
		{
			name: "happy path",
			setup: func(cr *FakeCategoryRepository) {
				cr.GetCategoriesFn = func(_ context.Context, _ string, _ *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
					return &model.PaginatedResponse[category.Category]{Data: []category.Category{{Name: "a"}}, Total: 1}, nil
				}
			},
			wantTotal: 1,
		},
		{
			name: "repo error propagates",
			setup: func(cr *FakeCategoryRepository) {
				cr.GetCategoriesFn = func(_ context.Context, _ string, _ *category.GetCategoriesQuery) (*model.PaginatedResponse[category.Category], error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &FakeCategoryRepository{}
			tt.setup(cr)
			svc := service.NewCategoryService(nil, cr)
			got, err := svc.GetCategories(newEchoCtx(), "user-1", &category.GetCategoriesQuery{})
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

func TestCategoryService_GetCategoryByID(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name    string
		setup   func(cr *FakeCategoryRepository)
		wantErr string
	}{
		{
			name: "happy path",
			setup: func(cr *FakeCategoryRepository) {
				cr.GetCategoryByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*category.Category, error) {
					return &category.Category{Name: "home"}, nil
				}
			},
		},
		{
			name: "repo error propagates",
			setup: func(cr *FakeCategoryRepository) {
				cr.GetCategoryByIDFn = func(_ context.Context, _ string, _ uuid.UUID) (*category.Category, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &FakeCategoryRepository{}
			tt.setup(cr)
			svc := service.NewCategoryService(nil, cr)
			got, err := svc.GetCategoryByID(newEchoCtx(), "user-1", id)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "home", got.Name)
		})
	}
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name    string
		setup   func(cr *FakeCategoryRepository)
		wantErr string
	}{
		{
			name: "happy path",
			setup: func(cr *FakeCategoryRepository) {
				cr.UpdateCategoryFn = func(_ context.Context, _ string, _ uuid.UUID, _ *category.UpdateCategoryPayload) (*category.Category, error) {
					return &category.Category{Name: "renamed"}, nil
				}
			},
		},
		{
			name: "repo error propagates",
			setup: func(cr *FakeCategoryRepository) {
				cr.UpdateCategoryFn = func(_ context.Context, _ string, _ uuid.UUID, _ *category.UpdateCategoryPayload) (*category.Category, error) {
					return nil, errRepoBoom
				}
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &FakeCategoryRepository{}
			tt.setup(cr)
			svc := service.NewCategoryService(nil, cr)
			got, err := svc.UpdateCategory(newEchoCtx(), "user-1", id, &category.UpdateCategoryPayload{})
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "renamed", got.Name)
		})
	}
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name    string
		setup   func(cr *FakeCategoryRepository)
		wantErr string
	}{
		{
			name: "happy path",
			setup: func(cr *FakeCategoryRepository) {
				cr.DeleteCategoryFn = func(_ context.Context, _ string, _ uuid.UUID) error { return nil }
			},
		},
		{
			name: "repo error propagates",
			setup: func(cr *FakeCategoryRepository) {
				cr.DeleteCategoryFn = func(_ context.Context, _ string, _ uuid.UUID) error { return errRepoBoom }
			},
			wantErr: "repo boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &FakeCategoryRepository{}
			tt.setup(cr)
			svc := service.NewCategoryService(nil, cr)
			err := svc.DeleteCategory(newEchoCtx(), "user-1", id)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
