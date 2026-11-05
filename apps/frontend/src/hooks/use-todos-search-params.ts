import { useDebounce } from "@/api/hooks/use-debounce";
import type { TGetTodosQuery } from "@/api/hooks/use-todo-query";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

export type TodosTab = "draft" | "active" | "completed" | "archived" | "all";
export type TodosPriorityFilter = "low" | "medium" | "high" | "all";
export type TodosSort = NonNullable<TGetTodosQuery["sort"]>;
export type TodosOrder = "asc" | "desc";

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

function parsePage(value: string | null): number {
  const parsed = Number.parseInt(value ?? "", 10);
  return Number.isInteger(parsed) && parsed >= 1 ? parsed : 1;
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
    order !== DEFAULT_ORDER;

  return {
    // URL-derived filter values
    tab,
    category,
    priority,
    sort,
    order,
    page,
    sortValue: `${sort}-${order}`,
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
    clearFilters,
    hasActiveFilters,
  };
}
