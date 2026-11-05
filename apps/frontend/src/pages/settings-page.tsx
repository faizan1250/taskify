import { Link } from "react-router-dom";
import { useGetAllCategories } from "@/api/hooks/use-category-query";
import { useGetTodoStats } from "@/api/hooks/use-todo-query";
import { useTheme } from "@/components/theme-provider";
import { ThemeToggle } from "@/components/theme-toggle";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { useClerk, useUser } from "@clerk/clerk-react";
import {
  User,
  Palette,
  Bell,
  Shield,
  Database,
  ExternalLink,
} from "lucide-react";

const NOTIFICATION_ROWS = [
  {
    title: "Task Reminders",
    description: "Get notified about upcoming due dates",
  },
  {
    title: "Email Notifications",
    description: "Receive email updates about your tasks",
  },
  {
    title: "Browser Notifications",
    description: "Show desktop notifications",
  },
];

export function SettingsPage() {
  const { user } = useUser();
  const { openUserProfile } = useClerk();
  const { theme } = useTheme();

  const {
    data: stats,
    isLoading: statsLoading,
    isError: statsError,
  } = useGetTodoStats();
  const {
    data: categories,
    isLoading: categoriesLoading,
    isError: categoriesError,
  } = useGetAllCategories({ query: { page: 1, limit: 1 } });

  const themeName =
    theme === "system" ? "System" : theme === "dark" ? "Dark" : "Light";

  const statsLoading_ = statsLoading || categoriesLoading;
  const statsFailed = statsError || categoriesError;
  const categoryCount = categories?.total ?? categories?.data?.length ?? 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground">
          Manage your account and application preferences
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Profile Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              Profile
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-4">
              <img
                src={user?.imageUrl}
                alt={user?.fullName || "User"}
                className="w-16 h-16 rounded-full"
              />
              <div>
                <h3 className="font-semibold">{user?.fullName}</h3>
                <p className="text-sm text-muted-foreground">
                  {user?.primaryEmailAddress?.emailAddress}
                </p>
                <Badge variant="secondary" className="mt-1">
                  {(user?.publicMetadata?.role as string) || "User"}
                </Badge>
              </div>
            </div>
            <Separator />
            <div className="space-y-2">
              <p className="text-sm font-medium">Account Details</p>
              <div className="space-y-1 text-sm text-muted-foreground">
                <p>
                  Member since:{" "}
                  {user?.createdAt
                    ? new Date(user.createdAt).toLocaleDateString()
                    : "N/A"}
                </p>
                <p>
                  Last updated:{" "}
                  {user?.updatedAt
                    ? new Date(user.updatedAt).toLocaleDateString()
                    : "N/A"}
                </p>
              </div>
            </div>
            <Button
              variant="outline"
              className="w-full"
              onClick={() => openUserProfile()}
            >
              <ExternalLink className="h-4 w-4 mr-2" />
              Manage Account in Clerk
            </Button>
          </CardContent>
        </Card>

        {/* Appearance Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Palette className="h-5 w-5" />
              Appearance
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Theme</p>
                <p className="text-sm text-muted-foreground">
                  Current theme: {themeName}
                </p>
              </div>
              <ThemeToggle />
            </div>
            <Separator />
            <div className="space-y-2">
              <p className="text-sm font-medium">Theme Information</p>
              <div className="space-y-1 text-sm text-muted-foreground">
                <p>• Light theme for daytime use</p>
                <p>• Dark theme for low-light environments</p>
                <p>• System theme follows your OS preference</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              Notifications
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              {NOTIFICATION_ROWS.map((row) => (
                <div key={row.title} className="flex items-center justify-between gap-4">
                  <div>
                    <p className="font-medium">{row.title}</p>
                    <p className="text-sm text-muted-foreground">
                      {row.description}
                    </p>
                  </div>
                  <Switch disabled checked={false} aria-label={row.title} />
                </div>
              ))}
            </div>
            <p className="text-xs text-muted-foreground">
              Notifications are not available in this version — these controls
              are shown disabled so nothing looks actionable that isn&apos;t.
            </p>
          </CardContent>
        </Card>

        {/* Privacy & Security */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              Privacy & Security
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <div>
                <p className="font-medium">Data Encryption</p>
                <p className="text-sm text-muted-foreground">
                  Your data is encrypted both in transit and at rest
                </p>
              </div>
              <div>
                <p className="font-medium">Authentication</p>
                <p className="text-sm text-muted-foreground">
                  Powered by Clerk with industry-standard security
                </p>
              </div>
              <div>
                <p className="font-medium">Data Retention</p>
                <p className="text-sm text-muted-foreground">
                  Your data is stored securely and can be deleted upon request
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Data Management */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              Data Management
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {statsLoading_ ? (
              <div className="grid gap-4 md:grid-cols-3">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} className="h-24 w-full" />
                ))}
              </div>
            ) : statsFailed ? (
              <p className="text-sm text-muted-foreground text-center py-4" role="alert">
                Couldn&apos;t load usage stats. Refresh the page to try again.
              </p>
            ) : (
              <div className="grid gap-4 md:grid-cols-3">
                <div className="text-center p-4 border rounded-lg">
                  <p className="text-2xl font-bold text-blue-600">
                    {stats?.total ?? 0}
                  </p>
                  <p className="text-sm font-medium">Total Tasks</p>
                  <p className="text-xs text-muted-foreground">All time</p>
                </div>
                <div className="text-center p-4 border rounded-lg">
                  <p className="text-2xl font-bold text-green-600">
                    {stats?.completed ?? 0}
                  </p>
                  <p className="text-sm font-medium">Completed Tasks</p>
                  <p className="text-xs text-muted-foreground">All time</p>
                </div>
                <div className="text-center p-4 border rounded-lg">
                  <p className="text-2xl font-bold text-purple-600">
                    {categoryCount}
                  </p>
                  <p className="text-sm font-medium">Categories</p>
                  <p className="text-xs text-muted-foreground">Total created</p>
                </div>
              </div>
            )}
            <Separator />
            <p className="text-xs text-muted-foreground text-center">
              Bulk export, import, and delete-all are not available in this
              version. Manage individual tasks from the{" "}
              <Link to="/todos" className="underline underline-offset-2">
                Todos page
              </Link>
              .
            </p>
          </CardContent>
        </Card>
      </div>

      {/* App Information */}
      <Card>
        <CardHeader>
          <CardTitle>About Taskify</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="font-medium">Version</p>
              <p className="text-sm text-muted-foreground">1.0.0</p>
            </div>
            <div>
              <p className="font-medium">Last Updated</p>
              <p className="text-sm text-muted-foreground">
                {new Date().toLocaleDateString()}
              </p>
            </div>
            <div>
              <p className="font-medium">Built With</p>
              <p className="text-sm text-muted-foreground">
                React, TypeScript, Tailwind CSS
              </p>
            </div>
            <div>
              <p className="font-medium">Support</p>
              <p className="text-sm text-muted-foreground">
                Contact us for help and feedback
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
