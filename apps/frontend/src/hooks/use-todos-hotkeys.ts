import { isEditableTarget } from "@/lib/shortcuts";
import { useEffect, useRef, useState } from "react";

export interface TodosHotkeysCallbacks {
  /** Open the New Task dialog (C). */
  onOpenCreate: () => void;
  /** Move DOM focus to the list search input (/). */
  onFocusSearch: () => void;
  /** Toggle selection of the todo at the given visible index (X). */
  onToggleSelect: (index: number) => void;
  /** Open the edit dialog for the active todo (E). */
  onEditActive: () => void;
  /** Clear selection and reset the active ring (Esc). */
  onClear: () => void;
}

interface UseTodosHotkeysOptions extends TodosHotkeysCallbacks {
  /** False while the list query is loading — no keys fire. */
  enabled: boolean;
  /** Number of currently visible todos; the active index clamps to it. */
  count: number;
}

/**
 * Keyboard-first navigation for the task list.
 *
 * - `C` opens the New Task dialog, `/` focuses search, `J`/`K` move a
 *   roving active ring, `X` toggles selection of the active task,
 *   `E` opens its edit dialog, `Esc` clears selection.
 * - Single-key bindings never fire while typing, with modifiers held,
 *   or while any dialog is open (Radix owns Esc/focus there).
 * - `/` is captured (capture phase + stopPropagation) so the global
 *   palette shortcut for `/` does not also fire on this page.
 */
export function useTodosHotkeys(options: UseTodosHotkeysOptions) {
  const { enabled, count, ...callbacks } = options;
  const callbacksRef = useRef(callbacks);
  callbacksRef.current = callbacks;

  const [activeIndex, setActiveIndexState] = useState(0);
  const activeRef = useRef(0);

  const setActiveIndex = (index: number) => {
    activeRef.current = index;
    setActiveIndexState(index);
  };

  // Clamp the active index when the visible list shrinks.
  useEffect(() => {
    const clamped = count === 0 ? 0 : Math.min(activeRef.current, count - 1);
    if (clamped !== activeRef.current) {
      activeRef.current = clamped;
      setActiveIndexState(clamped);
    }
  }, [count]);

  const activeCountRef = useRef(count);
  activeCountRef.current = count;

  useEffect(() => {
    if (!enabled) return;

    const move = (delta: 1 | -1) => {
      const total = activeCountRef.current;
      if (total === 0) return;
      const next = Math.min(
        Math.max(activeRef.current + delta, 0),
        total - 1,
      );
      activeRef.current = next;
      setActiveIndexState(next);
    };

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      // A modal owns the keyboard while open (palette, create/edit, help).
      if (document.querySelector('[role="dialog"]')) return;

      if (event.key === "Escape") {
        if (
          isEditableTarget(event.target) &&
          event.target instanceof HTMLElement
        ) {
          event.target.blur();
        }
        callbacksRef.current.onClear();
        return;
      }
      if (isEditableTarget(event.target)) return;

      switch (event.key) {
        case "/":
          event.preventDefault();
          // Capture-phase listener: stop the global `/` (palette) handler.
          event.stopPropagation();
          callbacksRef.current.onFocusSearch();
          break;
        case "c":
        case "C":
          callbacksRef.current.onOpenCreate();
          break;
        case "j":
        case "J":
          move(1);
          break;
        case "k":
        case "K":
          move(-1);
          break;
        case "x":
        case "X":
          callbacksRef.current.onToggleSelect(activeRef.current);
          break;
        case "e":
        case "E":
          callbacksRef.current.onEditActive();
          break;
      }
    };

    window.addEventListener("keydown", onKeyDown, true);
    return () => window.removeEventListener("keydown", onKeyDown, true);
  }, [enabled]);

  return { activeIndex, setActiveIndex };
}
