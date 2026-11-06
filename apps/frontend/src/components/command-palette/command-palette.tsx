import { useGetAllTodos } from "@/api/hooks/use-todo-query";
import { ShortcutHint } from "@/components/command-palette/shortcut-hint";
import { useTheme } from "@/components/theme-provider";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from "@/components/ui/command";
import {
  displayKeys,
  openNewTaskDialog,
  readRecentPicks,
  recordRecentPick,
  resolveThemeToggle,
  type RecentPick,
} from "@/lib/shortcuts";
import {
  AlarmClock,
  CalendarClock,
  CheckCheck,
  CheckSquare,
  Flag,
  Folder,
  History,
  Keyboard,
  LayoutDashboard,
  Link2,
  ListTodo,
  Plus,
  Settings,
  SunMoon,
} from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onOpenShortcuts: () => void;
}

export function CommandPalette({
  open,
  onOpenChange,
  onOpenShortcuts,
}: CommandPaletteProps) {
  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Command palette"
      description="Jump to pages, run actions, or apply task filters."
      className="sm:max-w-xl"
    >
      {/* Mounted only while the dialog is open, so data hooks below
          fire exclusively on palette open. */}
      {open && (
        <PaletteBody
          onOpenChange={onOpenChange}
          onOpenShortcuts={onOpenShortcuts}
        />
      )}
    </CommandDialog>
  );
}

const GO_ITEMS = [
  { label: "Go to Dashboard", href: "/dashboard", Icon: LayoutDashboard },
  { label: "Go to Tasks", href: "/todos", Icon: CheckSquare },
  { label: "Go to Categories", href: "/categories", Icon: Folder },
  { label: "Go to Settings", href: "/settings", Icon: Settings },
];

const FILTER_VIEWS = [
  { label: "Active tasks", href: "/todos?tab=active", Icon: ListTodo },
  { label: "Completed tasks", href: "/todos?tab=completed", Icon: CheckCheck },
  { label: "Overdue tasks", href: "/todos?overdue=true", Icon: AlarmClock },
  { label: "High priority tasks", href: "/todos?priority=high", Icon: Flag },
  {
    label: "Due soon",
    href: "/todos?sort=due_date&order=asc",
    Icon: CalendarClock,
  },
];

function PaletteBody({
  onOpenChange,
  onOpenShortcuts,
}: Omit<CommandPaletteProps, "open">) {
  const navigate = useNavigate();
  const { setTheme } = useTheme();
  const [picks, setPicks] = useState<RecentPick[]>(() => readRecentPicks());
  const { data: recentTasks } = useGetAllTodos({
    query: { page: 1, limit: 5, sort: "updated_at", order: "desc" },
  });

  const go = (pick: RecentPick) => {
    setPicks(recordRecentPick(pick));
    onOpenChange(false);
    navigate(pick.href);
  };

  const newTask = () => {
    setPicks(
      recordRecentPick({ label: "New task", detail: "Action", href: "/todos" }),
    );
    onOpenChange(false);
    // The New Task trigger lives in the app topbar; activate it so the
    // real create form opens. Fall back to the Tasks page when absent.
    requestAnimationFrame(() => {
      if (!openNewTaskDialog()) navigate("/todos");
    });
  };

  const toggleTheme = () => {
    onOpenChange(false);
    setTheme(resolveThemeToggle());
  };

  const openShortcuts = () => {
    onOpenChange(false);
    onOpenShortcuts();
  };

  const copyLink = async () => {
    const url = window.location.href;
    onOpenChange(false);
    try {
      await navigator.clipboard.writeText(url);
      toast.success("Link copied to clipboard");
    } catch {
      toast.error("Could not copy link");
    }
  };

  const showTaskRecents =
    picks.length === 0 && (recentTasks?.data?.length ?? 0) > 0;

  return (
    <>
      <CommandInput placeholder="Search tasks, pages, actions…" />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>

        {picks.length > 0 && (
          <CommandGroup heading="Recent">
            {picks.map((pick) => (
              <CommandItem
                key={pick.href}
                value={`recent ${pick.label}`}
                onSelect={() => go(pick)}
              >
                <History />
                <span>{pick.label}</span>
                {pick.detail && (
                  <span className="text-xs text-muted-foreground">
                    {pick.detail}
                  </span>
                )}
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        <CommandGroup heading="Go to">
          {GO_ITEMS.map(({ label, href, Icon }) => (
            <CommandItem
              key={href}
              value={label}
              onSelect={() => go({ label, detail: "Page", href })}
            >
              <Icon />
              <span>{label.replace(/^Go to /, "")}</span>
            </CommandItem>
          ))}
        </CommandGroup>

        <CommandGroup heading="Actions">
          <CommandItem value="new task create" onSelect={newTask}>
            <Plus />
            <span>New task</span>
            <CommandShortcut>
              <ShortcutHint keys={displayKeys(["N"])} />
            </CommandShortcut>
          </CommandItem>
          <CommandItem
            value="toggle theme dark light mode"
            onSelect={toggleTheme}
          >
            <SunMoon />
            <span>Toggle theme</span>
            <CommandShortcut>
              <ShortcutHint keys={displayKeys(["T"])} />
            </CommandShortcut>
          </CommandItem>
          <CommandItem value="keyboard shortcuts help" onSelect={openShortcuts}>
            <Keyboard />
            <span>Keyboard shortcuts</span>
            <CommandShortcut>
              <ShortcutHint keys={displayKeys(["?"])} />
            </CommandShortcut>
          </CommandItem>
          <CommandItem value="copy link to this view" onSelect={copyLink}>
            <Link2 />
            <span>Copy link to this view</span>
          </CommandItem>
        </CommandGroup>

        <CommandGroup heading="Filter views">
          {FILTER_VIEWS.map(({ label, href, Icon }) => (
            <CommandItem
              key={href}
              value={label}
              onSelect={() => go({ label, detail: "Filter", href })}
            >
              <Icon />
              <span>{label}</span>
            </CommandItem>
          ))}
        </CommandGroup>

        {showTaskRecents && (
          <CommandGroup heading="Recent tasks">
            {(recentTasks?.data ?? []).map((todo) => (
              <CommandItem
                key={todo.id}
                value={`task ${todo.title}`}
                onSelect={() =>
                  go({
                    label: todo.title,
                    detail: "Task",
                    href: `/todos?q=${encodeURIComponent(todo.title)}`,
                  })
                }
              >
                <CheckSquare />
                <span className="truncate">{todo.title}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
      </CommandList>
      <div className="hidden items-center gap-4 border-t px-3 py-2 text-[11px] text-muted-foreground sm:flex">
        <span className="flex items-center gap-1.5">
          <ShortcutHint keys={["↑", "↓"]} /> navigate
        </span>
        <span className="flex items-center gap-1.5">
          <ShortcutHint keys={["↵"]} /> select
        </span>
        <span className="ml-auto flex items-center gap-1.5">
          <ShortcutHint keys={["esc"]} /> close
        </span>
      </div>
    </>
  );
}
