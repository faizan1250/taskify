import "./index.css";
import { AppLayout } from "@/components/layouts/app-layout";
import { AuthLayout } from "@/components/layouts/auth-layout";
import { ProtectedRoute } from "@/components/protected-route";
import { PublicRoute } from "@/components/public-route";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Route,
  createBrowserRouter,
  createRoutesFromElements,
} from "react-router-dom";
import { Suspense, lazy } from "react";

const CategoriesPage = lazy(() =>
  import("@/pages/categories-page").then((m) => ({ default: m.CategoriesPage })),
);
const DashboardPage = lazy(() =>
  import("@/pages/dashboard-page").then((m) => ({ default: m.DashboardPage })),
);
const LandingPage = lazy(() =>
  import("@/pages/landing-page").then((m) => ({ default: m.LandingPage })),
);
const NotFoundPage = lazy(() =>
  import("@/pages/not-found-page").then((m) => ({ default: m.NotFoundPage })),
);
const SettingsPage = lazy(() =>
  import("@/pages/settings-page").then((m) => ({ default: m.SettingsPage })),
);
const TodosPage = lazy(() =>
  import("@/pages/todos-page").then((m) => ({ default: m.TodosPage })),
);

function PageFallback() {
  return (
    <div className="space-y-4 p-4" aria-busy="true" aria-label="Loading page">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-32 w-full" />
      <Skeleton className="h-24 w-full" />
    </div>
  );
}

function withSuspense(element: React.ReactNode) {
  return <Suspense fallback={<PageFallback />}>{element}</Suspense>;
}

const routes = createRoutesFromElements(
  <>
    <Route
      path="/"
      element={withSuspense(
        <PublicRoute>
          <LandingPage />
        </PublicRoute>,
      )}
    />
    <Route
      path="/auth/*"
      element={
        <PublicRoute>
          <AuthLayout />
        </PublicRoute>
      }
    />
    <Route
      path="/dashboard"
      element={withSuspense(
        <ProtectedRoute>
          <AppLayout>
            <DashboardPage />
          </AppLayout>
        </ProtectedRoute>,
      )}
    />
    <Route
      path="/todos"
      element={withSuspense(
        <ProtectedRoute>
          <AppLayout>
            <TodosPage />
          </AppLayout>
        </ProtectedRoute>,
      )}
    />
    <Route
      path="/categories"
      element={withSuspense(
        <ProtectedRoute>
          <AppLayout>
            <CategoriesPage />
          </AppLayout>
        </ProtectedRoute>,
      )}
    />
    <Route
      path="/settings"
      element={withSuspense(
        <ProtectedRoute>
          <AppLayout>
            <SettingsPage />
          </AppLayout>
        </ProtectedRoute>,
      )}
    />
    <Route path="*" element={withSuspense(<NotFoundPage />)} />
  </>,
);

const router = createBrowserRouter(routes);

export default router;
