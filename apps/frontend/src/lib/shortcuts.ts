import { useEffect, useRef } from "react";

/**
 * Central registry + helpers for global keyboard shortcuts.
 *
 * Key labels use `"mod"` as a placeholder for the platform modifier
 * (⌘ on macOS/iOS, Ctrl elsewhere) — resolve with {@link displayKeys}
 * before rendering.
 */

export interface ShortcutDef {
  id: "palette" | "search" | "new-task" | "toggle-theme" | "shortcuts-help";
  /** Short label used in toasts / aria text. */
  label: string;
  /** One-line description shown in the shortcuts dialog. */
  description: string;
  /** Display labels, e.g. `["mod", "K"]`. */
  keys: string[];
}

export const SHORTCUT_DEFS: readonly ShortcutDef[] = [
  {
    id: "palette",
    label: "Command palette",
    description: "Open the command palette",
    keys: ["mod", "K"],
  },
  {
    id: "search",
    label: "Search",
    description: "Open the palette with search focused",
    keys: ["/"],
  },
  {
    id: "new-task",
    label: "New task",
    description: "Open the New Task dialog",
    keys: ["N"],
  },
  {
    id: "toggle-theme",
    label: "Toggle theme",
    description: "Switch between light and dark mode",
    keys: ["T"],
  },
  {
    id: "shortcuts-help",
    label: "Shortcuts",
    description: "Show all keyboard shortcuts",
    keys: ["?"],
  },
];

export function isMacPlatform(): boolean {
  if (typeof navigator === "undefined") return false;
  const nav = navigator as Navigator & {
    userAgentData?: { platform?: string };
  };
  const platform = nav.userAgentData?.platform ?? nav.platform ?? "";
  return /mac|iphone|ipad|ipod/i.test(platform);
}

/** Resolve `"mod"` placeholders to the platform modifier label. */
export function displayKeys(keys: readonly string[]): string[] {
  const mod = isMacPlatform() ? "⌘" : "Ctrl";
  return keys.map((key) => (key === "mod" ? mod : key));
}

/**
 * True when the key event originated from an editable element, in which
 * case single-key shortcuts must not fire (the user is typing).
 */
export function isEditableTarget(target: EventTarget | null): boolean {
  if (typeof HTMLElement === "undefined") return false;
  if (!(target instanceof HTMLElement)) return false;
  if (target.isContentEditable) return true;
  const tag = target.tagName;
  return tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";
}

/** Selector for the topbar New Task trigger (a TodoCreateForm trigger). */
export const NEW_TASK_TRIGGER_SELECTOR = "[data-new-task-trigger]";

/**
 * Open the New Task dialog by activating its trigger button.
 * Returns false when the trigger is not mounted.
 */
export function openNewTaskDialog(): boolean {
  if (typeof document === "undefined") return false;
  const trigger = document.querySelector<HTMLElement>(
    NEW_TASK_TRIGGER_SELECTOR,
  );
  if (!trigger) return false;
  trigger.click();
  return true;
}

/** Resolve the opposite of the currently applied color scheme. */
export function resolveThemeToggle(): "light" | "dark" {
  if (typeof document === "undefined") return "light";
  return document.documentElement.classList.contains("dark") ? "light" : "dark";
}

// --- Recent picks (palette "Recent" group) --------------------------------

export interface RecentPick {
  label: string;
  detail?: string;
  href: string;
}

interface StoredPick extends RecentPick {
  at: number;
}

const RECENT_KEY = "taskify:palette-recent";
const RECENT_MAX = 5;

export function readRecentPicks(): RecentPick[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed
      .filter(
        (entry): entry is StoredPick =>
          typeof entry === "object" &&
          entry !== null &&
          typeof (entry as StoredPick).label === "string" &&
          typeof (entry as StoredPick).href === "string",
      )
      .slice(0, RECENT_MAX);
  } catch {
    return [];
  }
}

/** Prepend a pick (dedupe by href), persist, and return the new list. */
export function recordRecentPick(pick: RecentPick): RecentPick[] {
  const next: StoredPick[] = [
    { ...pick, at: Date.now() },
    ...readRecentPicks()
      .filter((entry) => entry.href !== pick.href)
      .map((entry) => ({ ...entry, at: 0 })),
  ].slice(0, RECENT_MAX);
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(next));
  } catch {
    // Storage unavailable (private mode) — recent picks just don't persist.
  }
  return next;
}

// --- Global hotkeys ---------------------------------------------------------

export interface PaletteHotkeyCallbacks {
  onTogglePalette: () => void;
  onOpenPalette: () => void;
  onOpenShortcuts: () => void;
  onNewTask: () => void;
  onToggleTheme: () => void;
}

/**
 * Global keybindings for the palette. Single-key shortcuts are ignored
 * while typing; ⌘K/Ctrl+K works everywhere (it never produces text).
 */
export function usePaletteHotkeys(callbacks: PaletteHotkeyCallbacks) {
  const ref = useRef(callbacks);
  ref.current = callbacks;

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        ref.current.onTogglePalette();
        return;
      }
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      if (isEditableTarget(event.target)) return;

      switch (event.key) {
        case "/":
          event.preventDefault();
          ref.current.onOpenPalette();
          break;
        case "?":
          ref.current.onOpenShortcuts();
          break;
        case "n":
        case "N":
          ref.current.onNewTask();
          break;
        case "t":
        case "T":
          ref.current.onToggleTheme();
          break;
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);
}
