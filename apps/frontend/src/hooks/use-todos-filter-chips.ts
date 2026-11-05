import {
  getTodosSortLabel,
  type TodosPriorityFilter,
} from "@/hooks/use-todos-search-params";
import { useMemo } from "react";

export interface FilterChip {
  key: string;
  label: string;
  onRemove: () => void;
}

interface UseTodosFilterChipsInput {
  search: string;
  category: string;
  priority: TodosPriorityFilter;
  sortValue: string;
  isDefaultSort: boolean;
  overdueOnly: boolean;
  dueFrom: string | null;
  dueTo: string | null;
  categories: readonly { id: string; name: string }[] | undefined;
  onClearSearch: () => void;
  onClearCategory: () => void;
  onClearPriority: () => void;
  onClearSort: () => void;
  onClearOverdue: () => void;
  onClearDueFrom: () => void;
  onClearDueTo: () => void;
}

const PRIORITY_LABELS: Record<Exclude<TodosPriorityFilter, "all">, string> = {
  low: "Low",
  medium: "Medium",
  high: "High",
};

/**
 * Derives the active-filter chips for the todos filter bar. Each chip
 * carries its own remover so every chip renders an independent X —
 * clearing one filter never touches the others.
 */
export function useTodosFilterChips({
  search,
  category,
  priority,
  sortValue,
  isDefaultSort,
  overdueOnly,
  dueFrom,
  dueTo,
  categories,
  onClearSearch,
  onClearCategory,
  onClearPriority,
  onClearSort,
  onClearOverdue,
  onClearDueFrom,
  onClearDueTo,
}: UseTodosFilterChipsInput): FilterChip[] {
  return useMemo(() => {
    const chips: FilterChip[] = [];

    if (search !== "") {
      chips.push({
        key: "q",
        label: `Search: ${search}`,
        onRemove: onClearSearch,
      });
    }

    if (category !== "all") {
      const name =
        categories?.find((item) => item.id === category)?.name ?? "Category";
      chips.push({
        key: "category",
        label: `Category: ${name}`,
        onRemove: onClearCategory,
      });
    }

    if (priority !== "all") {
      chips.push({
        key: "priority",
        label: `Priority: ${PRIORITY_LABELS[priority]}`,
        onRemove: onClearPriority,
      });
    }

    if (!isDefaultSort) {
      chips.push({
        key: "sort",
        label: `Sort: ${getTodosSortLabel(sortValue)}`,
        onRemove: onClearSort,
      });
    }

    if (overdueOnly) {
      chips.push({
        key: "overdue",
        label: "Overdue",
        onRemove: onClearOverdue,
      });
    }

    if (dueFrom !== null) {
      chips.push({
        key: "dueFrom",
        label: `Due from ${dueFrom}`,
        onRemove: onClearDueFrom,
      });
    }

    if (dueTo !== null) {
      chips.push({
        key: "dueTo",
        label: `Due to ${dueTo}`,
        onRemove: onClearDueTo,
      });
    }

    return chips;
  }, [
    search,
    category,
    categories,
    priority,
    sortValue,
    isDefaultSort,
    overdueOnly,
    dueFrom,
    dueTo,
    onClearSearch,
    onClearCategory,
    onClearPriority,
    onClearSort,
    onClearOverdue,
    onClearDueFrom,
    onClearDueTo,
  ]);
}
