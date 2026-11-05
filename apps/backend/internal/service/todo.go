package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/faizan1250/taskify/internal/errs"
	"github.com/faizan1250/taskify/internal/lib/aws"
	"github.com/faizan1250/taskify/internal/middleware"
	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/todo"
	"github.com/faizan1250/taskify/internal/repository"
	"github.com/faizan1250/taskify/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type TodoService struct {
	server       *server.Server
	todoRepo     repository.TodoRepositoryInterface
	categoryRepo repository.CategoryRepositoryInterface
	awsClient    *aws.AWS
}

func NewTodoService(server *server.Server, todoRepo repository.TodoRepositoryInterface,
	categoryRepo repository.CategoryRepositoryInterface, awsClient *aws.AWS,
) *TodoService {
	return &TodoService{
		server:       server,
		todoRepo:     todoRepo,
		categoryRepo: categoryRepo,
		awsClient:    awsClient,
	}
}

// s3Available reports whether attachment storage (AWS S3 client) is configured.
func (s *TodoService) s3Available() bool {
	return s != nil && s.awsClient != nil && s.awsClient.S3 != nil
}

// errS3Unavailable is returned when attachment storage is not configured.
func errS3Unavailable() error {
	return &errs.HTTPError{
		Code:    errs.MakeUpperCaseWithUnderscores(http.StatusText(http.StatusServiceUnavailable)),
		Message: "attachment storage is currently unavailable",
		Status:  http.StatusServiceUnavailable,
	}
}

func (s *TodoService) CreateTodo(ctx echo.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	// Validate parent todo exists and belongs to user (if provided)
	if payload.ParentTodoID != nil {
		parentTodo, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, *payload.ParentTodoID)
		if err != nil {
			logger.Error().Err(err).Msg("parent todo validation failed")
			return nil, err
		}

		if !parentTodo.CanHaveChildren() {
			err := errs.NewBadRequestError("Parent todo cannot have children (subtasks can't have subtasks)", false, nil, nil, nil)
			logger.Warn().Msg("parent todo cannot have children")
			return nil, err
		}
	}

	// Validate category exists and belongs to user (if provided)
	if payload.CategoryID != nil {
		_, err := s.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, *payload.CategoryID)
		if err != nil {
			logger.Error().Err(err).Msg("category validation failed")
			return nil, err
		}
	}

	todoItem, err := s.todoRepo.CreateTodo(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create todo")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_created").
		Str("todo_id", todoItem.ID.String()).
		Str("title", todoItem.Title).
		Str("category_id", func() string {
			if todoItem.CategoryID != nil {
				return todoItem.CategoryID.String()
			}
			return ""
		}()).
		Str("priority", string(todoItem.Priority)).
		Msg("Todo created successfully")

	return todoItem, nil
}

func (s *TodoService) GetTodoByID(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.PopulatedTodo, error) {
	logger := middleware.GetLogger(ctx)

	todoItem, err := s.todoRepo.GetTodoByID(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch todo by ID")
		return nil, err
	}

	return todoItem, nil
}

func (s *TodoService) GetTodos(ctx echo.Context, userID string, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
	logger := middleware.GetLogger(ctx)

	result, err := s.todoRepo.GetTodos(ctx.Request().Context(), userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch todos")
		return nil, err
	}

	return result, nil
}

