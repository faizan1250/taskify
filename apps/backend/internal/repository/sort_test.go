package repository_test

import (
	"testing"

	"github.com/faizan1250/taskify/internal/repository"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func intPtr(i int) *int { return &i }

func TestResolveSortColumn(t *testing.T) {
	allowedTodo := repository.AllowedTodoSortColumnsForTest()
	allowedCategory := repository.AllowedCategorySortColumnsForTest()
	tests := []struct {
		name     string
		sort     *string
		allowed  map[string]string
		fallback string
		want     string
	}{
		{"nil sort returns fallback", nil, allowedTodo, "t.created_at", "t.created_at"},
		{"valid todo sort maps to qualified column", strPtr("due_date"), allowedTodo, "t.created_at", "t.due_date"},
		{"valid category sort maps to column", strPtr("name"), allowedCategory, "name", "name"},
		{"unknown sort returns fallback", strPtr("nonexistent"), allowedTodo, "t.created_at", "t.created_at"},
		{"empty sort returns fallback", strPtr(""), allowedTodo, "t.created_at", "t.created_at"},
		{
			"sql injection in sort returns fallback",
			strPtr("created_at; DROP TABLE todos; --"),
			allowedTodo, "t.created_at", "t.created_at",
		},
		{
			"boolean injection in sort returns fallback",
			strPtr("created_at DESC, (SELECT 1 WHERE 1=1)"),
			allowedTodo, "t.created_at", "t.created_at",
		},
		{
			"unqualified column name is not trusted",
			strPtr("created_at"),
			map[string]string{"created_at": "t.created_at"}, "t.created_at", "t.created_at",
		},
		{
			"todo sort key rejected for categories",
			strPtr("title"),
			allowedCategory, "name", "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, repository.ResolveSortColumnForTest(tt.sort, tt.allowed, tt.fallback))
		})
	}
}

func TestResolveSortOrder(t *testing.T) {
	tests := []struct {
		name     string
		order    *string
		fallback string
		want     string
	}{
		{"nil order returns fallback", nil, "DESC", "DESC"},
		{"asc passes through", strPtr("asc"), "DESC", "asc"},
		{"desc passes through", strPtr("desc"), "asc", "desc"},
		{"uppercase rejected", strPtr("DESC"), "asc", "asc"},
		{"injection rejected", strPtr("desc; DROP TABLE todos; --"), "asc", "asc"},
		{"empty rejected", strPtr(""), "asc", "asc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, repository.ResolveSortOrderForTest(tt.order, tt.fallback))
		})
	}
}

func TestNormalizePage(t *testing.T) {
	assert.Equal(t, 1, repository.NormalizePageForTest(nil))
	assert.Equal(t, 1, repository.NormalizePageForTest(intPtr(0)))
	assert.Equal(t, 1, repository.NormalizePageForTest(intPtr(-3)))
	assert.Equal(t, 1, repository.NormalizePageForTest(intPtr(1)))
	assert.Equal(t, 4, repository.NormalizePageForTest(intPtr(4)))
}

func TestNormalizeLimit(t *testing.T) {
	assert.Equal(t, 20, repository.NormalizeLimitForTest(nil, 20))
	assert.Equal(t, 50, repository.NormalizeLimitForTest(nil, 50))
	assert.Equal(t, 20, repository.NormalizeLimitForTest(intPtr(0), 20))
	assert.Equal(t, 20, repository.NormalizeLimitForTest(intPtr(-5), 20))
	assert.Equal(t, 25, repository.NormalizeLimitForTest(intPtr(25), 20))
	assert.Equal(t, 100, repository.NormalizeLimitForTest(intPtr(100), 20))
	assert.Equal(t, 100, repository.NormalizeLimitForTest(intPtr(101), 20))
	assert.Equal(t, 100, repository.NormalizeLimitForTest(intPtr(10000), 50))
}
