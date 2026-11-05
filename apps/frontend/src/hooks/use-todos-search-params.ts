import { useDebounce } from "@/api/hooks/use-debounce";
import type { TGetTodosQuery } from "@/api/hooks/use-todo-query";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

export type TodosTab = "draft" | "active" | "completed" | "archived" | "all";
export type TodosPriorityFilter = "low" | "medium" | "high" | "all";
export type TodosSort = NonNullable<TGetTodosQuery["sort"]>;
export type TodosOrder = "asc" | "desc";

export interface TodosSortOption {
  value: `${TodosSort}-${TodosOrder}`;
  label: string;
}

/**
 * Single source of truth for the sort dropdown and filter-chip labels.
 * Every `TodosSort` the API accepts is represented — including
 * `created_at` and `status`, which the dropdown previously omitted.
 */
export const TODOS_SORT_OPTIONS: readonly TodosSortOption[] = [
  { value: "updated_at-desc", label: "Latest" },
  { value: "updated_at-asc", label: "Oldest" },
  { value: "created_at-desc", label: "Newest created" },
  { value: "created_at-asc", label: "Oldest created" },
  { value: "title-asc", label: "Title A-Z" },
  { value: "title-desc", label: "Title Z-A" },
  { value: "priority-desc", label: "High Priority" },
  { value: "status-asc", label: "Status A-Z" },
  { value: "status-desc", label: "Status Z-A" },
  { value: "due_date-asc", label: "Due Date" },
] as const;

export function getTodosSortLabel(sortValue: string): string {
  return (
    TODOS_SORT_OPTIONS.find((option) => option.value === sortValue)?.label ??
    sortValue
  );
}

const TABS: readonly TodosTab[] = [
  "all",
  "draft",
  "active",
  "completed",
  "archived",
] as const;
const PRIORITIES: readonly TodosPriorityFilter[] = [
  "all",
  "low",
  "medium",
  "high",
] as const;
const SORTS: readonly TodosSort[] = [
  "created_at",
  "updated_at",
  "title",
  "priority",
  "due_date",
  "status",
] as const;

const DEFAULT_TAB: TodosTab = "all";
const DEFAULT_PRIORITY: TodosPriorityFilter = "all";
const DEFAULT_SORT: TodosSort = "updated_at";
const DEFAULT_ORDER: TodosOrder = "desc";
const SEARCH_DEBOUNCE_MS = 300;

function parseTab(value: string | null): TodosTab {
  return value !== null &&
    (TABS as readonly string[]).includes(value as TodosTab)
    ? (value as TodosTab)
    : DEFAULT_TAB;
}

function parsePriority(value: string | null): TodosPriorityFilter {
  return value !== null &&
    (PRIORITIES as readonly string[]).includes(value as TodosPriorityFilter)
    ? (value as TodosPriorityFilter)
    : DEFAULT_PRIORITY;
}

function parseSort(value: string | null): TodosSort {
  return value !== null &&
    (SORTS as readonly string[]).includes(value as TodosSort)
    ? (value as TodosSort)
    : DEFAULT_SORT;
}

function parseOrder(value: string | null): TodosOrder {
  return value === "asc" || value === "desc" ? value : DEFAULT_ORDER;
}

function parseOverdue(value: string | null): boolean {
  return value === "true";
}

const DATE_RE = /^\d{4}-\d{2}-\d{2}$/;

function parseDueDate(value: string | null): string | null {
  if (value === null || !DATE_RE.test(value)) return null;
  const time = Date.parse(`${value}T00:00:00.000Z`);
  return Number.isNaN(time) ? null : value;
}

/**
 * Convert a `YYYY-MM-DD` URL value into an ISO datetime string for the
 * `dueFrom`/`dueTo` API params. `from` snaps to start-of-day UTC, `to`
 * snaps to end-of-day UTC so the range is inclusive of both dates.
 */
export function toDueBoundISO(
  date: string | null,
  bound: "from" | "to",
): string | undefined {
  if (date === null) return undefined;
  return bound === "from"
    ? `${date}T00:00:00.000Z`
    : `${date}T23:59:59.999Z`;
}

function parsePage(value: string | null): number {
  const parsed = Number.parseInt(value ?? "", 10);
  return Number.isInteger(parsed) && parsed >= 1 ? parsed : 1;
}

function parseTaskId(value: string | null): string | null {
  if (value === null) return null;
  const trimmed = value.trim();
  return trimmed === "" ? null : trimmed;
}

function isDefault(key: string, value: string): boolean {
  switch (key) {
    case "tab":
      return value === DEFAULT_TAB;
    case "category":
      return value === "all";
    case "priority":
      return value === DEFAULT_PRIORITY;
    case "sort":
      return value === DEFAULT_SORT;
    case "order":
      return value === DEFAULT_ORDER;
    case "page":
      return value === "1";
    case "q":
      return value === "";
    case "overdue":
      return value === "false";
    case "dueFrom":
    case "dueTo":
      return value === "";
    case "task":
    case "edit":
      // `task`/`edit` are selection state, not filters — their absence is
      // the default, so `setParam` deletes them when unset.
      return value === "";
    default:
      return false;
  }
}

/**
 * Todos filter state owned by the URL (useSearchParams) so filters are
 * shareable via link and survive refresh. Search input is local state
 * debounced into `q` (replace history to avoid per-keystroke entries).
 * Defaults are omitted from the URL to keep links clean.
 */
