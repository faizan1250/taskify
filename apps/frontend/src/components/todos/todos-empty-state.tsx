import { TodoCreateForm } from "@/components/todos/todo-create-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { ClipboardList, Plus, RotateCcw, SearchX } from "lucide-react";
import { Link } from "react-router-dom";

export interface EmptyFilterSummary {
  key: string;
  label: string;
}

interface TodosFilteredEmptyStateProps {
  /** Human tab name when the status tab is narrowed (null when "All"). */
  tabLabel: string | null;
  /** Active filter labels, reusing the filter-bar chip wording. */
  filters: readonly EmptyFilterSummary[];
  hasSearch: boolean;
  hasDueRange: boolean;
  hasOverdueOnly: boolean;
  onClearAll: () => void;
}

function tipFor({
  hasSearch,
  hasDueRange,
  hasOverdueOnly,
}: Pick<
  TodosFilteredEmptyStateProps,
  "hasSearch" | "hasDueRange" | "hasOverdueOnly"
>): string {
  // Ordered by how precisely the tip names the excluding constraint.
  // Each claim mirrors backend list semantics: search matches title or
  // description, date bounds and overdue compare against due_date, so
  // tasks without a due date never match those filters.
  if (hasSearch)
    return "Search looks at titles and descriptions — a shorter keyword matches more.";
  if (hasDueRange)
    return "Date filters only match tasks with a due date inside the range.";
  if (hasOverdueOnly)
    return "Overdue only matches dated tasks past their due date.";
  return "Each filter narrows the list — removing one may be enough.";
}

/**
 * Dead-end recovery for zero-result filtered lists. Names the active
 * constraints at the point of failure and offers a one-click reset,
 * instead of leaving "try adjusting your filters" with no action.
 */
export function TodosFilteredEmptyState({
  tabLabel,
  filters,
  hasSearch,
  hasDueRange,
  hasOverdueOnly,
  onClearAll,
}: TodosFilteredEmptyStateProps) {
  const constraints = [
    ...(tabLabel ? [`Tab: ${tabLabel}`] : []),
    ...filters.map((filter) => filter.label),
  ];
  return (
    <Card>
      <CardContent className="py-12 text-center">
        <SearchX
          className="mx-auto mb-4 h-12 w-12 text-muted-foreground opacity-50"
          aria-hidden="true"
        />
        <h3 className="mb-2 text-lg font-medium">
          No tasks match these filters
        </h3>
        {constraints.length > 0 && (
          <p className="mx-auto mb-1 max-w-md text-sm text-muted-foreground">
            Active: {constraints.join(" · ")}
          </p>
        )}
        <p className="mx-auto mb-4 max-w-md text-sm text-muted-foreground">
          {tipFor({ hasSearch, hasDueRange, hasOverdueOnly })}
        </p>
        <Button onClick={onClearAll}>
          <RotateCcw className="mr-2 h-4 w-4" aria-hidden="true" />
          Clear all filters
        </Button>
      </CardContent>
    </Card>
  );
}

/**
 * First-run coaching for a genuinely empty task list. Teaches the three
 * moves that make the rest of the app useful (capture, organize, triage)
 * instead of a bare "create your first task" prompt.
 */
export function TodosFirstRunEmptyState() {
  return (
    <Card>
      <CardContent className="py-12 text-center">
        <ClipboardList
          className="mx-auto mb-4 h-12 w-12 text-muted-foreground opacity-50"
          aria-hidden="true"
        />
        <h3 className="mb-2 text-lg font-medium">Create your first task</h3>
        <ol className="mx-auto mb-4 max-w-md space-y-1 text-left text-sm text-muted-foreground">
          <li>
            <span className="font-medium text-foreground">1. Capture it</span> —
            press{" "}
            <kbd className="rounded border bg-muted px-1 font-sans">C</kbd>{" "}
            anywhere to create a task.
          </li>
          <li>
            <span className="font-medium text-foreground">2. Organize it</span>{" "}
            — group tasks with{" "}
            <Link to="/categories" className="underline underline-offset-2">
              categories
            </Link>{" "}
            so filters stay useful as you grow.
          </li>
          <li>
            <span className="font-medium text-foreground">3. Triage it</span> —
            the{" "}
            <Link to="/" className="underline underline-offset-2">
              dashboard
            </Link>{" "}
            surfaces what is overdue and due this week.
          </li>
        </ol>
        <TodoCreateForm>
          <Button>
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            Create Task
          </Button>
        </TodoCreateForm>
      </CardContent>
    </Card>
  );
}
