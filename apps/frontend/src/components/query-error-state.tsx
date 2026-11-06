import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { WifiOff } from "lucide-react";

interface QueryErrorStateProps {
  title: string;
  /** Reassurance + what is preserved (filters, lanes update automatically). */
  description?: string;
  onRetry: () => void;
  isRetrying?: boolean;
  /**
   * default: full-width block (task list). compact: tighter dashboard-lane
   * block. banner: inline strip above stale data.
   */
  variant?: "default" | "compact" | "banner";
  className?: string;
}

/**
 * Honest query-failure state with retry. Rendered instead of an empty state
 * when a list/lane query errors — an outage must never read as "no data".
 * Visual language matches the R2/R3 states (centered muted icon, semibold
 * title, outline retry); the banner variant is destructive-tinted so stale
 * data underneath is clearly flagged.
 */
export function QueryErrorState({
  title,
  description,
  onRetry,
  isRetrying = false,
  variant = "default",
  className,
}: QueryErrorStateProps) {
  if (variant === "banner") {
    return (
      <div
        role="alert"
        className={cn(
          "flex flex-wrap items-center justify-between gap-2 rounded-[7px] border border-destructive/30 bg-destructive/5 px-3 py-2",
          className,
        )}
      >
        <p className="flex min-w-0 items-center gap-2 text-sm">
          <WifiOff className="h-4 w-4 shrink-0 text-destructive" aria-hidden="true" />
          <span className="truncate font-medium">{title}</span>
        </p>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onRetry}
          disabled={isRetrying}
        >
          {isRetrying ? "Retrying…" : "Retry"}
        </Button>
      </div>
    );
  }

  const compact = variant === "compact";
  return (
    <div
      role="alert"
      className={cn(
        "text-center text-muted-foreground",
        compact ? "px-4 py-4" : "px-4 py-8",
        className,
      )}
    >
      <WifiOff
        className={cn("mx-auto mb-2 opacity-50", compact ? "h-8 w-8" : "h-12 w-12")}
        aria-hidden="true"
      />
      <p className="font-medium text-foreground">{title}</p>
      {description ? (
        <p className="mx-auto mt-1 max-w-60 text-sm">{description}</p>
      ) : null}
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="mt-4"
        onClick={onRetry}
        disabled={isRetrying}
      >
        {isRetrying ? "Retrying…" : "Retry"}
      </Button>
    </div>
  );
}
