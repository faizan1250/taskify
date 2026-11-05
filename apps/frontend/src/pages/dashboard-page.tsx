import { useMemo } from "react";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useGetTodoStats, useGetAllTodos } from "@/api/hooks/use-todo-query";
import { useGetAllCategories } from "@/api/hooks/use-category-query";
import { TodoCard } from "@/components/todos/todo-card";
import { TodoCreateForm } from "@/components/todos/todo-create-form";
import { TodoDetailDrawer } from "@/components/todos/todo-detail-drawer";
import { DashboardEmptyState } from "@/components/dashboard/dashboard-empty-state";
import { FirstRunChecklist } from "@/components/dashboard/first-run-checklist";
import { toDueBoundISO } from "@/hooks/use-todos-search-params";
import {
  AlertTriangle,
  Archive,
  CalendarDays,
  CheckCircle,
  Clock,
  FolderOpen,
  Inbox,
  Plus,
  ChevronRight,
} from "lucide-react";

function toDateKey(date: Date): string {
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}

export function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useGetTodoStats();
  const { data: recentTodos, isLoading: todosLoading } = useGetAllTodos({
    query: { page: 1, limit: 5, sort: "updated_at", order: "desc" },
  });
  const { data: categories, isLoading: categoriesLoading } = useGetAllCategories({
    query: { page: 1, limit: 10 },
  });

  // Due-this-week lane: today → +7 days, shared with the todos filter URL.
  const weekRange = useMemo(() => {
    const from = new Date();
    const to = new Date();
    to.setDate(to.getDate() + 7);
    return { dueFrom: toDateKey(from), dueTo: toDateKey(to) };
  }, []);

  const { data: weekTodos, isLoading: weekLoading } = useGetAllTodos({
    query: {
      page: 1,
      limit: 5,
      sort: "due_date",
      order: "asc",
      dueFrom: toDueBoundISO(weekRange.dueFrom, "from"),
      dueTo: toDueBoundISO(weekRange.dueTo, "to"),
    },
  });

  const { data: overdueTodos, isLoading: overdueLoading } = useGetAllTodos({
    query: { page: 1, limit: 5, sort: "due_date", order: "asc", overdue: true },
  });

  // Category counts: tallied client-side from a wider recent slice so each
  // category row can link to its pre-filtered task list.
  const { data: countSample } = useGetAllTodos({
    query: { page: 1, limit: 100, sort: "updated_at", order: "desc" },
  });
  const countsByCategory = useMemo(() => {
    const counts = new Map<string, number>();
    for (const todo of countSample?.data ?? []) {
      if (!todo.categoryId) continue;
      counts.set(todo.categoryId, (counts.get(todo.categoryId) ?? 0) + 1);
    }
    return counts;
  }, [countSample]);

  const statCards = [
    {
      title: "Total Tasks",
      value: stats?.total || 0,
      icon: CheckCircle,
      color: "text-blue-600 dark:text-blue-400",
      bgColor:
        "border border-blue-500/30 bg-gradient-to-b from-blue-500/15 to-blue-500/5",
      to: "/todos",
      filterLabel: "View all tasks",
    },
    {
      title: "Active",
      value: stats?.active || 0,
      icon: Clock,
      color: "text-green-600 dark:text-green-400",
      bgColor:
        "border border-green-500/30 bg-gradient-to-b from-green-500/15 to-green-500/5",
      to: "/todos?tab=active",
      filterLabel: "View active tasks",
    },
    {
      title: "Overdue",
      value: stats?.overdue || 0,
      icon: AlertTriangle,
      color: "text-red-600 dark:text-red-400",
      bgColor:
        "border border-red-500/30 bg-gradient-to-b from-red-500/15 to-red-500/5",
      to: "/todos?overdue=true&sort=due_date&order=asc",
      filterLabel: "View overdue tasks",
    },
    {
      title: "Completed",
      value: stats?.completed || 0,
      icon: Archive,
      color: "text-gray-600 dark:text-gray-400",
      bgColor:
        "border border-gray-500/30 bg-gradient-to-b from-gray-500/15 to-gray-500/5",
      to: "/todos?tab=completed",
      filterLabel: "View completed tasks",
    },
  ];

  const weekHref = `/todos?dueFrom=${weekRange.dueFrom}&dueTo=${weekRange.dueTo}&sort=due_date&order=asc`;
  const overdueHref = "/todos?overdue=true&sort=due_date&order=asc";

  return (
    <div className="dash-scope space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground">
            Overview of your tasks and productivity
          </p>
        </div>
        <TodoCreateForm>
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            New Task
          </Button>
        </TodoCreateForm>
      </div>

      {/* First-run onboarding checklist (dismissible, localStorage). */}
      {!statsLoading && (
        <FirstRunChecklist
          hasTask={(stats?.total ?? 0) > 0}
          hasCategory={(categories?.total ?? 0) > 0}
          hasCompleted={(stats?.completed ?? 0) > 0}
        />
      )}

      {/* Stats Cards — each links to its filtered task list */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat) => {
          const Icon = stat.icon;
          return (
            <Link
              key={stat.title}
              to={stat.to}
              aria-label={`${stat.title}: ${stat.value}. ${stat.filterLabel}`}
              className="dash-focusable rounded-[7px] transition-colors hover:border-ring"
            >
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">
                    {stat.title}
                  </CardTitle>
                  <div className={`p-2 rounded-full ${stat.bgColor}`}>
                    <Icon className={`h-4 w-4 ${stat.color}`} aria-hidden="true" />
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {statsLoading ? (
                      <Skeleton className="h-8 w-12" />
                    ) : (
                      stat.value
                    )}
                  </div>
                  <span className="dash-sr-only">{stat.filterLabel}</span>
                </CardContent>
              </Card>
            </Link>
          );
        })}
      </div>

      {/* Attention lanes: overdue + due this week */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card aria-labelledby="dash-overdue-title">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle id="dash-overdue-title" className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-red-500" aria-hidden="true" />
              Overdue
              {!overdueLoading && (overdueTodos?.total ?? 0) > 0 && (
                <span
                  role="status"
                  className="rounded-full bg-destructive/10 px-2 py-0.5 text-xs font-medium text-destructive"
                >
                  {overdueTodos?.total}
                </span>
              )}
            </CardTitle>
            <Link to={overdueHref}>
              <Button variant="ghost" size="sm" className="dash-focusable">
                View All
                <ChevronRight className="h-4 w-4 ml-1" aria-hidden="true" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent className="space-y-3">
            {overdueLoading ? (
              <div className="space-y-3" aria-busy="true" aria-label="Loading overdue tasks">
                {Array.from({ length: 2 }).map((_, i) => (
                  <div key={i} className="space-y-2">
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-3 w-2/3" />
                  </div>
                ))}
              </div>
            ) : overdueTodos?.data?.length ? (
              <div className="space-y-3">
                {overdueTodos.data.slice(0, 5).map((todo) => (
                  <div key={todo.id} className="dash-item border border-transparent">
                    <TodoCard todo={todo} compact />
                  </div>
                ))}
              </div>
            ) : (
              <DashboardEmptyState
                icon={CheckCircle}
                title="Nothing overdue"
                description="Every task is on track. New due dates will surface here automatically."
              />
            )}
          </CardContent>
        </Card>

        <Card aria-labelledby="dash-week-title">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle id="dash-week-title" className="flex items-center gap-2">
              <CalendarDays className="h-4 w-4 text-blue-500" aria-hidden="true" />
              Due this week
            </CardTitle>
            <Link to={weekHref}>
              <Button variant="ghost" size="sm" className="dash-focusable">
                View All
                <ChevronRight className="h-4 w-4 ml-1" aria-hidden="true" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent className="space-y-3">
            {weekLoading ? (
              <div className="space-y-3" aria-busy="true" aria-label="Loading tasks due this week">
                {Array.from({ length: 2 }).map((_, i) => (
                  <div key={i} className="space-y-2">
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-3 w-2/3" />
                  </div>
                ))}
              </div>
            ) : weekTodos?.data?.length ? (
              <div className="space-y-3">
                {weekTodos.data.slice(0, 5).map((todo) => (
                  <div key={todo.id} className="dash-item border border-transparent">
                    <TodoCard todo={todo} compact />
                  </div>
                ))}
              </div>
            ) : (
              <DashboardEmptyState
                icon={CalendarDays}
                title="A clear week ahead"
                description="No due dates in the next 7 days. Assign a due date to see it here."
                action={
                  <Link to="/todos?tab=active">
                    <Button variant="outline" size="sm" className="dash-focusable">
                      Review active tasks
                    </Button>
                  </Link>
                }
              />
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Recent Tasks */}
        <Card className="lg:col-span-2" aria-labelledby="dash-recent-title">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle id="dash-recent-title">Recent Tasks</CardTitle>
            <Link to="/todos">
              <Button variant="ghost" size="sm" className="dash-focusable">
                View All
                <ChevronRight className="h-4 w-4 ml-1" aria-hidden="true" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent className="space-y-3">
            {todosLoading ? (
              <div className="space-y-3" aria-busy="true" aria-label="Loading recent tasks">
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="space-y-2">
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-3 w-2/3" />
                  </div>
                ))}
              </div>
            ) : recentTodos?.data?.length ? (
              <div className="space-y-3">
                {recentTodos.data.slice(0, 5).map((todo) => (
                  <div key={todo.id} className="dash-item border border-transparent">
                    <TodoCard todo={todo} compact />
                  </div>
                ))}
              </div>
            ) : (
              <DashboardEmptyState
                icon={Inbox}
                title="No tasks yet"
                description="Create your first task to get started — it will show up here."
                action={
                  <TodoCreateForm>
                    <Button variant="outline" className="dash-focusable">
                      <Plus className="h-4 w-4 mr-2" aria-hidden="true" />
                      Create Task
                    </Button>
                  </TodoCreateForm>
                }
              />
            )}
          </CardContent>
        </Card>

        {/* Categories */}
        <Card aria-labelledby="dash-categories-title">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle id="dash-categories-title">Categories</CardTitle>
            <Link to="/categories">
              <Button variant="ghost" size="sm" className="dash-focusable">
                Manage
                <ChevronRight className="h-4 w-4 ml-1" aria-hidden="true" />
              </Button>
            </Link>
          </CardHeader>
          <CardContent>
            {categoriesLoading ? (
              <div className="space-y-2" aria-busy="true" aria-label="Loading categories">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-8 w-full" />
                ))}
              </div>
            ) : categories?.data?.length ? (
              <div className="space-y-2">
                {categories.data.slice(0, 6).map((category) => {
                  const count = countsByCategory.get(category.id) ?? 0;
                  return (
                    <Link
                      key={category.id}
                      to={`/todos?category=${category.id}`}
                      aria-label={`${category.name}, ${count} recent task${count === 1 ? "" : "s"}. View tasks in this category`}
                      className="dash-focusable flex items-center justify-between rounded-[7px] border p-2 transition-colors hover:border-ring"
                    >
                      <span className="flex items-center gap-2">
                        <span
                          className="w-3 h-3 rounded-full"
                          style={{ backgroundColor: category.color }}
                          aria-hidden="true"
                        />
                        <span className="text-sm font-medium">
                          {category.name}
                        </span>
                      </span>
                      <span
                        role="status"
                        aria-label={`${count} tasks`}
                        className="rounded-full bg-muted px-2 py-0.5 text-xs tabular-nums text-muted-foreground"
                      >
                        {count}
                      </span>
                    </Link>
                  );
                })}
                {categories.data.length > 6 && (
                  <p className="text-xs text-muted-foreground text-center pt-2">
                    +{categories.data.length - 6} more categories
                  </p>
                )}
              </div>
            ) : (
              <DashboardEmptyState
                icon={FolderOpen}
                title="No categories yet"
                description="Categories keep filters and lanes useful as you grow."
                action={
                  <Link to="/categories">
                    <Button variant="outline" size="sm" className="dash-focusable">
                      Create Category
                    </Button>
                  </Link>
                }
              />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Route-less detail drawer (`?task=<id>`) for recent-task cards. */}
      <TodoDetailDrawer />
    </div>
  );
}
