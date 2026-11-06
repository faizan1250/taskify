package todo

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ------------------------------------------------------------

type CreateTodoPayload struct {
	Title        string      `json:"title" validate:"required,min=1,max=255"`
	Description  *string     `json:"description" validate:"omitempty,max=1000"`
	Priority     *Priority   `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate      *time.Time  `json:"dueDate"`
	ParentTodoID *uuid.UUID  `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID   *uuid.UUID  `json:"categoryId" validate:"omitempty,uuid"`
	Metadata     *Metadata   `json:"metadata"`
	Recurrence   *Recurrence `json:"recurrence" validate:"omitempty,oneof=none daily weekly monthly"`
}

func (p *CreateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type UpdateTodoPayload struct {
	ID           uuid.UUID   `param:"id" validate:"required,uuid"`
	Title        *string     `json:"title" validate:"omitempty,min=1,max=255"`
	Description  *string     `json:"description" validate:"omitempty,max=1000"`
	Status       *Status     `json:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority     *Priority   `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate      *time.Time  `json:"dueDate"`
	ParentTodoID *uuid.UUID  `json:"parentTodoId" validate:"omitempty,uuid"`
	CategoryID   *uuid.UUID  `json:"categoryId" validate:"omitempty,uuid"`
	Metadata     *Metadata   `json:"metadata"`
	Recurrence   *Recurrence `json:"recurrence" validate:"omitempty,oneof=none daily weekly monthly"`
}

func (p *UpdateTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetTodosQuery struct {
	Page         *int        `query:"page" validate:"omitempty,min=1"`
	Limit        *int        `query:"limit" validate:"omitempty,min=1,max=100"`
	Sort         *string     `query:"sort" validate:"omitempty,oneof=created_at updated_at title priority due_date status position"`
	Order        *string     `query:"order" validate:"omitempty,oneof=asc desc"`
	Search       *string     `query:"search" validate:"omitempty,min=1"`
	Status       *Status     `query:"status" validate:"omitempty,oneof=draft active completed archived"`
	Priority     *Priority   `query:"priority" validate:"omitempty,oneof=low medium high"`
	CategoryID   *uuid.UUID  `query:"categoryId" validate:"omitempty,uuid"`
	ParentTodoID *uuid.UUID  `query:"parentTodoId" validate:"omitempty,uuid"`
	DueFrom      *time.Time  `query:"dueFrom"`
	DueTo        *time.Time  `query:"dueTo"`
	Overdue      *bool       `query:"overdue"`
	Completed    *bool       `query:"completed"`
	Recurrence   *Recurrence `query:"recurrence" validate:"omitempty,oneof=none daily weekly monthly"`
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

type GetTodoStatsPayload struct {
	// IncludeCounts opts out of the per-category breakdown (one extra
	// aggregate query) when false. Nil (param absent) defaults to true so
	// the endpoint stays a single call instead of N+1 category queries.
	// DueThisWeek is always computed in the base aggregate query.
	IncludeCounts *bool `query:"includeCounts"`
}

func (p *GetTodoStatsPayload) Validate() error {
	return nil
}

// IncludeCategoryCounts reports whether the per-category breakdown should
// be computed. Absent param means include.
func (p *GetTodoStatsPayload) IncludeCategoryCounts() bool {
	if p == nil || p.IncludeCounts == nil {
		return true
	}
	return *p.IncludeCounts
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
	IDs   []uuid.UUID   `json:"ids" validate:"required,min=1,max=100"`
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

// ------------------------------------------------------------
// Reorder Todos DTOs
// ------------------------------------------------------------

// MaxReorderTodoIDs bounds reorder requests so one call cannot hold a
// transaction open over an unbounded row set.
const MaxReorderTodoIDs = 100

// ReorderTodosPayload carries the caller's full desired order for drag
// reorder. The server assigns position = index, so the payload is the
// complete order, not a sparse patch. Duplicate IDs are rejected: two
// slots for one row is a client bug, not a merge to guess at.
type ReorderTodosPayload struct {
	OrderedIDs []uuid.UUID `json:"orderedIds" validate:"required,min=1,max=100"`
}

func (p *ReorderTodosPayload) Validate() error {
	validate := validator.New()
	if err := validate.Struct(p); err != nil {
		return err
	}
	if err := validateBulkIDs(p.OrderedIDs); err != nil {
		return err
	}
	seen := make(map[uuid.UUID]struct{}, len(p.OrderedIDs))
	for _, id := range p.OrderedIDs {
		if _, ok := seen[id]; ok {
			return errors.New("orderedIds must not contain duplicates")
		}
		seen[id] = struct{}{}
	}
	return nil
}

// PositionsForOrder maps each ID to its zero-based slot as a float position.
// Pure domain helper shared by the repository reorder transaction and
// Docker-free unit tests; float slots leave room for fractional indexing
// without a schema change.
func PositionsForOrder(ids []uuid.UUID) map[uuid.UUID]float64 {
	positions := make(map[uuid.UUID]float64, len(ids))
	for i, id := range ids {
		positions[id] = float64(i)
	}
	return positions
}
