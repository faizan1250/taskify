import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useQueryClient, type QueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/api/query-utils";
import {
  useCreateTodo,
  useGetTodoById,
  useUpdateTodo,
  type TGetTodoByIdResponse,
  type TGetTodosResponse,
} from "@/api/hooks/use-todo-query";
import { useUploadTodoAttachment } from "@/api/hooks/use-todo-attachment-query";
import { isEditableTarget } from "@/lib/shortcuts";
import { TodoCommentsPanel } from "./todo-comments-panel";
import { TodoAttachments } from "./todo-attachments";
import { SubtaskGhostRow, SubtaskRow } from "./subtask-row";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetTitle,
} from "@/components/ui/sheet";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { cn, formatFileSize } from "@/lib/utils";
import {
  Archive,
  Calendar,
  CheckCircle2,
  Circle,
  Clock,
  FileText,
  FileWarning,
  ListTree,
  Loader2,
  MessageSquare,
  Paperclip,
  Plus,
  RotateCcw,
  Upload,
  X,
} from "lucide-react";
import { toast } from "sonner";

type DrawerTab = "details" | "subtasks" | "comments" | "attachments";

const statusMeta = {
  draft: { icon: Circle, color: "text-gray-500", dot: "bg-gray-500", label: "Draft" },
  active: { icon: Clock, color: "text-blue-500", dot: "bg-blue-500", label: "Active" },
  completed: { icon: CheckCircle2, color: "text-green-500", dot: "bg-green-500", label: "Completed" },
  archived: { icon: Archive, color: "text-gray-500", dot: "bg-gray-400", label: "Archived" },
} as const;

const priorityMeta = {
  low: { variant: "success", dot: "bg-green-500", label: "Low" },
  medium: { variant: "warning", dot: "bg-amber-500", label: "Medium" },
  high: { variant: "danger", dot: "bg-red-500", label: "High" },
} as const;

const MAX_ATTACHMENT_BYTES = 10 * 1024 * 1024;

/**
 * Best-effort instant paint: when the drawer opens from a visible card,
 * the list cache already holds the full populated todo (same
 * `ZPopulatedTodo` contract as `getTodoById`), so seed it while the
 * detail query revalidates.
 */
function findCachedTodo(
  queryClient: QueryClient,
  id: string,
): TGetTodoByIdResponse | undefined {
  const lists = queryClient.getQueriesData<TGetTodosResponse>({
    queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
  });
  for (const [, response] of lists) {
    const found = response?.data.find((todo) => todo.id === id);
    if (found) return found;
  }
  return queryClient.getQueryData<TGetTodoByIdResponse>([
    QUERY_KEYS.TODOS.GET_TODO_BY_ID,
    id,
  ]);
}

