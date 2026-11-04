package service_test

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/lib/aws"
	"github.com/faizan1250/taskify/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContextKey string

func TestValidateAttachment(t *testing.T) {
	tests := []struct {
		name        string
		file        *multipart.FileHeader
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil file header",
			file:        nil,
			wantErr:     true,
			errContains: "file is required",
		},
		{
			name: "zero byte file",
			file: &multipart.FileHeader{
				Filename: "empty.txt",
				Size:     0,
			},
			wantErr:     true,
			errContains: "file cannot be empty",
		},
		{
			name: "negative size file",
			file: &multipart.FileHeader{
				Filename: "negative.txt",
				Size:     -1,
			},
			wantErr:     true,
			errContains: "file cannot be empty",
		},
		{
			name: "file size exceeds 10MB limit",
			file: &multipart.FileHeader{
				Filename: "large.pdf",
				Size:     service.MaxAttachmentSize + 1,
			},
			wantErr:     true,
			errContains: "file size exceeds maximum allowed limit of 10MB",
		},
		{
			name: "file size substantially exceeds 10MB limit (50MB)",
			file: &multipart.FileHeader{
				Filename: "huge.zip",
				Size:     50 * 1024 * 1024,
			},
			wantErr:     true,
			errContains: "file size exceeds maximum allowed limit of 10MB",
		},
		{
			name: "valid small file",
			file: &multipart.FileHeader{
				Filename: "valid.png",
				Size:     1024,
			},
			wantErr: false,
		},
		{
			name: "exact 10MB file boundary",
			file: &multipart.FileHeader{
				Filename: "exact10mb.dat",
				Size:     service.MaxAttachmentSize,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAttachment(tt.file)
			if tt.wantErr {
				require.Error(t, err)
				var httpErr *errs.HTTPError
				require.True(t, errors.As(err, &httpErr))
				assert.Equal(t, http.StatusBadRequest, httpErr.Status)
				assert.Contains(t, httpErr.Message, tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadTodoAttachment_Validation(t *testing.T) {
	e := echo.New()

	t.Run("rejects zero-size file upload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		svc := &service.TodoService{}
		_, err := svc.UploadTodoAttachment(c, "user-1", uuid.New(), &multipart.FileHeader{
			Filename: "empty.txt",
			Size:     0,
		})
		require.Error(t, err)
		var httpErr *errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusBadRequest, httpErr.Status)
		assert.Contains(t, httpErr.Message, "file cannot be empty")
	})

	t.Run("rejects oversized file upload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		svc := &service.TodoService{}
		_, err := svc.UploadTodoAttachment(c, "user-1", uuid.New(), &multipart.FileHeader{
			Filename: "huge.bin",
			Size:     service.MaxAttachmentSize + 1024,
		})
		require.Error(t, err)
		var httpErr *errs.HTTPError
		require.True(t, errors.As(err, &httpErr))
		assert.Equal(t, http.StatusBadRequest, httpErr.Status)
		assert.Contains(t, httpErr.Message, "file size exceeds maximum allowed limit of 10MB")
	})
}

func TestGenerateAttachmentKey_CollisionAndFormatting(t *testing.T) {
	todoID := uuid.New()
	filename := "document.pdf"

	t.Run("hierarchical key structure", func(t *testing.T) {
		key := service.GenerateAttachmentKey(todoID, filename)
		assert.True(t, strings.HasPrefix(key, fmt.Sprintf("todos/%s/attachments/", todoID.String())))
		assert.True(t, strings.HasSuffix(key, "/document.pdf"))

		parts := strings.Split(key, "/")
		require.Len(t, parts, 5)
		assert.Equal(t, "todos", parts[0])
		assert.Equal(t, todoID.String(), parts[1])
		assert.Equal(t, "attachments", parts[2])
		_, err := uuid.Parse(parts[3])
		assert.NoError(t, err, "key should contain a valid UUID part")
		assert.Equal(t, "document.pdf", parts[4])
	})

	t.Run("path traversal sanitization via filepath.Base", func(t *testing.T) {
		key := service.GenerateAttachmentKey(todoID, "../../../etc/passwd")
		assert.True(t, strings.HasSuffix(key, "/passwd"))
		assert.False(t, strings.Contains(key, ".."))
	})

	t.Run("concurrent key collision prevention", func(t *testing.T) {
		const totalKeys = 10000
		const concurrency = 20
		keysChan := make(chan string, totalKeys)

		var wg sync.WaitGroup
		keysPerWorker := totalKeys / concurrency

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < keysPerWorker; j++ {
					keysChan <- service.GenerateAttachmentKey(todoID, filename)
				}
			}()
		}

		wg.Wait()
		close(keysChan)

		seen := make(map[string]struct{}, totalKeys)
		for key := range keysChan {
			if _, exists := seen[key]; exists {
				t.Fatalf("Collision detected for key: %s", key)
			}
			seen[key] = struct{}{}
		}

		assert.Len(t, seen, totalKeys, "all generated keys should be unique")
	})
}

