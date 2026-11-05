package repository

// Export shims exposing the unexported sort/pagination helpers to black-box
// tests (package repository_test), following the testpackage linter and the
// repo convention already used by config, cron, router, service, and
// validation packages.

func ResolveSortColumnForTest(sort *string, allowed map[string]string, fallback string) string {
	return resolveSortColumn(sort, allowed, fallback)
}

func ResolveSortOrderForTest(order *string, fallback string) string {
	return resolveSortOrder(order, fallback)
}

func NormalizePageForTest(page *int) int {
	return normalizePage(page)
}

func NormalizeLimitForTest(limit *int, defaultLimit int) int {
	return normalizeLimit(limit, defaultLimit)
}

func AllowedTodoSortColumnsForTest() map[string]string {
	return allowedTodoSortColumns
}

func AllowedCategorySortColumnsForTest() map[string]string {
	return allowedCategorySortColumns
}
