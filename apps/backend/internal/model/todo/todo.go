package todo

import (
	"time"

	"github.com/faizan1250/taskify/internal/model"
	"github.com/faizan1250/taskify/internal/model/category"
	"github.com/faizan1250/taskify/internal/model/comment"
	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusArchived  Status = "archived"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Todo struct {
	model.Base
	UserID            string     `json:"userId" db:"user_id"`
	Title             string     `json:"title" db:"title"`
	Description       *string    `json:"description" db:"description"`
	Status            Status     `json:"status" db:"status"`
	Priority          Priority   `json:"priority" db:"priority"`
	DueDate           *time.Time `json:"dueDate" db:"due_date"`
	CompletedAt       *time.Time `json:"completedAt" db:"completed_at"`
	ParentTodoID      *uuid.UUID `json:"parentTodoId" db:"parent_todo_id"`
	CategoryID        *uuid.UUID `json:"categoryId" db:"category_id"`
	Metadata          *Metadata  `json:"metadata" db:"metadata"`
	SortOrder         int        `json:"sortOrder" db:"sort_order"`
	Position          float64    `json:"position" db:"position"`
	Recurrence        Recurrence `json:"recurrence" db:"recurrence"`
	NextRunAt         *time.Time `json:"nextRunAt" db:"next_run_at"`
	ReminderSentAt    *time.Time `json:"reminderSentAt" db:"reminder_sent_at"`
	OverdueNotifiedAt *time.Time `json:"overdueNotifiedAt" db:"overdue_notified_at"`
}

type Metadata struct {
	Tags       []string `json:"tags"`
	Reminder   *string  `json:"reminder"`
	Color      *string  `json:"color"`
	Difficulty *int     `json:"difficulty"`
}

type PopulatedTodo struct {
	Todo
	Category    *category.Category `json:"category" db:"category"`
	Children    []Todo             `json:"children" db:"children"`
	Comments    []comment.Comment  `json:"comments" db:"comments"`
	Attachments []TodoAttachment   `json:"attachments" db:"attachments"`
}

type TodoStats struct {
	Total       int                 `json:"total" db:"total"`
	Draft       int                 `json:"draft" db:"draft"`
	Active      int                 `json:"active" db:"active"`
	Completed   int                 `json:"completed" db:"completed"`
	Archived    int                 `json:"archived" db:"archived"`
	Overdue     int                 `json:"overdue" db:"overdue"`
	DueThisWeek int                 `json:"dueThisWeek" db:"due_this_week"`
	ByCategory  []CategoryTodoCount `json:"byCategory"`
}

// CategoryTodoCount is one row of the per-category breakdown returned by
// GET /todos/stats. CategoryID is nil for uncategorized todos; Name is ""
// in that case. Count covers all statuses so sum(ByCategory) == Total.
type CategoryTodoCount struct {
	CategoryID *uuid.UUID `json:"categoryId" db:"category_id"`
	Name       string     `json:"name" db:"name"`
	Count      int        `json:"count" db:"count"`
}

type UserWeeklyStats struct {
	UserID         string `json:"userId" db:"user_id"`
	CreatedCount   int    `json:"createdCount" db:"created_count"`
	CompletedCount int    `json:"completedCount" db:"completed_count"`
	ActiveCount    int    `json:"activeCount" db:"active_count"`
	OverdueCount   int    `json:"overdueCount" db:"overdue_count"`
}

func (t *Todo) IsOverdue() bool {
	return t.DueDate != nil && t.DueDate.Before(time.Now()) && t.Status != StatusCompleted
}

func (t *Todo) CanHaveChildren() bool {
	return t.ParentTodoID == nil
}