func (s *TodoService) UpdateTodo(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	// Validate parent todo exists and belongs to user (if provided)
	if payload.ParentTodoID != nil {
		parentTodo, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, *payload.ParentTodoID)
		if err != nil {
			logger.Error().Err(err).Msg("parent todo validation failed")
			return nil, err
		}

		if parentTodo.ID == payload.ID {
			err := errs.NewBadRequestError("Todo cannot be its own parent", false, nil, nil, nil)
			logger.Warn().Msg("todo cannot be its own parent")
			return nil, err
		}

		if !parentTodo.CanHaveChildren() {
			err := errs.NewBadRequestError("Parent todo cannot have children (subtasks can't have subtasks)", false, nil, nil, nil)
			logger.Warn().Msg("parent todo cannot have children")
			return nil, err
		}

		logger.Debug().Msg("parent todo validation passed")
	}

	// Validate category exists and belongs to user (if provided)
	if payload.CategoryID != nil {
		_, err := s.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, *payload.CategoryID)
		if err != nil {
			logger.Error().Err(err).Msg("category validation failed")
			return nil, err
		}

		logger.Debug().Msg("category validation passed")
	}

	updatedTodo, err := s.todoRepo.UpdateTodo(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update todo")
		return nil, err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_updated").
		Str("todo_id", updatedTodo.ID.String()).
		Str("title", updatedTodo.Title).
		Str("category_id", func() string {
			if updatedTodo.CategoryID != nil {
				return updatedTodo.CategoryID.String()
			}
			return ""
		}()).
		Str("priority", string(updatedTodo.Priority)).
		Str("status", string(updatedTodo.Status)).
		Msg("Todo updated successfully")

	return updatedTodo, nil
}

func (s *TodoService) DeleteTodo(ctx echo.Context, userID string, todoID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)
	reqCtx := ctx.Request().Context()

	attachmentKeys, err := s.todoRepo.GetAttachmentKeysForTodoAndSubtasks(reqCtx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get attachment keys for todo and subtasks")
		return err
	}

	err = s.todoRepo.DeleteTodo(reqCtx, userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete todo")
		return err
	}

	s.CleanupS3AttachmentsAsync(reqCtx, attachmentKeys)

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_deleted").
		Str("todo_id", todoID.String()).
		Msg("Todo deleted successfully")

	return nil
}

// s3DeleteMaxAttempts bounds async S3 delete retries so a transient
// failure does not permanently orphan an object.
const s3DeleteMaxAttempts = 3

// s3DeleteRetryBaseDelay is the base backoff between async S3 delete
// attempts; each retry waits twice as long as the previous one.
const s3DeleteRetryBaseDelay = 100 * time.Millisecond

// s3ObjectDeleter abstracts S3 object deletion for retry logic.
// *aws.S3Client satisfies it; tests inject scripted stubs.
type s3ObjectDeleter interface {
	DeleteObject(ctx context.Context, bucket, key string) error
}

// deleteS3ObjectWithRetry deletes one S3 object, retrying transient
// failures with exponential backoff. It reports how many attempts ran so
// callers can log escalation context on final failure.
func deleteS3ObjectWithRetry(ctx context.Context, deleter s3ObjectDeleter, bucket, key string) (int, error) {
	attempts := 0
	var err error
	for attempt := 1; attempt <= s3DeleteMaxAttempts; attempt++ {
		attempts = attempt
		if err = deleter.DeleteObject(ctx, bucket, key); err == nil {
			return attempts, nil
		}
		if attempt < s3DeleteMaxAttempts {
			select {
			case <-ctx.Done():
				return attempts, ctx.Err()
			case <-time.After(s3DeleteRetryBaseDelay << (attempt - 1)):
			}
		}
	}
	return attempts, err
}

func (s *TodoService) CleanupS3AttachmentsAsync(reqCtx context.Context, keys []string) {
	if len(keys) == 0 || s.awsClient == nil || s.awsClient.S3 == nil {
		return
	}

	bucket := ""
	if s.server != nil && s.server.Config != nil {
		bucket = s.server.Config.AWS.UploadBucket
	}

	go func() {
		defer func() {
			if r := recover(); r != nil && s.server != nil {
				s.server.Logger.Error().Interface("panic", r).Msg("recovered panic in S3 attachment deletion")
			}
		}()

		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(reqCtx), 30*time.Second)
		defer cancel()

		for _, key := range keys {
			attempts, err := deleteS3ObjectWithRetry(bgCtx, s.awsClient.S3, bucket, key)
			if err == nil {
				if attempts > 1 && s.server != nil {
					s.server.Logger.Warn().
						Str("s3_key", key).
						Int("attempts", attempts).
						Msg("deleted attachment from S3 after retry")
				}
				continue
			}
			if s.server != nil {
				s.server.Logger.Error().
					Err(err).
					Str("s3_key", key).
					Int("attempts", attempts).
					Msg("failed to delete attachment from S3")
			}
		}
	}()
}

