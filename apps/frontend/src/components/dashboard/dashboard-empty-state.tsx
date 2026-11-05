import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

interface DashboardEmptyStateProps {
  icon: LucideIcon;
  title: string;
  description: string;
  /** Optional CTA (e.g. a create-task or view-all button). */
  action?: ReactNode;
  className?: string;
}

/**
 * Contextual empty state for dashboard lanes — each lane explains its own
 * empty meaning instead of sharing one generic "no data" block.
 */
export function DashboardEmptyState({
  icon: Icon,
  title,
  description,
  action,
  className,
}: DashboardEmptyStateProps) {
  return (
    <div
      role="status"
      className={cn("px-4 py-6 text-center text-muted-foreground", className)}
    >
      <Icon className="mx-auto mb-2 h-10 w-10 opacity-50" aria-hidden="true" />
      <p className="text-sm font-medium text-foreground">{title}</p>
      <p className="mx-auto mt-1 max-w-60 text-sm">{description}</p>
      {action ? <div className="mt-3">{action}</div> : null}
    </div>
  );
}
