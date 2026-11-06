import { useApiClient } from "@/api";
import { QUERY_KEYS } from "@/api/query-utils";
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
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { useQueryClient } from "@tanstack/react-query";
import { Archive, CheckCheck, CirclePlay, Trash2, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface BulkActionsBarProps {
  selectedIds: string[];
  totalOnPage: number;
  onToggleSelectAll: (checked: boolean) => void;
  /** Remove succeeded ids from selection; failed ids stay selected. */
  onResolve: (succeededIds: string[]) => void;
  onClear: () => void;
}

type BulkStatus = "active" | "completed" | "archived";

const STATUS_LABEL: Record<BulkStatus, string> = {
  completed: "completed",
  active: "active",
  archived: "archived",
};

/**
 * Page-scoped bulk action bar for the tasks list. Runs one request per
 * selected task through the existing single-item endpoints (no bulk
 * contract needed) and reports a single summary toast — partial failures
 * keep the failed ids selected so nothing silently disappears.
 */
export function BulkActionsBar({
  selectedIds,
  totalOnPage,
  onToggleSelectAll,
  onResolve,
  onClear,
}: BulkActionsBarProps) {
  const api = useApiClient();
  const queryClient = useQueryClient();
  const [pending, setPending] = useState<string | null>(null);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const count = selectedIds.length;
  const allSelected = totalOnPage > 0 && count === totalOnPage;
  const busy = pending !== null;

  const invalidate = () => {
    queryClient.invalidateQueries({
      queryKey: [QUERY_KEYS.TODOS.ALL_TODOS],
    });
    queryClient.invalidateQueries({
      queryKey: [QUERY_KEYS.TODOS.TODO_STATS],
    });
  };

  const runBulkStatus = async (status: BulkStatus) => {
    if (count === 0 || busy) return;
    setPending(`status-${status}`);
    const ids = [...selectedIds];
    const failed: string[] = [];
    for (const id of ids) {
      try {
        const res = await api.Todo.updateTodo({
          params: { id },
          body: { status },
        });
        if (res.status !== 200) failed.push(id);
      } catch {
        failed.push(id);
      }
    }
    const succeeded = ids.filter((id) => !failed.includes(id));
    invalidate();
    onResolve(succeeded);
    setPending(null);
    if (failed.length === 0) {
      toast.success(
        `${succeeded.length} task${succeeded.length === 1 ? "" : "s"} marked ${STATUS_LABEL[status]}`,
      );
    } else if (succeeded.length === 0) {
      toast.error(
        `Couldn't update ${failed.length} task${failed.length === 1 ? "" : "s"} — still selected so you can retry`,
      );
    } else {
      toast.error(
        `${succeeded.length} updated, ${failed.length} failed — failed tasks stay selected`,
      );
    }
  };

  const runBulkDelete = async () => {
    if (count === 0 || busy) return;
    setPending("delete");
    setDeleteOpen(false);
    const ids = [...selectedIds];
    const failed: string[] = [];
    for (const id of ids) {
      try {
        const res = await api.Todo.deleteTodo({ params: { id } });
        if (res.status !== 204) failed.push(id);
      } catch {
        failed.push(id);
      }
    }
    const succeeded = ids.filter((id) => !failed.includes(id));
    invalidate();
    onResolve(succeeded);
    setPending(null);
    if (failed.length === 0) {
      toast.success(
        `${succeeded.length} task${succeeded.length === 1 ? "" : "s"} deleted`,
      );
    } else if (succeeded.length === 0) {
      toast.error(
        `Couldn't delete ${failed.length} task${failed.length === 1 ? "" : "s"} — still selected so you can retry`,
      );
    } else {
      toast.error(
        `${succeeded.length} deleted, ${failed.length} failed — failed tasks stay selected`,
      );
    }
  };

  return (
    <>
      <Card aria-label="Bulk actions">
        <CardContent className="py-3">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
            <div className="flex items-center gap-2">
              <Checkbox
                checked={
                  allSelected ? true : count > 0 ? "indeterminate" : false
                }
                onCheckedChange={(checked) =>
                  onToggleSelectAll(checked === true)
                }
                disabled={busy || totalOnPage === 0}
                aria-label={
                  allSelected
                    ? "Deselect all tasks on this page"
                    : "Select all tasks on this page"
                }
              />
              <span
                role="status"
                aria-live="polite"
                className="text-sm font-medium tabular-nums"
              >
                {count > 0
                  ? `${count} selected`
                  : `Select all ${totalOnPage} task${totalOnPage === 1 ? "" : "s"}`}
              </span>
            </div>
            <div
              className="flex flex-wrap items-center gap-2 sm:ml-auto"
              aria-label="Bulk actions for selected tasks"
            >
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={count === 0 || busy}
                onClick={() => void runBulkStatus("completed")}
              >
                <CheckCheck aria-hidden="true" />
                {pending === "status-completed"
                  ? "Completing…"
                  : "Mark complete"}
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={count === 0 || busy}
                onClick={() => void runBulkStatus("active")}
              >
                <CirclePlay aria-hidden="true" />
                {pending === "status-active"
                  ? "Activating…"
                  : "Mark active"}
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={count === 0 || busy}
                onClick={() => void runBulkStatus("archived")}
              >
                <Archive aria-hidden="true" />
                {pending === "status-archived"
                  ? "Archiving…"
                  : "Archive"}
              </Button>
              <Button
                type="button"
                variant="destructive"
                size="sm"
                disabled={count === 0 || busy}
                onClick={() => setDeleteOpen(true)}
              >
                <Trash2 aria-hidden="true" />
                {pending === "delete" ? "Deleting…" : "Delete"}
              </Button>
              {count > 0 && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={busy}
                  onClick={onClear}
                  aria-label="Clear task selection"
                >
                  <X aria-hidden="true" />
                  Clear
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Delete {count} task{count === 1 ? "" : "s"}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              This permanently deletes the selected task
              {count === 1 ? "" : "s"} and
              {count === 1 ? " its" : " their"} subtasks. This action cannot
              be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={busy}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                void runBulkDelete();
              }}
              disabled={busy}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {pending === "delete" ? "Deleting…" : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
