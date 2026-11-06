import {
  CommandPalette,
  ShortcutHint,
  ShortcutsDialog,
} from "@/components/command-palette";
import { ErrorBoundary } from "@/components/error-boundary";
import { useTheme } from "@/components/theme-provider";
import { ThemeToggle } from "@/components/theme-toggle";
import { TodoCreateForm } from "@/components/todos/todo-create-form";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import {
  displayKeys,
  openNewTaskDialog,
  resolveThemeToggle,
  usePaletteHotkeys,
} from "@/lib/shortcuts";
import { cn } from "@/lib/utils";
import { UserButton } from "@clerk/clerk-react";
import {
  LayoutDashboard,
  CheckSquare,
  Folder,
  Settings,
  Menu,
  Search,
  Plus,
  Keyboard,
} from "lucide-react";
import { type ReactNode, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { toast } from "sonner";

const navigation = [
  { name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { name: "Todos", href: "/todos", icon: CheckSquare },
  { name: "Categories", href: "/categories", icon: Folder },
  { name: "Settings", href: "/settings", icon: Settings },
];

interface AppLayoutProps {
  children: ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const location = useLocation();
  const { setTheme } = useTheme();
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  usePaletteHotkeys({
    onTogglePalette: () => setPaletteOpen((value) => !value),
    onOpenPalette: () => setPaletteOpen(true),
    onOpenShortcuts: () => setShortcutsOpen(true),
    onNewTask: () => {
      setPaletteOpen(false);
      requestAnimationFrame(() => {
        if (!openNewTaskDialog()) {
          toast.error("Could not open the New Task dialog");
        }
      });
    },
    onToggleTheme: () => setTheme(resolveThemeToggle()),
  });

  const NavigationItems = ({ onNavigate }: { onNavigate?: () => void }) => (
    <>
      {navigation.map((item) => {
        const Icon = item.icon;
        const isActive = location.pathname === item.href;

        return (
          <Link
            key={item.name}
            to={item.href}
            aria-current={isActive ? "page" : undefined}
            onClick={onNavigate}
          >
            <Button
              variant={isActive ? "secondary" : "ghost"}
              className={cn(
                "w-full justify-start gap-2",
                isActive && "bg-secondary",
              )}
            >
              <Icon className="h-4 w-4" />
              {item.name}
            </Button>
          </Link>
        );
      })}
    </>
  );

  return (
    <div className="min-h-screen bg-background">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-[7px] focus:bg-primary focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-foreground focus:outline-none"
      >
        Skip to main content
      </a>
      {/* Desktop Sidebar */}
      <div className="hidden md:fixed md:inset-y-0 md:flex md:w-64 md:flex-col">
        <div className="flex flex-col flex-grow pt-5 bg-card border-r overflow-y-auto">
          <div className="flex items-center flex-shrink-0 px-4">
            <h1 className="text-xl font-bold">Taskify</h1>
          </div>
          <div className="mt-8 flex-grow flex flex-col">
            <nav className="flex-1 px-4 space-y-1" aria-label="Primary">
              <NavigationItems />
            </nav>
          </div>
        </div>
      </div>

      {/* Mobile Navigation */}
      <div className="md:hidden">
        <div className="bg-card border-b px-4 py-3 flex items-center justify-between">
          <h1 className="text-xl font-bold">Taskify</h1>
          <Sheet open={mobileNavOpen} onOpenChange={setMobileNavOpen}>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Open navigation menu"
                aria-expanded={mobileNavOpen}
              >
                <Menu className="h-5 w-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-64">
              <div className="py-4">
                <h1 className="text-xl font-bold mb-8">Taskify</h1>
                <nav className="space-y-1" aria-label="Mobile">
                  <NavigationItems onNavigate={() => setMobileNavOpen(false)} />
                </nav>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>

      {/* Main Content */}
      <div className="md:pl-64">
        {/* Top Bar */}
        <div className="sticky top-0 z-40 flex h-16 shrink-0 items-center gap-x-4 border-b bg-card px-4 sm:gap-x-6 sm:px-6 lg:px-8">
          <div className="flex flex-1 items-center gap-x-2 self-stretch py-3 sm:gap-x-3">
            <Button
              variant="outline"
              className="h-9 min-w-9 justify-start gap-2 px-3 font-normal text-muted-foreground sm:w-full sm:max-w-md"
              onClick={() => setPaletteOpen(true)}
              aria-label="Search tasks and commands"
            >
              <Search className="h-4 w-4 shrink-0" />
              <span className="hidden truncate sm:inline">
                Search tasks, pages, actions…
              </span>
              <span className="ml-auto hidden sm:inline-flex">
                <ShortcutHint keys={displayKeys(["mod", "K"])} />
              </span>
            </Button>
            <div className="ml-auto flex items-center gap-x-1 sm:gap-x-2">
              <TodoCreateForm>
                <Button size="sm" data-new-task-trigger aria-label="New task">
                  <Plus className="h-4 w-4" />
                  <span className="hidden md:inline">New Task</span>
                </Button>
              </TodoCreateForm>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Keyboard shortcuts"
                onClick={() => setShortcutsOpen(true)}
                className="hidden sm:inline-flex"
              >
                <Keyboard className="h-5 w-5" />
              </Button>
              <ThemeToggle />
              <UserButton
                appearance={{
                  elements: {
                    avatarBox: "h-8 w-8",
                  },
                }}
              />
            </div>
          </div>
        </div>

        {/* Page Content */}
        <main id="main-content" tabIndex={-1} className="py-6 outline-none">
          <div className="px-4 sm:px-6 lg:px-8">
            <ErrorBoundary fallbackTitle="Page failed to load">
              {children}
            </ErrorBoundary>
          </div>
        </main>
      </div>

      <CommandPalette
        open={paletteOpen}
        onOpenChange={setPaletteOpen}
        onOpenShortcuts={() => setShortcutsOpen(true)}
      />
      <ShortcutsDialog open={shortcutsOpen} onOpenChange={setShortcutsOpen} />
    </div>
  );
}