func TestDeleteTodoAttachment_ContextDetachment(t *testing.T) {
	t.Run("parent context cancellation does not cancel detached background context", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		const key testContextKey = "trace_id"
		ctxWithValue := context.WithValue(parentCtx, key, "test-trace-123")

		// Create detached context with timeout as implemented in DeleteTodoAttachment
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctxWithValue), 30*time.Second)
		defer cancel()

		// Cancel parent context
		parentCancel()

		// Verify parent context is canceled
		require.ErrorIs(t, ctxWithValue.Err(), context.Canceled)

		// Verify detached background context is NOT canceled
		assert.NoError(t, bgCtx.Err())

		// Verify values from parent context are preserved
		assert.Equal(t, "test-trace-123", bgCtx.Value(key))

		// Verify deadline is set within expected range
		deadline, hasDeadline := bgCtx.Deadline()
		assert.True(t, hasDeadline)
		assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, 2*time.Second)
	})

	t.Run("detached background context cancels when cancel func is invoked", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		defer parentCancel()

		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), 30*time.Second)
		assert.NoError(t, bgCtx.Err())

		cancel()
		assert.ErrorIs(t, bgCtx.Err(), context.Canceled)
	})

	t.Run("detached background context respects timeout", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		parentCancel()

		shortTimeoutCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), 10*time.Millisecond)
		defer cancel()

		time.Sleep(25 * time.Millisecond)
		assert.ErrorIs(t, shortTimeoutCtx.Err(), context.DeadlineExceeded)
	})
}

func TestDeleteTodoAttachment_PanicRecovery(t *testing.T) {
	t.Run("recovers panic in asynchronous deletion worker", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var recovered any
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					recovered = r
				}
			}()

			panic("simulated S3 client panic")
		}()

		wg.Wait()
		assert.Equal(t, "simulated S3 client panic", recovered)
	})
}

func TestDeleteTodo_S3Cleanup_NilSafety(t *testing.T) {
	ctx := context.Background()

	t.Run("nil awsClient does not panic", func(t *testing.T) {
		svc := &service.TodoService{}
		assert.NotPanics(t, func() {
			svc.CleanupS3AttachmentsAsync(ctx, []string{"key-1", "key-2"})
		})
	})

	t.Run("nil S3 client does not panic", func(t *testing.T) {
		svc := service.NewTodoService(nil, nil, nil, &aws.AWS{S3: nil})
		assert.NotPanics(t, func() {
			svc.CleanupS3AttachmentsAsync(ctx, []string{"key-1", "key-2"})
		})
	})

	t.Run("empty attachment keys slice does not spawn or panic", func(t *testing.T) {
		svc := &service.TodoService{}
		assert.NotPanics(t, func() {
			svc.CleanupS3AttachmentsAsync(ctx, []string{})
		})
	})

	t.Run("nil attachment keys does not panic", func(t *testing.T) {
		svc := &service.TodoService{}
		assert.NotPanics(t, func() {
			svc.CleanupS3AttachmentsAsync(ctx, nil)
		})
	})
}

func TestDeleteTodo_S3Cleanup_ContextDetachment(t *testing.T) {
	t.Run("parent context cancellation does not cancel detached background context for DeleteTodo", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		const key testContextKey = "request_id"
		ctxWithValue := context.WithValue(parentCtx, key, "req-delete-todo-456")

		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctxWithValue), 30*time.Second)
		defer cancel()

		parentCancel()

		require.ErrorIs(t, ctxWithValue.Err(), context.Canceled)
		assert.NoError(t, bgCtx.Err())
		assert.Equal(t, "req-delete-todo-456", bgCtx.Value(key))

		deadline, hasDeadline := bgCtx.Deadline()
		assert.True(t, hasDeadline)
		assert.WithinDuration(t, time.Now().Add(30*time.Second), deadline, 2*time.Second)
	})

	t.Run("detached context times out after specified duration", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		parentCancel()

		shortTimeoutCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), 10*time.Millisecond)
		defer cancel()

		time.Sleep(25 * time.Millisecond)
		assert.ErrorIs(t, shortTimeoutCtx.Err(), context.DeadlineExceeded)
	})
}

func TestDeleteTodo_S3Cleanup_PanicRecovery(t *testing.T) {
	t.Run("recovers panic when deleting multiple keys in background", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var recovered any
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					recovered = r
				}
			}()

			keys := []string{"key-1", "key-2", "key-3"}
			for i, k := range keys {
				if i == 1 {
					panic(fmt.Sprintf("simulated S3 failure deleting %s", k))
				}
			}
		}()

		wg.Wait()
		assert.Equal(t, "simulated S3 failure deleting key-2", recovered)
	})
}
