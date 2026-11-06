package repository

// maxPageLimit caps client-requested page sizes. It mirrors the
// `max=100` validation rule on the list query DTOs.
const maxPageLimit = 100

// allowedTodoSortColumns maps client-supplied sort keys for todos to safe,
// fully-qualified SQL identifiers. Any value not present in this map falls
// back to the default column so unsanitized input can never reach the query.
var allowedTodoSortColumns = map[string]string{ //nolint:gochecknoglobals // read-only sort-key whitelist shared by todo/category query builders; Go has no const maps and this table is never mutated.
	"created_at": "t.created_at",
	"updated_at": "t.updated_at",
	"title":      "t.title",
	"priority":   "t.priority",
	"due_date":   "t.due_date",
	"status":     "t.status",
	"position":   "t.position",
}

// allowedCategorySortColumns maps client-supplied sort keys for categories
// to safe SQL identifiers.
var allowedCategorySortColumns = map[string]string{ //nolint:gochecknoglobals // read-only sort-key whitelist shared by todo/category query builders; Go has no const maps and this table is never mutated.
	"created_at": "created_at",
	"updated_at": "updated_at",
	"name":       "name",
}

// resolveSortColumn returns the whitelisted SQL identifier for a
// client-supplied sort key, or fallback when sort is nil or not allowed.
func resolveSortColumn(sort *string, allowed map[string]string, fallback string) string {
	if sort == nil {
		return fallback
	}
	if column, ok := allowed[*sort]; ok {
		return column
	}
	return fallback
}

// resolveSortOrder returns the client-supplied sort direction when it is
// exactly "asc" or "desc", or fallback otherwise.
func resolveSortOrder(order *string, fallback string) string {
	if order == nil {
		return fallback
	}
	switch *order {
	case "asc", "desc":
		return *order
	default:
		return fallback
	}
}

// normalizePage returns a safe page number, defaulting to 1 when page is
// nil or out of range.
func normalizePage(page *int) int {
	if page == nil || *page < 1 {
		return 1
	}
	return *page
}

// normalizeLimit returns a safe page size, defaulting to defaultLimit when
// limit is nil or out of range and capping at maxPageLimit.
func normalizeLimit(limit *int, defaultLimit int) int {
	if limit == nil || *limit < 1 {
		return defaultLimit
	}
	if *limit > maxPageLimit {
		return maxPageLimit
	}
	return *limit
}
