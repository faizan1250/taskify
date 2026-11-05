package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func intPtr(i int) *int { return &i }

func TestResolveSortColumn(t *testing.T) {
	tests := []struct {
		name     string
		sort     *string
		allowed  map[string]string
		fallback string
		want     string
	}{
		{"nil sort returns fallback", nil, allowedTodoSortColumns, "t.created_at", "t.created_at"},
		{"valid todo sort maps to qualified column", strPtr("due_date"), allowedTodoSortColumns, "t.created_at", "t.due_date"},
		{"valid category sort maps to column", strPtr("name"), allowedCategorySortColumns, "name", "name"},
		{"unknown sort returns fallback", strPtr("nonexistent"), allowedTodoSortColumns, "t.created_at", "t.created_at"},
		{"empty sort returns fallback", strPtr(""), allowedTodoSortColumns, "t.created_at", "t.created_at"},
		{
			"sql injection in sort returns fallback",
			strPtr("created_at; DROP TABLE todos; --"),
			allowedTodoSortColumns, "t.created_at", "t.created_at",
		},
		{
			"boolean injection in sort returns fallback",
			strPtr("created_at DESC, (SELECT 1 WHERE 1=1)"),
			allowedTodoSortColumns, "t.created_at", "t.created_at",
		},
		{
			"unqualified column name is not trusted",
			strPtr("created_at"),
			map[string]string{"created_at": "t.created_at"}, "t.created_at", "t.created_at",
		},
		{
			"todo sort key rejected for categories",
			strPtr("title"),
			allowedCategorySortColumns, "name", "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveSortColumn(tt.sort, tt.allowed, tt.fallback))
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
			assert.Equal(t, tt.want, resolveSortOrder(tt.order, tt.fallback))
		})
	}
}

func TestNormalizePage(t *testing.T) {
	assert.Equal(t, 1, normalizePage(nil))
	assert.Equal(t, 1, normalizePage(intPtr(0)))
	assert.Equal(t, 1, normalizePage(intPtr(-3)))
	assert.Equal(t, 1, normalizePage(intPtr(1)))
	assert.Equal(t, 4, normalizePage(intPtr(4)))
}

func TestNormalizeLimit(t *testing.T) {
	assert.Equal(t, 20, normalizeLimit(nil, 20))
	assert.Equal(t, 50, normalizeLimit(nil, 50))
	assert.Equal(t, 20, normalizeLimit(intPtr(0), 20))
	assert.Equal(t, 20, normalizeLimit(intPtr(-5), 20))
	assert.Equal(t, 25, normalizeLimit(intPtr(25), 20))
	assert.Equal(t, 100, normalizeLimit(intPtr(100), 20))
	assert.Equal(t, 100, normalizeLimit(intPtr(101), 20))
	assert.Equal(t, 100, normalizeLimit(intPtr(10000), 50))
}
