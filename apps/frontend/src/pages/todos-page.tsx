import { useGetAllCategories } from "@/api/hooks/use-category-query";
import {
  useGetAllTodos,
  useGetTodoStats,
  type TGetTodosQuery,
} from "@/api/hooks/use-todo-query";
import { ShortcutHint } from "@/components/command-palette";
import { QueryErrorState } from "@/components/query-error-state";
import { TodoCard } from "@/components/todos/todo-card";
import { TodoCreateForm } from "@/components/todos/todo-create-form";
import { TodoDetailDrawer } from "@/components/todos/todo-detail-drawer";
import {
  TodosFilteredEmptyState,
  TodosFirstRunEmptyState,
} from "@/components/todos/todos-empty-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  TODOS_SORT_OPTIONS,
  toDueBoundISO,
  useTodosSearchParams,
  type TodosPriorityFilter,
  type TodosSort,
  type TodosTab,
} from "@/hooks/use-todos-search-params";
import { useSavedViews } from "@/hooks/use-saved-views";
import { useTodosFilterChips } from "@/hooks/use-todos-filter-chips";
import { useTodosHotkeys } from "@/hooks/use-todos-hotkeys";
import { cn } from "@/lib/utils";
import { Plus, Search, ChevronDown, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

export function TodosPage() {
  const {
    tab,
    category,
    priority,
    sort,
    order,
    page,
    sortValue,
    overdueOnly,
    dueFrom,
    dueTo,
    searchInput,
    setSearchInput,
    debouncedSearch,
    setTab,
    setCategory,
    setPriority,
    setSortOrder,
    setPage,
    setOverdueOnly,
    setDueFrom,
    setDueTo,
    clearDueRange,
    clearFilters,
    setSelectedTaskId,
    hasActiveFilters,
    activeFilterCount,
  } = useTodosSearchParams();

  const [selectedTodos, setSelectedTodos] = useState<string[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [viewName, setViewName] = useState("");
  const searchRef = useRef<HTMLInputElement>(null);

  const dueRangeInvalid =
    dueFrom !== null && dueTo !== null && dueFrom > dueTo;

  const {
    data: todos,
    isLoading,
    isError: todosError,
    isFetching: todosFetching,
    refetch: refetchTodos,
  } = useGetAllTodos({
    query: {
      page,
      limit: 20,
      search: debouncedSearch || undefined,
      status: tab === "all" ? undefined : tab,
      categoryId: category === "all" ? undefined : category,
      priority: priority === "all" ? undefined : priority,
      overdue: overdueOnly ? true : undefined,
      dueFrom: dueRangeInvalid
        ? undefined
        : toDueBoundISO(dueFrom, "from"),
      dueTo: dueRangeInvalid ? undefined : toDueBoundISO(dueTo, "to"),
      sort,
      order,
    },
  });

  const { data: categories } = useGetAllCategories({
    query: { page: 1, limit: 100 },
  });

  const { data: stats } = useGetTodoStats();

  const chips = useTodosFilterChips({
    search: debouncedSearch,
    category,
    priority,
    sortValue,
    isDefaultSort: sortValue === "updated_at-desc",
    overdueOnly,
    dueFrom,
    dueTo,
    categories: categories?.data,
    onClearSearch: () => setSearchInput(""),
    onClearCategory: () => setCategory("all"),
    onClearPriority: () => setPriority("all"),
    onClearSort: () => setSortOrder("updated_at", "desc"),
    onClearOverdue: () => setOverdueOnly(false),
    onClearDueFrom: () => setDueFrom(undefined),
    onClearDueTo: () => setDueTo(undefined),
  });

  const { views, saveView, applyView, deleteView } = useSavedViews();

  const handleSaveView = () => {
    const params: Record<string, string> = {};
    if (tab !== "all") params.tab = tab;
    if (category !== "all") params.category = category;
    if (priority !== "all") params.priority = priority;
    if (sort !== "updated_at") params.sort = sort;
    if (order !== "desc") params.order = order;
    if (debouncedSearch !== "") params.q = debouncedSearch;
    if (overdueOnly) params.overdue = "true";
    if (dueFrom !== null) params.dueFrom = dueFrom;
    if (dueTo !== null) params.dueTo = dueTo;
    saveView(viewName, params);
    setViewName("");
  };

  // Return to the top of the list when paginating (skipped on mount).
  const didMountRef = useRef(false);
  useEffect(() => {
    if (!didMountRef.current) {
      didMountRef.current = true;
      return;
    }
    window.scrollTo(0, 0);
  }, [page]);

  const handleSelectTodo = (todoId: string, checked: boolean) => {
    if (checked) {
      setSelectedTodos((prev) => [...prev, todoId]);
    } else {
      setSelectedTodos((prev) => prev.filter((id) => id !== todoId));
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked && todos?.data) {
      setSelectedTodos(todos.data.map((todo) => todo.id));
    } else {
      setSelectedTodos([]);
    }
  };

  const tabCounts = {
    all: stats?.total || 0,
    draft: stats?.draft || 0,
    active: stats?.active || 0,
    completed: stats?.completed || 0,
    archived: stats?.archived || 0,
  };

  const visibleTodos = todos?.data ?? [];

  const { activeIndex, setActiveIndex } = useTodosHotkeys({
    enabled: !isLoading,
    count: visibleTodos.length,
    onOpenCreate: () => setCreateOpen(true),
    onFocusSearch: () => searchRef.current?.focus(),
    onToggleSelect: (index) => {
      const target = visibleTodos[index];
      if (target) {
        handleSelectTodo(target.id, !selectedTodos.includes(target.id));
      }
    },
    onEditActive: () => {
      const target = visibleTodos[activeIndex];
      if (target) setSelectedTaskId(target.id, { edit: true });
    },
    onClear: () => {
      setSelectedTodos([]);
      setActiveIndex(0);
    },
  });

  const activeTodo = visibleTodos[activeIndex];
  const activeId = activeTodo?.id;

  // Keep the keyboard-active row in view (instant: no motion to reduce).
  useEffect(() => {
    if (!activeId) return;
    document
      .querySelector(`[data-todo-row="${activeId}"]`)
      ?.scrollIntoView({ block: "nearest" });
  }, [activeId]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Tasks</h1>
          <p className="text-muted-foreground">
            Manage and organize your tasks
          </p>
        </div>
        <TodoCreateForm open={createOpen} onOpenChange={setCreateOpen}>
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            New Task
          </Button>
        </TodoCreateForm>
      </div>

      {/* Filters and Search */}
      <Card>
        <Collapsible defaultOpen>
          <CardHeader>
            <CollapsibleTrigger className="group flex w-full cursor-pointer items-center justify-between gap-2 text-left">
              <span className="flex items-center gap-2">
                <CardTitle className="text-lg">Filters & Search</CardTitle>
                {activeFilterCount > 0 && (
                  <Badge
                    variant="secondary"
                    aria-label={`${activeFilterCount} active filters`}
                  >
                    {activeFilterCount}
                  </Badge>
                )}
              </span>
              <ChevronDown
                className="h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-150 group-data-[state=open]:rotate-180"
                aria-hidden="true"
              />
            </CollapsibleTrigger>
            <p
              className="text-sm text-muted-foreground"
              role="status"
              aria-live="polite"
            >
              {isLoading
                ? "Loading tasks…"
                : `${todos?.total ?? 0} result${todos?.total === 1 ? "" : "s"}${
                    todosFetching ? " · Updating…" : ""
                  }`}
            </p>
          </CardHeader>
          <CollapsibleContent>
            <CardContent className="space-y-4">
              <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Search tasks..."
                    ref={searchRef}
                    value={searchInput}
                    onChange={(e) => {
                      setSearchInput(e.target.value);
                    }}
                    className="pl-10"
                  />
                </div>
                <Select value={category} onValueChange={setCategory}>
                  <SelectTrigger className="w-full sm:w-48">
                    <SelectValue placeholder="Category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                    {categories?.data?.map((category) => (
                      <SelectItem key={category.id} value={category.id}>
                        <div className="flex items-center gap-2">
                          <div
                            className="w-3 h-3 rounded-full"
                            style={{ backgroundColor: category.color }}
                          />
                          {category.name}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Select
                  value={priority}
                  onValueChange={(value) =>
                    setPriority(value as TodosPriorityFilter)
                  }
                >
                  <SelectTrigger className="w-full sm:w-32">
                    <SelectValue placeholder="Priority" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All</SelectItem>
                    <SelectItem value="high">High</SelectItem>
                    <SelectItem value="medium">Medium</SelectItem>
                    <SelectItem value="low">Low</SelectItem>
                  </SelectContent>
                </Select>
                <Select
                  value={sortValue}
                  onValueChange={(value) => {
                    const [nextSort, nextOrder] = value.split("-");
                    setSortOrder(
                      nextSort as TodosSort,
                      nextOrder as "asc" | "desc",
                    );
                  }}
                >
                  <SelectTrigger className="w-full sm:w-40">
                    <SelectValue placeholder="Sort by" />
                  </SelectTrigger>
                  <SelectContent>
                    {TODOS_SORT_OPTIONS.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {hasActiveFilters && (
                  <Button variant="outline" onClick={clearFilters}>
                    Clear
                  </Button>
                )}
              </div>
              {chips.length > 0 && (
                <div
                  className="flex flex-wrap items-center gap-2"
                  aria-label="Active filters"
                >
                  {chips.map((chip) => (
                    <Badge key={chip.key} variant="secondary">
                      <span className="max-w-48 truncate">{chip.label}</span>
                      <button
                        type="button"
                        onClick={chip.onRemove}
                        aria-label={`Remove ${chip.label} filter`}
                        className="cursor-pointer rounded-sm p-0.5 text-muted-foreground transition-colors hover:text-foreground"
                      >
                        <X className="h-3 w-3" aria-hidden="true" />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
          <div className="flex flex-col sm:flex-row sm:items-end gap-4 border-t pt-4">
            <div className="flex items-center gap-2.5">
              <Switch
                id="overdue-only"
                checked={overdueOnly}
                onCheckedChange={setOverdueOnly}
                aria-describedby={
                  stats?.overdue ? "overdue-count" : undefined
                }
              />
              <Label htmlFor="overdue-only" className="cursor-pointer">
                Overdue only
              </Label>
              {typeof stats?.overdue === "number" && stats.overdue > 0 && (
                <span
                  id="overdue-count"
                  className="rounded-full bg-destructive/10 px-2 py-0.5 text-xs font-medium text-destructive"
                >
                  {stats.overdue}
                </span>
              )}
            </div>
            <div className="flex flex-col sm:flex-row gap-4 sm:items-end">
              <div className="grid gap-1.5">
                <Label htmlFor="due-from" className="text-xs">
                  Due from
                </Label>
                <Input
                  id="due-from"
                  type="date"
                  value={dueFrom ?? ""}
                  max={dueTo ?? undefined}
                  onChange={(e) => setDueFrom(e.target.value || undefined)}
                  className="w-full sm:w-40"
                  aria-invalid={dueRangeInvalid}
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="due-to" className="text-xs">
                  Due to
                </Label>
                <Input
                  id="due-to"
                  type="date"
                  value={dueTo ?? ""}
                  min={dueFrom ?? undefined}
                  onChange={(e) => setDueTo(e.target.value || undefined)}
                  className="w-full sm:w-40"
                  aria-invalid={dueRangeInvalid}
                />
              </div>
              {(dueFrom !== null || dueTo !== null) && (
                <Button variant="ghost" size="sm" onClick={clearDueRange}>
                  Reset dates
                </Button>
              )}
            </div>
          </div>
          {dueRangeInvalid && (
            <p role="alert" className="text-xs text-destructive">
              Start date must be on or before the end date — date filter
              paused until fixed.
            </p>
          )}
          <div className="space-y-3 border-t pt-4">
            <Label className="text-xs text-muted-foreground">Saved views</Label>
            {views.length > 0 && (
              <div
                className="flex flex-wrap items-center gap-2"
                aria-label="Saved views"
              >
                {views.map((view) => (
                  <Badge key={view.id} variant="outline">
                    <button
                      type="button"
                      onClick={() => applyView(view.id)}
                      title="Apply this saved view"
                      className="cursor-pointer rounded-sm transition-colors hover:text-primary"
                    >
                      <span className="max-w-40 truncate">{view.name}</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => deleteView(view.id)}
                      aria-label={`Delete saved view ${view.name}`}
                      className="cursor-pointer rounded-sm p-0.5 text-muted-foreground transition-colors hover:text-foreground"
                    >
                      <X className="h-3 w-3" aria-hidden="true" />
                    </button>
                  </Badge>
                ))}
              </div>
            )}
            <div className="flex flex-col gap-2 sm:flex-row">
              <Input
                placeholder="Name this view…"
                value={viewName}
                onChange={(e) => setViewName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleSaveView();
                }}
                maxLength={40}
                className="w-full sm:max-w-56"
                aria-label="Saved view name"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleSaveView}
                disabled={viewName.trim() === ""}
                className="sm:self-center"
              >
                Save current filters
              </Button>
            </div>
          </div>
          </CardContent>
          </CollapsibleContent>
        </Collapsible>
      </Card>

      {/* Status Tabs */}
      <Tabs
        value={tab}
        onValueChange={(value) => setTab(value as TodosTab | "all")}
      >
        <TabsList className="flex w-full justify-start gap-1 overflow-x-auto">
          <TabsTrigger value="all" className="flex-none">
            All ({tabCounts.all})
          </TabsTrigger>
          <TabsTrigger value="draft" className="flex-none">
            Draft ({tabCounts.draft})
          </TabsTrigger>
          <TabsTrigger value="active" className="flex-none">
            Active ({tabCounts.active})
          </TabsTrigger>
          <TabsTrigger value="completed" className="flex-none">
            Completed ({tabCounts.completed})
          </TabsTrigger>
          <TabsTrigger value="archived" className="flex-none">
            Archived ({tabCounts.archived})
          </TabsTrigger>
        </TabsList>

        <TabsContent value={tab} className="space-y-4">
          {/* Bulk Actions */}
          {selectedTodos.length > 0 && (
            <Card>
              <CardContent className="py-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Checkbox
                      checked={selectedTodos.length === todos?.data?.length}
                      onCheckedChange={handleSelectAll}
                    />
                    <span className="text-sm font-medium">
                      {selectedTodos.length} selected
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Todo List */}
          <div className="space-y-3">
            {todosError && (todos?.data?.length ?? 0) > 0 && (
              <QueryErrorState
                variant="banner"
                title="Couldn't refresh tasks — showing last loaded results."
                onRetry={() => void refetchTodos()}
                isRetrying={todosFetching}
              />
            )}
            {isLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Card key={i}>
                    <CardContent className="p-4">
                      <div className="space-y-2">
                        <Skeleton className="h-4 w-full" />
                        <Skeleton className="h-3 w-2/3" />
                        <div className="flex gap-2">
                          <Skeleton className="h-5 w-16" />
                          <Skeleton className="h-5 w-16" />
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : todosError ? (
              <QueryErrorState
                title="Couldn't load tasks"
                description="Check your connection and try again — your filters are kept."
                onRetry={() => void refetchTodos()}
                isRetrying={todosFetching}
              />
            ) : todos?.data?.length ? (
              <>
                {/* Select All Option */}
                {todos.data.length > 1 && (
                  <Card>
                    <CardContent className="py-3">
                      <div className="flex items-center gap-2">
                        <Checkbox
                          checked={selectedTodos.length === todos.data.length}
                          onCheckedChange={handleSelectAll}
                        />
                        <span className="text-sm text-muted-foreground">
                          Select all {todos.data.length} tasks
                        </span>
                      </div>
                    </CardContent>
                  </Card>
                )}

                {todos.data.map((todo, index) => (
                  <div
                    key={todo.id}
                    data-todo-row={todo.id}
                    onMouseEnter={() => setActiveIndex(index)}
                    className={cn(
                      "-m-1 flex items-start gap-3 rounded-[7px] p-1",
                      index === activeIndex &&
                        "ring-2 ring-primary ring-offset-2 ring-offset-background",
                    )}
                  >
                    <div className="pt-4">
                      <Checkbox
                        checked={selectedTodos.includes(todo.id)}
                        onCheckedChange={(checked) =>
                          handleSelectTodo(todo.id, checked as boolean)
                        }
                      />
                    </div>
                    <div className="flex-1">
                      <TodoCard todo={todo} />
                    </div>
                  </div>
                ))}

                {/* Pagination */}
                {todos.totalPages > 1 && (
                  <div className="flex items-center justify-between">
                    <p className="text-sm text-muted-foreground">
                      Showing {(page - 1) * 20 + 1} to{" "}
                      {Math.min(page * 20, todos.total)} of {todos.total} tasks
                    </p>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={page <= 1}
                        onClick={() => setPage(page - 1)}
                      >
                        Previous
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={page >= todos.totalPages}
                        onClick={() => setPage(page + 1)}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                )}
              </>
            ) : hasActiveFilters ? (
              <TodosFilteredEmptyState
                tabLabel={
                  tab === "all"
                    ? null
                    : tab.charAt(0).toUpperCase() + tab.slice(1)
                }
                filters={chips}
                hasSearch={debouncedSearch !== ""}
                hasDueRange={dueFrom !== null || dueTo !== null}
                hasOverdueOnly={overdueOnly}
                onClearAll={clearFilters}
              />
            ) : (
              <TodosFirstRunEmptyState />
            )}
          </div>
          {/* Keyboard hints */}
          <div
            className="flex flex-wrap items-center gap-x-4 gap-y-2 pt-1 text-xs text-muted-foreground"
            aria-label="Task list keyboard shortcuts"
          >
            <span className="inline-flex items-center gap-1.5">
              <ShortcutHint keys={["C"]} /> New task
            </span>
            <span className="inline-flex items-center gap-1.5">
              <ShortcutHint keys={["/"]} /> Search
            </span>
            <span className="inline-flex items-center gap-1.5">
              <ShortcutHint keys={["J", "K"]} /> Move
            </span>
            <span className="inline-flex items-center gap-1.5">
              <ShortcutHint keys={["X"]} /> Select
            </span>
            <span className="inline-flex items-center gap-1.5">
              <ShortcutHint keys={["E"]} /> Edit
            </span>
            <span className="inline-flex items-center gap-1.5">
              <ShortcutHint keys={["Esc"]} /> Clear
            </span>
          </div>
        </TabsContent>
      </Tabs>

      {/* Route-less detail drawer (`?task=<id>`); `E` opens it in edit mode. */}
      <TodoDetailDrawer />
    </div>
  );
}

// Re-export query type for filter components that build sort options.
export type { TGetTodosQuery };
