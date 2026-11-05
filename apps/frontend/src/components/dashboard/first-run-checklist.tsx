import { useState } from "react";
import { Link } from "react-router-dom";
import { Check, ListChecks, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

const STORAGE_KEY = "taskify:dashboard:first-run-dismissed";

interface FirstRunChecklistProps {
  hasTask: boolean;
  hasCategory: boolean;
  hasCompleted: boolean;
}

interface ChecklistStep {
  key: string;
  label: string;
  hint: string;
  done: boolean;
  to: string;
  cta: string;
}

/**
 * First-run onboarding checklist. Dismissal persists in localStorage; the
 * card also auto-hides once every step is complete.
 */
export function FirstRunChecklist({
  hasTask,
  hasCategory,
  hasCompleted,
}: FirstRunChecklistProps) {
  const [dismissed, setDismissed] = useState<boolean>(() => {
    try {
      return window.localStorage.getItem(STORAGE_KEY) === "1";
    } catch {
      return false;
    }
  });

  if (dismissed) return null;

  const steps: ChecklistStep[] = [
    {
      key: "task",
      label: "Create your first task",
      hint: "Capture anything on your mind — you can refine it later.",
      done: hasTask,
      to: "/todos",
      cta: "Go to tasks",
    },
    {
      key: "category",
      label: "Organize with a category",
      hint: "Group related work so filters and lanes stay useful.",
      done: hasCategory,
      to: "/categories",
      cta: "Manage categories",
    },
    {
      key: "complete",
      label: "Complete a task",
      hint: "Tick the checkbox on any task to feel the streak.",
      done: hasCompleted,
      to: "/todos?tab=active",
      cta: "View active tasks",
    },
  ];

  const doneCount = steps.filter((step) => step.done).length;
  if (doneCount === steps.length) return null;

  const dismiss = () => {
    try {
      window.localStorage.setItem(STORAGE_KEY, "1");
    } catch {
      // Storage unavailable (private mode) — hide for this session only.
    }
    setDismissed(true);
  };

  return (
    <Card aria-labelledby="first-run-checklist-title">
      <CardHeader className="flex flex-row items-center justify-between gap-2 pb-2">
        <div className="flex items-center gap-2">
          <ListChecks className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <CardTitle id="first-run-checklist-title" className="text-base">
            Getting started
          </CardTitle>
        </div>
        <div className="flex items-center gap-2">
          <span
            role="status"
            aria-live="polite"
            className="text-xs text-muted-foreground"
          >
            {doneCount} of {steps.length} done
          </span>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={dismiss}
            aria-label="Dismiss getting started checklist"
            className="dash-focusable h-8 w-8"
          >
            <X className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <ol className="space-y-2">
          {steps.map((step) => (
            <li
              key={step.key}
              className="flex items-start justify-between gap-3 rounded-[7px] border p-3"
            >
              <div className="flex items-start gap-3">
                <span
                  aria-hidden="true"
                  className={cn(
                    "mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border",
                    step.done
                      ? "border-green-500/40 bg-green-500/15"
                      : "border-border bg-muted",
                  )}
                >
                  {step.done ? (
                    <Check className="h-3 w-3 text-green-600 dark:text-green-400" />
                  ) : null}
                </span>
                <div>
                  <p className="text-sm font-medium">
                    {step.label}
                    <span className="dash-sr-only">
                      {step.done ? " — completed" : " — not started"}
                    </span>
                  </p>
                  <p className="text-xs text-muted-foreground">{step.hint}</p>
                </div>
              </div>
              {!step.done && (
                <Link
                  to={step.to}
                  className="dash-focusable shrink-0 rounded-[7px] border px-2 py-1 text-xs font-medium transition-colors hover:border-ring hover:text-foreground"
                >
                  {step.cta}
                </Link>
              )}
            </li>
          ))}
        </ol>
      </CardContent>
    </Card>
  );
}