export function useTodosSearchParams() {
  const [searchParams, setSearchParams] = useSearchParams();

  const urlSearch = searchParams.get("q") ?? "";
  const [searchInput, setSearchInput] = useState(urlSearch);

  // Sync input when URL changes externally (back/forward, pasted link).
  useEffect(() => {
    setSearchInput(urlSearch);
  }, [urlSearch]);

  const debouncedSearch = useDebounce(searchInput, SEARCH_DEBOUNCE_MS);

  // Push debounced input into the URL, resetting to page 1.
  useEffect(() => {
    if (debouncedSearch === urlSearch) return;
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (debouncedSearch === "") next.delete("q");
        else next.set("q", debouncedSearch);
        next.delete("page");
        return next;
      },
      { replace: true },
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedSearch]);

  const tab = parseTab(searchParams.get("tab"));
  const category: string = searchParams.get("category") ?? "all";
  const priority = parsePriority(searchParams.get("priority"));
  const sort = parseSort(searchParams.get("sort"));
  const order = parseOrder(searchParams.get("order"));
  const page = parsePage(searchParams.get("page"));
  const overdueOnly = parseOverdue(searchParams.get("overdue"));
  const dueFrom = parseDueDate(searchParams.get("dueFrom"));
  const dueTo = parseDueDate(searchParams.get("dueTo"));
  // Route-less detail drawer selection. `?task=<id>` deep-links a task;
  // `?task=<id>&edit=1` additionally arms inline title editing on open.
  const selectedTaskId = parseTaskId(searchParams.get("task"));

  const setParam = useCallback(
    (patch: Record<string, string | undefined>, replace = false) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          for (const [key, value] of Object.entries(patch)) {
            if (value === undefined || isDefault(key, value)) next.delete(key);
            else next.set(key, value);
          }
          // Any filter change (other than an explicit page nav) resets to page 1.
          if (!("page" in patch)) next.delete("page");
          return next;
        },
        { replace },
      );
    },
    [setSearchParams],
  );

  const setTab = useCallback(
    (value: TodosTab) => setParam({ tab: value }),
    [setParam],
  );
  const setCategory = useCallback(
    (value: string) => setParam({ category: value }),
    [setParam],
  );
  const setPriority = useCallback(
    (value: TodosPriorityFilter) => setParam({ priority: value }),
    [setParam],
  );
  const setSortOrder = useCallback(
    (nextSort: TodosSort, nextOrder: TodosOrder) =>
      setParam({ sort: nextSort, order: nextOrder }),
    [setParam],
  );
  const setPage = useCallback(
    (nextPage: number) =>
      setParam({ page: String(Math.max(1, nextPage)) }, true),
    [setParam],
  );
  const setOverdueOnly = useCallback(
    (value: boolean) => setParam({ overdue: value ? "true" : undefined }),
    [setParam],
  );
  const setDueFrom = useCallback(
    (value: string | undefined) => setParam({ dueFrom: value || undefined }),
    [setParam],
  );
  const setDueTo = useCallback(
    (value: string | undefined) => setParam({ dueTo: value || undefined }),
    [setParam],
  );
  const setDueRange = useCallback(
    (from: string | undefined, to: string | undefined) =>
      setParam({ dueFrom: from || undefined, dueTo: to || undefined }),
    [setParam],
  );
  /**
   * Open (`id`) or close (`undefined`) the route-less detail drawer.
   * Unlike filter setters this never touches `page` — selecting a task
   * must not reset pagination. Opening pushes a history entry so the
   * browser Back button closes the drawer; closing replaces to avoid a
   * reopen-on-back loop. `edit: true` arms inline title editing on open.
   */
  const setSelectedTaskId = useCallback(
    (id: string | undefined, options?: { edit?: boolean }) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (id === undefined) {
            next.delete("task");
            next.delete("edit");
          } else {
            next.set("task", id);
            if (options?.edit) next.set("edit", "1");
            else next.delete("edit");
          }
          return next;
        },
        { replace: id === undefined },
      );
    },
    [setSearchParams],
  );
  const clearDueRange = useCallback(
    () => setParam({ dueFrom: undefined, dueTo: undefined }),
    [setParam],
  );
  const clearFilters = useCallback(() => {
    setSearchInput("");
    setSearchParams({}, { replace: true });
  }, [setSearchParams]);

  const hasActiveFilters =
    debouncedSearch !== "" ||
    tab !== DEFAULT_TAB ||
    category !== "all" ||
    priority !== DEFAULT_PRIORITY ||
    sort !== DEFAULT_SORT ||
    order !== DEFAULT_ORDER ||
    overdueOnly ||
    dueFrom !== null ||
    dueTo !== null;

  const activeFilterCount =
    (debouncedSearch !== "" ? 1 : 0) +
    (category !== "all" ? 1 : 0) +
    (priority !== DEFAULT_PRIORITY ? 1 : 0) +
    (sort !== DEFAULT_SORT || order !== DEFAULT_ORDER ? 1 : 0) +
    (overdueOnly ? 1 : 0) +
    (dueFrom !== null ? 1 : 0) +
    (dueTo !== null ? 1 : 0);

  return {
    // URL-derived filter values
    tab,
    category,
    priority,
    sort,
    order,
    page,
    sortValue: `${sort}-${order}`,
    overdueOnly,
    dueFrom,
    dueTo,
    // Route-less detail drawer selection (`?task=<id>`)
    selectedTaskId,
    // Search: immediate input + debounced URL-backed value
    searchInput,
    setSearchInput,
    debouncedSearch,
    // Setters
    setTab,
    setCategory,
    setPriority,
    setSortOrder,
    setPage,
    setOverdueOnly,
    setDueFrom,
    setDueTo,
    setDueRange,
    setSelectedTaskId,
    clearDueRange,
    clearFilters,
    hasActiveFilters,
    activeFilterCount,
  };
}
