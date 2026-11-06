package todo

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ------------------------------------------------------------

type CreateTodoPayload struct {
	Title        string     `json:"title" validate:"required,min=1,max=255"`
	Description  *string    `json:"description" validate:"omitempty,max=1000"`
	Priority     *Priority  `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate      *time.Time `json:"dueDate"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID   *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"`
	Metadata     *Metadata  `json:"metadata"`
}

func (p *CreateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type UpdateTodoPayload struct {
	ID           uuid.UUID  `param:"id" validate:"required,uuid"`
	Title        *string    `json:"title" validate:"omitempty,min=1,max=255"`
	Description  *string    `json:"description" validate:"omitempty,max=1000"`
	Status       *Status    `json:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority     *Priority  `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate      *time.Time `json:"dueDate"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID   *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"`
	Metadata     *Metadata  `json:"metadata"`
}

func (p *UpdateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetTodosQuery struct {
	Page         *int       `query:"page" validate:"omitempty,min=1"`
	Limit        *int       `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort         *string    `query:"sort" validate:"omitempty,oneof=created_at updated_at title priority due_date status"`
	Order        *string    `query:"order" validate:"omitempty,oneof=asc desc"`
	Search       *string    `query:"search" validate:"omitempty,min=1"`
	Status       *Status    `query:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority     *Priority  `query:"priority" validate:"omitempty,oneof=low medium high"`
	CategoryID   *uuid.UUID `query:"categoryId" validate:"omitempty,uuid"`
	ParentTodoID *uuid.UUID `query:"parentTodoId" validate:"omitempty,uuid"`
	DueFrom      *time.Time `query:"dueFrom"`
	DueTo        *time.Time `query:"dueTo"`
	Overdue      *bool      `query:"overdue"`
	Completed    *bool      `query:"completed"`
}

func (q *GetTodosQuery) Validate() error {
	validate := validator.New()

	if err := validate.Struct(q); err != nil {
		return err
	}

	// Set defaults for pagination
	if q.Page == nil {
		defaultPage := 1
		q.Page = &defaultPage
	}
	if q.Limit == nil {
		defaultLimit := 20
		q.Limit = &defaultLimit
	}
	if q.Sort == nil {
		defaultSort := "created_at"
		q.Sort = &defaultSort
	}
	if q.Order == nil {
		defaultOrder := "desc"
		q.Order = &defaultOrder
	}

	return nil
}

// ------------------------------------------------------------

type GetTodoByIDPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *GetTodoByIDPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type DeleteTodoPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetTodoStatsPayload struct{}

func (p *GetTodoStatsPayload) Validate() error {
	return nil
}

// ------------------------------------------------------------
// Todo Attachment DTOs
// ------------------------------------------------------------

type UploadTodoAttachmentPayload struct {
	TodoID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *UploadTodoAttachmentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type DeleteTodoAttachmentPayload struct {
	TodoID       uuid.UUID `param:"id" validate:"required,uuid"`
	AttachmentID uuid.UUID `param:"attachmentId" validate:"required,uuid"`
}

func (p *DeleteTodoAttachmentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetAttachmentPresignedURLPayload struct {
	TodoID       uuid.UUID `param:"id" validate:"required,uuid"`
	AttachmentID uuid.UUID `param:"attachmentId" validate:"required,uuid"`
}

func (p *GetAttachmentPresignedURLPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------
// Bulk Todo DTOs
// ------------------------------------------------------------

// MaxBulkTodoIDs bounds bulk requests so one call cannot hold a
// transaction open over an unbounded row set.
const MaxBulkTodoIDs = 100

// BulkTodoPatch is the shared per-item mutation applied to every ID in a
// bulk update. At least one field must be set; nil fields are left untouched.
type BulkTodoPatch struct {
	Status     *Status    `json:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority   *Priority  `json:"priority" validate:"omitempty,oneof=low medium high"`
	CategoryID *uuid.UUID `json:"categoryId" validate:"omitempty,uuid"`
	DueDate    *time.Time `json:"dueDate"`
}

// IsEmpty reports whether the patch carries no mutation.
func (p BulkTodoPatch) IsEmpty() bool {
	return p.Status == nil && p.Priority == nil && p.CategoryID == nil && p.DueDate == nil
}

// BulkTodoFailure is the per-ID error entry of a partial-failure report.
// IDs absent from both the success list and Failed were never attempted,
// which only happens when the whole operation fails with a 4xx/5xx instead.
type BulkTodoFailure struct {
	ID    uuid.UUID `json:"id"`
	Error string    `json:"error"`
}

// BulkUpdateTodosPayload applies Patch to every todo in IDs.
type BulkUpdateTodosPayload struct {
	IDs   []uuid.UUID  `json:"ids" validate:"required,min=1,max=100"`
	Patch BulkTodoPatch `json:"patch"`
}

func (p *BulkUpdateTodosPayload) Validate() error {
	validate := validator.New()
	if err := validate.Struct(p); err != nil {
		return err
	}
	if err := validateBulkIDs(p.IDs); err != nil {
		return err
	}
	if p.Patch.IsEmpty() {
		return errors.New("patch must include at least one of: status, priority, categoryId, dueDate")
	}
	return nil
}

// BulkUpdateTodosResult carries the partial-failure report for a bulk update.
type BulkUpdateTodosResult struct {
	Updated []Todo            `json:"updated"`
	Failed  []BulkTodoFailure `json:"failed"`
}

// HasFailures reports whether any ID failed while others succeeded.
func (r *BulkUpdateTodosResult) HasFailures() bool {
	return r != nil && len(r.Failed) > 0
}

// BulkDeleteTodosPayload deletes every todo in IDs.
type BulkDeleteTodosPayload struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1,max=100"`
}

func (p *BulkDeleteTodosPayload) Validate() error {
	validate := validator.New()
	if err := validate.Struct(p); err != nil {
		return err
	}
	return validateBulkIDs(p.IDs)
}

// BulkDeleteTodosResult carries the partial-failure report for a bulk delete.
type BulkDeleteTodosResult struct {
	Deleted []uuid.UUID       `json:"deleted"`
	Failed  []BulkTodoFailure `json:"failed"`
}

// HasFailures reports whether any ID failed while others succeeded.
func (r *BulkDeleteTodosResult) HasFailures() bool {
	return r != nil && len(r.Failed) > 0
}

// validateBulkIDs rejects nil UUID entries that JSON or callers may sneak
// past the length-only struct tags.
func validateBulkIDs(ids []uuid.UUID) error {
	for _, id := range ids {
		if id == uuid.Nil {
			return errors.New("ids must not contain empty UUIDs")
		}
	}
	return nil
}
