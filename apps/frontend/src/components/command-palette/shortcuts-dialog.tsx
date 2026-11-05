import { ShortcutHint } from "@/components/command-palette/shortcut-hint";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { SHORTCUT_DEFS, displayKeys } from "@/lib/shortcuts";

/** `?` dialog: every global shortcut with its keys and description. */
export function ShortcutsDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Keyboard shortcuts</DialogTitle>
          <DialogDescription>
            Global shortcuts that work anywhere in the app.
          </DialogDescription>
        </DialogHeader>
        <ul className="space-y-1">
          {SHORTCUT_DEFS.map((shortcut) => (
            <li
              key={shortcut.id}
              className="flex items-center justify-between gap-4 rounded-[7px] px-2 py-2 text-sm"
            >
              <span>
                <span className="font-medium">{shortcut.label}</span>
                <span className="block text-xs text-muted-foreground">
                  {shortcut.description}
                </span>
              </span>
              <ShortcutHint keys={displayKeys(shortcut.keys)} />
            </li>
          ))}
        </ul>
        <p className="border-t pt-3 text-xs text-muted-foreground">
          Single-key shortcuts are ignored while typing. Press{" "}
          <ShortcutHint keys={displayKeys(["?"])} className="inline-flex" />{" "}
          anytime to reopen this dialog.
        </p>
      </DialogContent>
    </Dialog>
  );
}
