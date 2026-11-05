import { cn } from "@/lib/utils";

/**
 * Small platform-style kbd chip used for shortcut hints in the topbar,
 * the command palette footer, and the shortcuts dialog.
 */
export function ShortcutHint({
  keys,
  className,
}: {
  keys: readonly string[];
  className?: string;
}) {
  return (
    <span
      className={cn("inline-flex items-center gap-1", className)}
      aria-hidden="true"
    >
      {keys.map((key) => (
        <kbd
          key={key}
          className="rounded-[5px] border border-border bg-muted/70 px-1.5 py-0.5 font-sans text-[10px] font-semibold leading-none text-muted-foreground shadow-[0_1px_0_rgba(0,0,0,0.15)]"
        >
          {key}
        </kbd>
      ))}
    </span>
  );
}
