import {
  useCreateTodo,
  useDeleteTodo,
  useUpdateTodo,
  type TGetTodosResponse,
} from "@/api/hooks/use-todo-query";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import {
  ArrowUpFromLine,
  Check,
  LoaderCircle,
  MoreHorizontal,
  Pencil,
  Trash2,
  X,
} from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

export type TSubtask = TGetTodosResponse["data"][number]["children"][number];

interface SubtaskRowProps {
  subtask: TSubtask;
}

export function SubtaskRow({ subtask }: SubtaskRowProps) {
  const updateTodo = useUpdateTodo();
  const deleteTodo = useDeleteTodo();
  const createTodo = useCreateTodo();
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(subtask.title);
  const skipBlurCommit = useRef(false);
  const isDone = subtask.status === "completed";
  const busy =
    updateTodo.isPending || deleteTodo.isPending || createTodo.isPending;

  const startRename = () => {
    skipBlurCommit.current = false;
    setDraft(subtask.title);
    setEditing(true);
  };

  const cancelRename = () => {
    skipBlurCommit.current = true;
    setDraft(subtask.title);
    setEditing(false);
  };

  const handleToggle = async () => {
    try {
      await updateTodo.mutateAsync({
        todoId: subtask.id,
        body: { status: isDone ? "active" : "completed" },
      });
    } catch {
      toast.error("Failed to update subtask");
    }
  };

  const commitRename = async () => {
    const next = draft.trim();
    if (!next || next === subtask.title) {
      setEditing(false);
      setDraft(subtask.title);
      return;
    }
    try {
      await updateTodo.mutateAsync({
        todoId: subtask.id,
        body: { title: next },
      });
      setEditing(false);
    } catch {
      toast.error("Failed to rename subtask");
    }
  };

  const handleRemove = async () => {
    try {
      await deleteTodo.mutateAsync({ todoId: subtask.id });
      toast.success("Subtask removed");
    } catch {
      toast.error("Failed to remove subtask");
    }
  };

  // The update endpoint treats an explicit null parent the same as an
  // omitted one, so detaching isn't possible via PATCH. Converting copies
  // the scalar fields onto a new top-level task, then removes the subtask.
  const handleConvert = async () => {
    try {
      const created = await createTodo.mutateAsync({
        body: {
          title: subtask.title,
          description: subtask.description ?? undefined,
          priority: subtask.priority,
          dueDate: subtask.dueDate ?? undefined,
          categoryId: subtask.categoryId ?? undefined,
        },
      });
      if (created.status !== subtask.status) {
        await updateTodo.mutateAsync({
          todoId: created.id,
          body: { status: subtask.status },
        });
      }
      await deleteTodo.mutateAsync({ todoId: subtask.id });
      toast.success("Converted to standalone task");
    } catch {
      toast.error("Failed to convert subtask");
    }
  };

  return (
    <li className="group flex items-center gap-2 rounded-md px-2 py-1.5 hover:bg-muted/60">
      <Checkbox
        checked={isDone}
        onCheckedChange={handleToggle}
        disabled={busy}
        aria-label={`Toggle subtask ${subtask.title}`}
        className="h-3.5 w-3.5"
      />
      {editing ? (
        <>
          <Input
            autoFocus
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") void commitRename();
              if (e.key === "Escape") cancelRename();
            }}
            onBlur={() => {
              if (skipBlurCommit.current) return;
              void commitRename();
            }}
            disabled={updateTodo.isPending}
            aria-label="Subtask title"
            className="h-7 flex-1 text-sm"
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-6 w-6 shrink-0"
            onClick={() => void commitRename()}
            disabled={updateTodo.isPending}
            aria-label="Save subtask title"
          >
            <Check className="h-3.5 w-3.5" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-6 w-6 shrink-0"
            onClick={cancelRename}
            aria-label="Cancel rename"
          >
            <X className="h-3.5 w-3.5" />
          </Button>
        </>
      ) : (
        <>
          <button
            type="button"
            onDoubleClick={startRename}
            title="Double-click to rename"
            aria-label={`Rename subtask ${subtask.title}`}
            className={cn(
              "flex-1 truncate rounded-sm text-left text-sm focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary",
              isDone && "line-through text-muted-foreground",
            )}
          >
            {subtask.title}
          </button>
          {subtask.priority && (
            <span
              className={cn(
                "h-1.5 w-1.5 shrink-0 rounded-full",
                subtask.priority === "high" && "bg-red-500",
                subtask.priority === "medium" && "bg-yellow-500",
                subtask.priority === "low" && "bg-green-500",
              )}
              title={`${subtask.priority} priority`}
            />
          )}
          <span className="shrink-0 text-[11px] capitalize text-muted-foreground">
            {subtask.status}
          </span>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="h-6 w-6 shrink-0 opacity-0 group-hover:opacity-100 focus-visible:opacity-100"
                disabled={busy}
                aria-label={`Subtask actions for ${subtask.title}`}
              >
                <MoreHorizontal className="h-3.5 w-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onSelect={startRename}>
                <Pencil className="mr-2 h-4 w-4" />
                Rename
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={() => void handleConvert()}>
                <ArrowUpFromLine className="mr-2 h-4 w-4" />
                Convert to standalone task
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onSelect={() => void handleRemove()}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Remove
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </>
      )}
    </li>
  );
}

/** Optimistic placeholder shown while a subtask create is in flight. */
export function SubtaskGhostRow({ title }: { title: string }) {
  return (
    <li
      aria-hidden
      className="flex items-center gap-2 rounded-md px-2 py-1.5 text-muted-foreground"
    >
      <LoaderCircle className="h-3.5 w-3.5 shrink-0 motion-safe:animate-spin" />
      <span className="flex-1 truncate text-sm">{title}</span>
      <span className="shrink-0 text-[11px]">Adding…</span>
    </li>
  );
}
