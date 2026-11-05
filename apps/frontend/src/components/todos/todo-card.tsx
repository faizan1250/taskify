import { TodoCommentsDialog } from "./todo-comments-dialog";
import { TodoEditForm } from "./todo-edit-form";
import {
  useUpdateTodo,
  useDeleteTodo,
  type TGetTodosResponse,
} from "@/api/hooks/use-todo-query";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import {
  MoreHorizontal,
  Calendar,
  MessageSquare,
  Edit,
  Trash2,
  Clock,
  CheckCircle2,
  Circle,
  Archive,
  Paperclip,
  ChevronDown,
  ChevronRight,
  ListTree,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface TodoCardProps {
  todo: TGetTodosResponse["data"][number];
  compact?: boolean;
}

const priorityConfig = {
  low: {
    variant: "success",
    label: "Low",
  },
  medium: {
    variant: "warning",
    label: "Medium",
  },
  high: {
    variant: "danger",
    label: "High",
  },
} as const;

const statusConfig = {
  draft: { icon: Circle, color: "text-gray-500", label: "Draft" },
  active: { icon: Clock, color: "text-blue-500", label: "Active" },
  completed: {
    icon: CheckCircle2,
    color: "text-green-500",
    label: "Completed",
  },
  archived: { icon: Archive, color: "text-gray-500", label: "Archived" },
};

export function TodoCard({ todo, compact = false }: TodoCardProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [subtasksExpanded, setSubtasksExpanded] = useState(false);
  const updateTodo = useUpdateTodo();
  const deleteTodo = useDeleteTodo();

  const StatusIcon = statusConfig[todo.status].icon;
  const subtasks = todo.children ?? [];
  const completedSubtasks = subtasks.filter(
    (child) => child.status === "completed",
  ).length;

  const handleStatusToggle = async () => {
    const newStatus = todo.status === "completed" ? "active" : "completed";
    await updateTodo.mutateAsync({
      todoId: todo.id,
      body: {
        status: newStatus,
      },
    });
    toast.success(
      newStatus === "completed"
        ? "Task marked as completed!"
        : "Task marked as active",
    );
  };

  const handleSubtaskToggle = async (childId: string, completed: boolean) => {
    try {
      await updateTodo.mutateAsync({
        todoId: childId,
        body: { status: completed ? "active" : "completed" },
      });
    } catch {
      toast.error("Failed to update subtask");
    }
  };

  const handleDelete = async () => {
    try {
      await deleteTodo.mutateAsync({ todoId: todo.id });
      toast.success("Task deleted successfully");
      setShowDeleteDialog(false);
    } catch {
      toast.error("Failed to delete task");
    }
  };

  const isOverdue =
    todo.dueDate &&
    new Date(todo.dueDate) < new Date() &&
    todo.status !== "completed";

  return (
    <>
      <Card
        className={cn(
          "transition-all hover:shadow-md",
          todo.status === "completed" && "opacity-75",
          compact &&
            "shadow-none border-0 border-l-2 border-l-transparent hover:border-l-primary",
        )}
      >
        <CardContent className={cn("p-4", compact && "p-3")}>
          <div className="space-y-3">
            {/* Header with status and actions */}
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-start gap-3 flex-1 min-w-0">
                <Checkbox
                  checked={todo.status === "completed"}
                  onCheckedChange={handleStatusToggle}
                  disabled={updateTodo.isPending}
                  className="mt-0.5"
                />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <StatusIcon
                      className={cn("h-4 w-4", statusConfig[todo.status].color)}
                    />
                    <h3
                      className={cn(
                        "font-semibold truncate",
                        compact ? "text-sm" : "text-base",
                        todo.status === "completed" &&
                          "line-through text-muted-foreground",
                      )}
                    >
                      {todo.title}
                    </h3>
                    {isOverdue && (
                      <span
                        className="relative flex h-2 w-2 shrink-0"
                        role="status"
                        aria-label="Overdue"
                      >
                        <span className="absolute inline-flex h-full w-full rounded-full bg-red-500 opacity-60 motion-safe:animate-ping" />
                        <span className="relative inline-flex h-2 w-2 rounded-full bg-red-500" />
                      </span>
                    )}
                  </div>

                  {todo.description && !compact && (
                    <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
                      {todo.description}
                    </p>
                  )}
                </div>
              </div>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-8 w-8" aria-label={`Task actions for ${todo.title}`}>
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <TodoEditForm todo={todo}>
                    <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                      <Edit className="h-4 w-4 mr-2" />
                      Edit
                    </DropdownMenuItem>
                  </TodoEditForm>

                  <TodoCommentsDialog todoId={todo.id}>
                    <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                      <MessageSquare className="h-4 w-4 mr-2" />
                      Comments ({todo.comments?.length || 0})
                    </DropdownMenuItem>
                  </TodoCommentsDialog>

                  <DropdownMenuSeparator />

                  <DropdownMenuItem
                    onSelect={() => setShowDeleteDialog(true)}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Metadata */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 flex-wrap">
                <Badge variant={priorityConfig[todo.priority].variant}>
                  {priorityConfig[todo.priority].label}
                </Badge>

                {todo.category && (
                  <Badge
                    variant="secondary"
                    className="flex items-center gap-1"
                  >
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

                {todo.comments && todo.comments.length > 0 && (
                  <Badge variant="outline" className="flex items-center gap-1">
                    <MessageSquare className="h-3 w-3" />
                    {todo.comments.length}
                  </Badge>
                )}

                {todo.attachments && todo.attachments.length > 0 && (
                  <Badge variant="outline" className="flex items-center gap-1">
                    <Paperclip className="h-3 w-3" />
                    {todo.attachments.length}
                  </Badge>
                )}

                {subtasks.length > 0 && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="flex items-center gap-1 h-6 px-2 text-xs"
                    onClick={() => setSubtasksExpanded((v) => !v)}
                    aria-expanded={subtasksExpanded}
                    aria-label={
                      subtasksExpanded ? "Collapse subtasks" : "Expand subtasks"
                    }
                  >
                    <ListTree className="h-3 w-3" />
                    {completedSubtasks}/{subtasks.length}
                    {subtasksExpanded ? (
                      <ChevronDown className="h-3 w-3" />
                    ) : (
                      <ChevronRight className="h-3 w-3" />
                    )}
                  </Button>
                )}
              </div>

              {!compact && (
                <div className="text-xs text-muted-foreground">
                  {todo.status === "completed" && todo.completedAt
                    ? `Completed ${new Date(todo.completedAt).toLocaleDateString()}`
                    : `Updated ${new Date(todo.updatedAt).toLocaleDateString()}`}
                </div>
              )}
            </div>

            {/* Expandable subtasks */}
            {subtasksExpanded && subtasks.length > 0 && (
              <ul className="mt-1 space-y-1 border-t pt-2" aria-label="Subtasks">
                {subtasks.map((child) => {
                  const isDone = child.status === "completed";
                  return (
                    <li
                      key={child.id}
                      className="flex items-center gap-2 rounded-md px-2 py-1.5 hover:bg-muted/60"
                    >
                      <Checkbox
                        checked={isDone}
                        onCheckedChange={() =>
                          handleSubtaskToggle(child.id, isDone)
                        }
                        disabled={updateTodo.isPending}
                        aria-label={`Toggle subtask ${child.title}`}
                        className="h-3.5 w-3.5"
                      />
                      <span
                        className={cn(
                          "flex-1 truncate text-sm",
                          isDone && "line-through text-muted-foreground",
                        )}
                      >
                        {child.title}
                      </span>
                      {child.priority && (
                        <span
                          className={cn(
                            "h-1.5 w-1.5 rounded-full shrink-0",
                            child.priority === "high" && "bg-red-500",
                            child.priority === "medium" && "bg-yellow-500",
                            child.priority === "low" && "bg-green-500",
                          )}
                          title={`${child.priority} priority`}
                        />
                      )}
                      <span className="shrink-0 text-[11px] capitalize text-muted-foreground">
                        {child.status}
                      </span>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        </CardContent>
      </Card>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Task</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{todo.title}"? This action cannot
              be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              disabled={deleteTodo.isPending}
            >
              {deleteTodo.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