function toDateInputValue(iso: string | null | undefined): string {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

/** Parse a `YYYY-MM-DD` date input as a local date, like the edit dialog. */
function fromDateInputValue(value: string): string {
  const [year, month, day] = value.split("-").map(Number);
  return new Date(year, month - 1, day).toISOString();
}

/**
 * Route-less task detail drawer. Selection lives in the URL (`?task=<id>`)
 * so details are deep-linkable and survive refresh; no route is added.
 * Mount once per page that shows task cards (tasks list, dashboard).
 */
export function TodoDetailDrawer() {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTask = searchParams.get("task");
  const taskId = rawTask !== null && rawTask.trim() !== "" ? rawTask.trim() : null;
  const editArmed = searchParams.get("edit") === "1";

  const queryClient = useQueryClient();
  const cachedTodo =
    taskId !== null ? findCachedTodo(queryClient, taskId) : undefined;
  const detailQuery = useGetTodoById({
    id: taskId ?? "",
    enabled: taskId !== null,
  });

  const todo = detailQuery.data ?? cachedTodo ?? null;
  const isLoading = taskId !== null && todo === null && detailQuery.isLoading;
  const isNotFound =
    taskId !== null && todo === null && !detailQuery.isLoading;

  const close = () => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete("task");
        next.delete("edit");
        return next;
      },
      { replace: true },
    );
  };

  return (
    <Sheet
      open={taskId !== null}
      onOpenChange={(open) => {
        if (!open) close();
      }}
    >
      <SheetContent
        side="right"
        className="flex w-full flex-col gap-0 p-0 sm:max-w-md"
        aria-label={todo ? `Details for ${todo.title}` : "Task details"}
      >
        {/* Radix requires a title; the visible title is the inline-edit field. */}
        <SheetTitle className="sr-only">Task details</SheetTitle>
        <SheetDescription className="sr-only">
          View and edit task details, subtasks, comments, and attachments.
        </SheetDescription>
        {taskId === null ? null : isLoading ? (
          <DrawerSkeleton />
        ) : todo === null || isNotFound ? (
          <DrawerNotFound taskId={taskId} />
        ) : (
          <DrawerBody
            key={taskId}
            todo={todo}
            editArmed={editArmed}
            onConsumeEditArm={() => {
              if (editArmed) {
                setSearchParams(
                  (prev) => {
                    const next = new URLSearchParams(prev);
                    next.delete("edit");
                    return next;
                  },
                  { replace: true },
                );
              }
            }}
          />
        )}
      </SheetContent>
    </Sheet>
  );
}

function DrawerSkeleton() {
  return (
    <div className="space-y-4 p-4 pt-12">
      <div className="flex items-center gap-3">
        <Skeleton className="h-5 w-5 rounded" />
        <Skeleton className="h-6 flex-1" />
      </div>
      <div className="flex gap-2">
        <Skeleton className="h-6 w-20" />
        <Skeleton className="h-6 w-24" />
      </div>
      <Skeleton className="h-9 w-full" />
      <Skeleton className="h-32 w-full" />
      <Skeleton className="h-24 w-full" />
    </div>
  );
}

function DrawerNotFound({ taskId }: { taskId: string }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 p-8 text-center">
      <FileWarning className="h-12 w-12 text-muted-foreground opacity-50" />
      <h2 className="text-lg font-medium">Task not found</h2>
      <p className="text-sm text-muted-foreground">
        No task with id <span className="font-mono text-xs">{taskId}</span>{" "}
        exists. It may have been deleted, or the link is stale.
      </p>
    </div>
  );
}

interface DrawerBodyProps {
  todo: TGetTodoByIdResponse;
  editArmed: boolean;
  onConsumeEditArm: () => void;
}