func (s *TodoService) GetTodoStats(ctx echo.Context, userID string) (*todo.TodoStats, error) {
	logger := middleware.GetLogger(ctx)

	stats, err := s.todoRepo.GetTodoStats(ctx.Request().Context(), userID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch todo statistics")
		return nil, err
	}

	return stats, nil
}

// MaxAttachmentSize defines the maximum allowed size for uploaded attachments (10MB).
const MaxAttachmentSize = 10 * 1024 * 1024

// ValidateAttachment validates the attachment file header constraints.
func ValidateAttachment(file *multipart.FileHeader) error {
	if file == nil {
		return errs.NewBadRequestError("file is required", false, nil, nil, nil)
	}
	if file.Size <= 0 {
		return errs.NewBadRequestError("file cannot be empty", false, nil, nil, nil)
	}
	if file.Size > MaxAttachmentSize {
		return errs.NewBadRequestError("file size exceeds maximum allowed limit of 10MB", false, nil, nil, nil)
	}
	return nil
}

// GenerateAttachmentKey builds a collision-resistant S3 key scoped to the todo and a unique UUID.
func GenerateAttachmentKey(todoID uuid.UUID, filename string) string {
	return fmt.Sprintf("todos/%s/attachments/%s/%s", todoID.String(), uuid.New().String(), filepath.Base(filename))
}

func (s *TodoService) UploadTodoAttachment(
	ctx echo.Context,
	userID string,
	todoID uuid.UUID,
	file *multipart.FileHeader,
) (*todo.TodoAttachment, error) {
	logger := middleware.GetLogger(ctx)

	// Validate attachment size and presence
	if err := ValidateAttachment(file); err != nil {
		logger.Error().Err(err).Msg("attachment validation failed")
		return nil, err
	}

	// Verify todo exists and belongs to user
	_, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return nil, err
	}

	// Fail fast when attachment storage is unconfigured instead of
	// panicking on a nil S3 client.
	if !s.s3Available() {
		logger.Error().Msg("attachment upload rejected: S3 client is not configured")
		return nil, errS3Unavailable()
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		logger.Error().Err(err).Msg("failed to open uploaded file")
		return nil, errs.NewBadRequestError("failed to open uploaded file", false, nil, nil, nil)
	}
	defer func() {
		_ = src.Close()
	}()

	// Detect MIME type
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		logger.Error().Err(err).Msg("failed to read file for MIME detection")
		return nil, errs.NewBadRequestError("failed to process file", false, nil, nil, nil)
	}
	mimeType := http.DetectContentType(buffer[:n])

	// Rewind file stream before upload
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		logger.Error().Err(err).Msg("failed to rewind file stream")
		return nil, errs.NewBadRequestError("failed to process file", false, nil, nil, nil)
	}

	s3Key := GenerateAttachmentKey(todoID, file.Filename)

	// Upload to S3
	s3Key, err = s.awsClient.S3.UploadFile(
		ctx.Request().Context(),
		s.server.Config.AWS.UploadBucket,
		s3Key,
		src,
		file.Size,
		mimeType,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to upload file to S3")
		return nil, errors.Wrap(err, "failed to upload file")
	}

	// Create attachment record
	attachment, err := s.todoRepo.UploadTodoAttachment(
		ctx.Request().Context(),
		todoID,
		userID,
		s3Key,
		file.Filename,
		file.Size,
		mimeType,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create attachment record")

		reqCtx := ctx.Request().Context()
		cleanupKey := s3Key

		// Delete orphaned upload from S3 asynchronously using detached context
		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.server.Logger.Error().Interface("panic", r).Msg("recovered panic in S3 attachment rollback cleanup")
				}
			}()

			// Skip when storage is unconfigured instead of panicking on a nil S3 client.
			if !s.s3Available() {
				return
			}

			bucket := s.server.Config.AWS.UploadBucket
			bgCtx, cancel := context.WithTimeout(context.WithoutCancel(reqCtx), 30*time.Second)
			defer cancel()

			if attempts, cleanErr := deleteS3ObjectWithRetry(bgCtx, s.awsClient.S3, bucket, cleanupKey); cleanErr != nil {
				s.server.Logger.Error().
					Err(cleanErr).
					Str("s3_key", cleanupKey).
					Int("attempts", attempts).
					Msg("failed to cleanup orphaned attachment from S3 after DB failure")
			}
		}()

		return nil, err
	}

	logger.Info().
		Str("attachment_id", attachment.ID.String()).
		Str("s3_key", s3Key).
		Msg("uploaded todo attachment")

	return attachment, nil
}

