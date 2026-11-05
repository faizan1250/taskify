import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

export interface SavedView {
  id: string;
  name: string;
  createdAt: string;
  /** URL param snapshot (tab, category, priority, sort, order, q, overdue, dueFrom, dueTo). Never includes `page`. */
  params: Record<string, string>;
}

const STORAGE_KEY = "taskify:todos-saved-views:v1";
const MAX_VIEWS = 12;

function readStoredViews(): SavedView[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(
      (entry): entry is SavedView =>
        typeof entry === "object" &&
        entry !== null &&
        typeof (entry as SavedView).id === "string" &&
        typeof (entry as SavedView).name === "string" &&
        typeof (entry as SavedView).params === "object" &&
        (entry as SavedView).params !== null,
    );
  } catch {
    return [];
  }
}

function newId(): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return crypto.randomUUID();
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

/**
 * Named filter presets persisted in localStorage. Applying a view writes
 * its param snapshot to the URL (resetting to page 1); the todos search
 * input re-syncs from `q` automatically via `useTodosSearchParams`.
 */
export function useSavedViews() {
  const [, setSearchParams] = useSearchParams();
  const [views, setViews] = useState<SavedView[]>(readStoredViews);

  useEffect(() => {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(views));
    } catch {
      // Storage full or unavailable (private mode) — views stay in memory.
    }
  }, [views]);

  const saveView = useCallback(
    (name: string, params: Record<string, string>) => {
      const trimmed = name.trim();
      if (trimmed === "") return;
      setViews((prev) => {
        const next: SavedView[] = [
          {
            id: newId(),
            name: trimmed,
            createdAt: new Date().toISOString(),
            params,
          },
          ...prev,
        ];
        return next.slice(0, MAX_VIEWS);
      });
    },
    [],
  );

  const applyView = useCallback(
    (id: string) => {
      const view = views.find((entry) => entry.id === id);
      if (!view) return;
      // Replace the whole param set so stale filters from the current
      // URL can't leak into the applied view. `page` is intentionally
      // absent, which resets pagination to 1.
      setSearchParams(view.params);
    },
    [views, setSearchParams],
  );

  const deleteView = useCallback((id: string) => {
    setViews((prev) => prev.filter((entry) => entry.id !== id));
  }, []);

  return { views, saveView, applyView, deleteView };
}