function DrawerBody({ todo, editArmed, onConsumeEditArm }: DrawerBodyProps) {
  const [tab, setTab] = useState<DrawerTab>("details");
  const updateTodo = useUpdateTodo();
  const titleRef = useRef<HTMLInputElement>(null);

  const StatusIcon = statusMeta[todo.status].icon;
  const subtasks = todo.children ?? [];
  const completedSubtasks = subtasks.filter(
    (child) => child.status === "completed",
  ).length;
  const commentCount = todo.comments?.length ?? 0;
  const attachmentCount = todo.attachments?.length ?? 0;

  // `?task=<id>&edit=1` (the `E` hotkey path): focus the title once.
  useEffect(() => {
    if (!editArmed) return;
    titleRef.current?.focus();
    titleRef.current?.select();
    onConsumeEditArm();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // `E` focuses inline title editing while the drawer owns the keyboard.
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      if (event.key !== "e" && event.key !== "E") return;
      if (isEditableTarget(event.target)) return;
      // A nested modal (e.g. delete confirm) owns the keyboard instead.
      if (
        document.querySelector('[role="alertdialog"], [role="menu"]')
      )
        return;
      event.preventDefault();
      titleRef.current?.focus();
      titleRef.current?.select();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  const handleStatusToggle = () => {
    const newStatus = todo.status === "completed" ? "active" : "completed";
    updateTodo.mutate({
      todoId: todo.id,
      body: { status: newStatus },
    });
  };

  const isOverdue =
    todo.dueDate &&
    new Date(todo.dueDate) < new Date() &&
    todo.status !== "completed";

  return (
    <>
      {/* Header: status toggle + badges (title lives in the Details tab as
          the inline-edit field so it stays editable, not decorative). */}
      <div className="space-y-3 border-b p-4 pr-12">
        <div className="flex items-start gap-3">
          <Checkbox
            checked={todo.status === "completed"}
            onCheckedChange={handleStatusToggle}
            disabled={updateTodo.isPending}
            className="mt-1"
            aria-label={
              todo.status === "completed"
                ? `Mark ${todo.title} as active`
                : `Mark ${todo.title} as completed`
            }
          />
          <div className="min-w-0 flex-1">
            <p className="flex items-center gap-1.5 text-sm font-medium">
              <StatusIcon
                className={cn("h-4 w-4", statusMeta[todo.status].color)}
              />
              {statusMeta[todo.status].label}
            </p>
            <p className="truncate text-xs text-muted-foreground">
              Updated {new Date(todo.updatedAt).toLocaleString()}
            </p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={priorityMeta[todo.priority].variant}>
            {priorityMeta[todo.priority].label}
          </Badge>
          {todo.category && (
            <Badge variant="secondary" className="flex items-center gap-1">
              <div
                className="w-2 h-2 rounded-full"
                style={{ backgroundColor: todo.category.color }}
              />
              {todo.category.name}
            </Badge>
          )}
          {todo.dueDate && (
            <Badge
              variant={isOverdue ? "danger" : "outline"}
              className="flex items-center gap-1"
            >
              <Calendar className="h-3 w-3" />
              {new Date(todo.dueDate).toLocaleDateString()}
            </Badge>
          )}
        </div>
      </div>

      <Tabs
        value={tab}
        onValueChange={(value) => setTab(value as DrawerTab)}
        className="flex min-h-0 flex-1 flex-col gap-0"
      >
        <div className="border-b px-4 pt-3">
          <TabsList aria-label="Task detail sections">
            <TabsTrigger value="details">Details</TabsTrigger>
            <TabsTrigger value="subtasks">
              Subtasks
              {subtasks.length > 0 && (
                <span className="text-xs text-muted-foreground">
                  {completedSubtasks}/{subtasks.length}
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger value="comments">
              Comments
              {commentCount > 0 && (
                <span className="text-xs text-muted-foreground">
                  {commentCount}
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger value="attachments">
              <Paperclip className="h-3.5 w-3.5" />
              Files
              {attachmentCount > 0 && (
                <span className="text-xs text-muted-foreground">
                  {attachmentCount}
                </span>
              )}
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent
          value="details"
          className="min-h-0 flex-1 overflow-y-auto"
        >
          <ScrollArea className="h-full">
            <DetailsTab todo={todo} titleRef={titleRef} />
          </ScrollArea>
        </TabsContent>

        <TabsContent
          value="subtasks"
          className="min-h-0 flex-1 overflow-y-auto"
        >
          <ScrollArea className="h-full">
            <SubtasksTab todo={todo} />
          </ScrollArea>
        </TabsContent>

        <TabsContent
          value="comments"
          className="min-h-0 flex-1 flex-col p-4 data-[state=active]:flex"
        >
          <TodoCommentsPanel todoId={todo.id} />
        </TabsContent>

        <TabsContent
          value="attachments"
          className="min-h-0 flex-1 overflow-y-auto"
        >
          <ScrollArea className="h-full">
            <AttachmentsTab todo={todo} />
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </>
  );
}

/* ------------------------------- Details tab ------------------------------ */

/**
 * Inline editing: every field auto-saves on blur/change through the
 * optimistic `useUpdateTodo` mutation (silent success, toast only on
 * error via the shared API error handler). No save button — the drawer
 * is the edit surface.
 */
function DetailsTab({
  todo,
  titleRef,
}: {
  todo: TGetTodoByIdResponse;
  titleRef: React.RefObject<HTMLInputElement | null>;
}) {
  const updateTodo = useUpdateTodo();
  const [title, setTitle] = useState(todo.title);
  const [description, setDescription] = useState(todo.description ?? "");
  const saving = updateTodo.isPending;

  const saveTitle = () => {
    const next = title.trim();
    if (next === "" || next === todo.title) {
      setTitle(todo.title);
      return;
    }
    updateTodo.mutate(
      { todoId: todo.id, body: { title: next } },
      { onError: () => setTitle(todo.title) },
    );
  };

  const saveDescription = () => {
    const next = description;
    if (next === (todo.description ?? "")) return;
    updateTodo.mutate(
      { todoId: todo.id, body: { description: next || undefined } },
      { onError: () => setDescription(todo.description ?? "") },
    );
  };

  return (
    <div className="space-y-4 p-4">
      <div className="space-y-1.5">
        <Label htmlFor="drawer-title">
          Title{" "}
          <span className="font-normal text-muted-foreground">
            — press{" "}
            <kbd className="rounded border bg-muted px-1 font-sans text-[11px]">
              E
            </kbd>{" "}
            to edit
          </span>
        </Label>
        <Input
          id="drawer-title"
          ref={titleRef}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          onBlur={saveTitle}
          onKeyDown={(e) => {
            if (e.key === "Enter") (e.target as HTMLInputElement).blur();
            if (e.key === "Escape") {
              setTitle(todo.title);
              (e.target as HTMLInputElement).blur();
            }
          }}
          disabled={saving}
          maxLength={255}
          className="text-base font-semibold"
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="drawer-description">Description</Label>
        <Textarea
          id="drawer-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          onBlur={saveDescription}
          onKeyDown={(e) => {
            if (e.key === "Escape") {
              setDescription(todo.description ?? "");
              (e.target as HTMLTextAreaElement).blur();
            }
          }}
          disabled={saving}
          rows={4}
          placeholder="Add a description…"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <Label htmlFor="drawer-status">Status</Label>
          <Select
            value={todo.status}
            onValueChange={(value: TGetTodoByIdResponse["status"]) =>
              updateTodo.mutate({ todoId: todo.id, body: { status: value } })
            }
            disabled={saving}
          >
            <SelectTrigger id="drawer-status">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(
                Object.entries(statusMeta) as [
                  TGetTodoByIdResponse["status"],
                  (typeof statusMeta)[TGetTodoByIdResponse["status"]],
                ][]
              ).map(([value, meta]) => (
                <SelectItem key={value} value={value}>
                  <div className="flex items-center gap-2">
                    <div className={cn("w-2 h-2 rounded-full", meta.dot)} />
                    {meta.label}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="drawer-priority">Priority</Label>
          <Select
            value={todo.priority}
            onValueChange={(value: TGetTodoByIdResponse["priority"]) =>
              updateTodo.mutate({ todoId: todo.id, body: { priority: value } })
            }
            disabled={saving}
          >
            <SelectTrigger id="drawer-priority">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(
                Object.entries(priorityMeta) as [
                  TGetTodoByIdResponse["priority"],
                  (typeof priorityMeta)[TGetTodoByIdResponse["priority"]],
                ][]
              ).map(([value, meta]) => (
                <SelectItem key={value} value={value}>
                  <div className="flex items-center gap-2">
                    <div className={cn("w-2 h-2 rounded-full", meta.dot)} />
                    {meta.label}
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="drawer-due-date">Due date</Label>
        <Input
          id="drawer-due-date"
          type="date"
          defaultValue={toDateInputValue(todo.dueDate)}
          key={`due-${todo.id}-${todo.dueDate ?? "none"}`}
          onChange={(e) => {
            const value = e.target.value;
            // The PATCH contract omits `undefined`, so clearing is a
            // no-op server-side — revert rather than fake a clear.
            if (value === "") {
              e.target.value = toDateInputValue(todo.dueDate);
              return;
            }
            updateTodo.mutate({
              todoId: todo.id,
              body: { dueDate: fromDateInputValue(value) },
            });
          }}
          disabled={saving}
        />
      </div>

      <dl className="space-y-1 border-t pt-3 text-xs text-muted-foreground">
        <div className="flex justify-between">
          <dt>Created</dt>
          <dd>{new Date(todo.createdAt).toLocaleString()}</dd>
        </div>
        <div className="flex justify-between">
          <dt>Last updated</dt>
          <dd>{new Date(todo.updatedAt).toLocaleString()}</dd>
        </div>
        {todo.completedAt && (
          <div className="flex justify-between">
            <dt>Completed</dt>
            <dd>{new Date(todo.completedAt).toLocaleString()}</dd>
          </div>
        )}
      </dl>
    </div>
  );
}

/* ------------------------------ Subtasks tab ------------------------------ */

function SubtasksTab({ todo }: { todo: TGetTodoByIdResponse }) {
  const [quickTitle, setQuickTitle] = useState("");
  const [pendingSubtasks, setPendingSubtasks] = useState<
    { key: number; title: string }[]
  >([]);
  const pendingKey = useRef(0);
  const createTodo = useCreateTodo();

  const subtasks = todo.children ?? [];
  const completedSubtasks = subtasks.filter(
    (child) => child.status === "completed",
  ).length;
  const subtaskProgress =
    subtasks.length > 0
      ? Math.round((completedSubtasks / subtasks.length) * 100)
      : 0;

  const handleQuickAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    const title = quickTitle.trim();
    if (!title || createTodo.isPending) return;
    const key = ++pendingKey.current;
    setQuickTitle("");
    setPendingSubtasks((prev) => [...prev, { key, title }]);
    try {
      await createTodo.mutateAsync({
        body: { title, parentTodoId: todo.id },
      });
    } catch {
      toast.error("Failed to add subtask");
    } finally {
      setPendingSubtasks((prev) => prev.filter((p) => p.key !== key));
    }
  };

  return (
    <div className="space-y-3 p-4">
      {subtasks.length > 0 || pendingSubtasks.length > 0 ? (
        <>
          <div className="flex items-center gap-2">
            <ListTree className="h-4 w-4 text-muted-foreground" />
            <Progress
              value={subtaskProgress}
              className="h-1.5 flex-1"
              aria-label={`${completedSubtasks} of ${subtasks.length} subtasks complete`}
            />
            <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">
              {completedSubtasks}/{subtasks.length}
            </span>
          </div>
          <ul className="space-y-1" aria-label="Subtasks">
            {subtasks.map((child) => (
              <SubtaskRow key={child.id} subtask={child} />
            ))}
            {pendingSubtasks.map((p) => (
              <SubtaskGhostRow key={p.key} title={p.title} />
            ))}
          </ul>
        </>
      ) : (
        <div className="py-4 text-center text-sm text-muted-foreground">
          <ListTree className="mx-auto mb-2 h-10 w-10 opacity-50" />
          <p>No subtasks yet.</p>
          <p className="text-xs">Break this task down below.</p>
        </div>
      )}

      <form
        onSubmit={(e) => void handleQuickAdd(e)}
        className="flex items-center gap-2 border-t pt-3"
      >
        <Plus className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        <Input
          value={quickTitle}
          onChange={(e) => setQuickTitle(e.target.value)}
          placeholder="Add a subtask, press Enter"
          aria-label={`Add a subtask to ${todo.title}`}
          disabled={createTodo.isPending}
          className="h-8 border-0 bg-transparent px-0 text-sm shadow-none focus-visible:ring-0"
        />
      </form>
      <p className="text-[11px] text-muted-foreground">
        Subtasks stay one level deep — convert one to a top-level task to
        break it out.
      </p>
    </div>
  );
}

/* ----------------------------- Attachments tab ---------------------------- */

function AttachmentsTab({ todo }: { todo: TGetTodoByIdResponse }) {
  const uploadAttachment = useUploadTodoAttachment();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const uploadKey = useRef(0);
  const [isDragging, setIsDragging] = useState(false);
  const [uploads, setUploads] = useState<
    { key: number; file: File; status: "uploading" | "error" }[]
  >([]);
  const busy = uploads.some((u) => u.status === "uploading");

  const uploadOne = async (key: number, file: File) => {
    setUploads((prev) =>
      prev.map((u) => (u.key === key ? { ...u, status: "uploading" as const } : u)),
    );
    try {
      await uploadAttachment.mutateAsync({ todoId: todo.id, file });
      setUploads((prev) => prev.filter((u) => u.key !== key));
      toast.success(`${file.name} uploaded`);
    } catch {
      setUploads((prev) =>
        prev.map((u) => (u.key === key ? { ...u, status: "error" as const } : u)),
      );
      toast.error(`Failed to upload ${file.name}`);
    }
  };

  const handleFiles = (files: File[]) => {
    for (const file of files) {
      if (file.size > MAX_ATTACHMENT_BYTES) {
        toast.error(`File ${file.name} is too large (max 10MB)`);
        continue;
      }
      const key = ++uploadKey.current;
      setUploads((prev) => [...prev, { key, file, status: "uploading" }]);
      void uploadOne(key, file);
    }
  };

  const handleFileSelect = async (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const files = Array.from(event.target.files ?? []);
    if (fileInputRef.current) fileInputRef.current.value = "";
    handleFiles(files);
  };

  return (
    <div className="space-y-4 p-4">
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={(e) => {
          e.preventDefault();
          setIsDragging(false);
          handleFiles(Array.from(e.dataTransfer.files ?? []));
        }}
        className={cn(
          "rounded-lg border border-dashed p-3 transition-colors",
          isDragging
            ? "border-primary bg-primary/5"
            : "border-muted-foreground/25",
        )}
        aria-label="Upload attachments dropzone"
      >
        <input
          ref={fileInputRef}
          type="file"
          multiple
          className="hidden"
          aria-label="Upload attachments"
          onChange={(e) => void handleFileSelect(e)}
          disabled={busy}
        />
        <Button
          type="button"
          variant="outline"
          className="w-full border-dashed"
          onClick={() => fileInputRef.current?.click()}
          disabled={busy}
        >
          {busy ? (
            <Loader2 className="h-4 w-4 mr-2 animate-spin" />
          ) : (
            <Upload className="h-4 w-4 mr-2" />
          )}
          {busy ? "Uploading…" : "Upload files"}
        </Button>
        <p className="mt-1 text-[11px] text-muted-foreground">
          Max 10MB per file. Drag &amp; drop here or browse.
        </p>
      </div>
      {uploads.length > 0 && (
        <ul className="space-y-2" aria-label="Upload progress">
          {uploads.map((u) => (
            <li
              key={u.key}
              className="flex items-center justify-between gap-2 p-3 bg-muted/50 rounded-lg"
            >
              <div className="flex items-center gap-3 min-w-0 flex-1">
                {u.status === "uploading" ? (
                  <Loader2 className="h-5 w-5 text-muted-foreground flex-shrink-0 animate-spin" />
                ) : (
                  <FileText className="h-5 w-5 text-destructive flex-shrink-0" />
                )}
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium truncate" title={u.file.name}>
                    {u.file.name}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {u.status === "uploading"
                      ? `Uploading… • ${formatFileSize(u.file.size)}`
                      : `Upload failed • ${formatFileSize(u.file.size)}`}
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                {u.status === "error" && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => void uploadOne(u.key, u.file)}
                    aria-label={`Retry upload ${u.file.name}`}
                  >
                    <RotateCcw className="h-4 w-4" />
                  </Button>
                )}
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    setUploads((prev) => prev.filter((x) => x.key !== u.key))
                  }
                  disabled={u.status === "uploading"}
                  aria-label={`Dismiss ${u.file.name}`}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}
      <TodoAttachments
        todoId={todo.id}
        attachments={todo.attachments ?? []}
        disabled={busy}
      />
    </div>
  );
}