func (s *TodoService) DeleteTodoAttachment(
	ctx echo.Context,
	userID string,
	todoID uuid.UUID,
	attachmentID uuid.UUID,
) error {
	logger := middleware.GetLogger(ctx)

	// Verify todo exists and belongs to user
	_, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return err
	}

	// Get attachment details for S3 deletion
	attachment, err := s.todoRepo.GetTodoAttachment(
		ctx.Request().Context(),
		todoID,
		attachmentID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get attachment details")
		return err
	}

	// Delete attachment record
	err = s.todoRepo.DeleteTodoAttachment(
		ctx.Request().Context(),
		todoID,
		attachmentID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete attachment record")
		return err
	}

	reqCtx := ctx.Request().Context()
	downloadKey := attachment.DownloadKey

	// Delete from S3 asynchronously
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.server.Logger.Error().Interface("panic", r).Msg("recovered panic in S3 attachment deletion")
			}
		}()

		// Best-effort delete: skip when storage is unconfigured instead of
		// panicking on a nil S3 client.
		if !s.s3Available() {
			return
		}

		bucket := s.server.Config.AWS.UploadBucket
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(reqCtx), 30*time.Second)
		defer cancel()

		attempts, err := deleteS3ObjectWithRetry(
			bgCtx,
			s.awsClient.S3,
			bucket,
			downloadKey,
		)
		if err != nil {
			s.server.Logger.Error().
				Err(err).
				Str("s3_key", downloadKey).
				Int("attempts", attempts).
				Msg("failed to delete attachment from S3")
		}
	}()

	logger.Info().Msg("deleted todo attachment")

	return nil
}

func (s *TodoService) GetAttachmentPresignedURL(
	ctx echo.Context,
	userID string,
	todoID uuid.UUID,
	attachmentID uuid.UUID,
) (string, error) {
	logger := middleware.GetLogger(ctx)

	// Verify todo exists and belongs to user
	_, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return "", err
	}

	// Get attachment details
	attachment, err := s.todoRepo.GetTodoAttachment(
		ctx.Request().Context(),
		todoID,
		attachmentID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get attachment details")
		return "", err
	}

	// Fail fast when attachment storage is unconfigured instead of
	// panicking on a nil S3 client.
	if !s.s3Available() {
		logger.Error().Msg("presigned URL generation rejected: S3 client is not configured")
		return "", errS3Unavailable()
	}

	// Generate presigned URL
	url, err := s.awsClient.S3.CreatePresignedUrl(
		ctx.Request().Context(),
		s.server.Config.AWS.UploadBucket,
		attachment.DownloadKey,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate presigned URL")
		return "", err
	}

	return url, nil
}
