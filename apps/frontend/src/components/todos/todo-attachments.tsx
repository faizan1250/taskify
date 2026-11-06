import { useState } from "react";
import type { TGetTodoByIdResponse } from "@/api/hooks";
import {
  useDeleteTodoAttachment,
  useGetTodoAttachmentPresignedUrl,
} from "@/api/hooks/use-todo-attachment-query";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { formatFileSize } from "@/lib/utils";
import { Download, FileText, Loader2, Trash2 } from "lucide-react";
import { toast } from "sonner";

interface TodoAttachmentsProps {
  todoId: string;
  attachments: TGetTodoByIdResponse["attachments"];
  disabled?: boolean;
}

type PendingDelete = {
  id: string;
  name: string;
} | null;

export function TodoAttachments({
  todoId,
  attachments,
  disabled = false,
}: TodoAttachmentsProps) {
  const deleteAttachment = useDeleteTodoAttachment();
  const getPresignedUrl = useGetTodoAttachmentPresignedUrl();
  const [pendingDelete, setPendingDelete] = useState<PendingDelete>(null);
  // Per-row pending ids — other rows stay interactive while one works.
  const [downloadingId, setDownloadingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const handleDownload = async (
    attachment: TGetTodoByIdResponse["attachments"][number],
  ) => {
    setDownloadingId(attachment.id);
    try {
      const result = await getPresignedUrl.mutateAsync({
        todoId,
        attachmentId: attachment.id,
      });

      // `window.open` returns null when the popup is blocked — say so
      // instead of silently swallowing the download.
      const win = window.open(result.url, "_blank", "noopener");
      if (!win) {
        toast.error(
          "Popup blocked — allow popups for this site, then try again.",
        );
      }
    } catch (error) {
      console.error("Failed to download attachment:", error);
    } finally {
      setDownloadingId((current) =>
        current === attachment.id ? null : current,
      );
    }
  };

  const handleConfirmDelete = async () => {
    if (!pendingDelete) return;

    setDeletingId(pendingDelete.id);
    try {
      await deleteAttachment.mutateAsync({
        todoId,
        attachmentId: pendingDelete.id,
      });
      toast.success("Attachment deleted successfully");
      setPendingDelete(null);
    } catch (error) {
      console.error("Failed to delete attachment:", error);
    } finally {
      setDeletingId(null);
    }
  };

  if (attachments.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No attachments yet. Uploaded files will appear here.
      </p>
    );
  }

  return (
    <>
      <div className="space-y-2">
        <h4 className="text-sm font-medium">
          Current Attachments ({attachments.length})
        </h4>
        <ul
          className="space-y-2 max-h-48 overflow-y-auto"
          aria-label="Attachments"
        >
          {attachments.map((attachment) => {
            const isDownloading = downloadingId === attachment.id;
            const isDeleting = deletingId === attachment.id;
            const rowBusy = isDownloading || isDeleting;
            return (
              <li
                key={attachment.id}
                className="flex items-center justify-between p-3 bg-muted/50 rounded-lg hover:bg-muted transition-colors"
              >
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <FileText
                    className="h-5 w-5 text-muted-foreground flex-shrink-0"
                    aria-hidden="true"
                  />
                  <div className="min-w-0 flex-1">
                    <p
                      className="text-sm font-medium truncate"
                      title={attachment.name}
                    >
                      {attachment.name}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {attachment.fileSize
                        ? formatFileSize(attachment.fileSize)
                        : "Unknown size"}
                      {attachment.mimeType &&
                        ` • ${attachment.mimeType.split("/")[1].toUpperCase()}`}
                      {isDownloading && " • Preparing download…"}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDownload(attachment)}
                    disabled={disabled || rowBusy}
                    title="Download attachment"
                    aria-label={`Download attachment ${attachment.name}`}
                  >
                    {isDownloading ? (
                      <Loader2
                        className="h-4 w-4 animate-spin"
                        aria-hidden="true"
                      />
                    ) : (
                      <Download className="h-4 w-4" aria-hidden="true" />
                    )}
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() =>
                      setPendingDelete({
                        id: attachment.id,
                        name: attachment.name,
                      })
                    }
                    disabled={disabled || rowBusy}
                    className="text-destructive hover:text-destructive"
                    title="Delete attachment"
                    aria-label={`Delete attachment ${attachment.name}`}
                  >
                    {isDeleting ? (
                      <Loader2
                        className="h-4 w-4 animate-spin"
                        aria-hidden="true"
                      />
                    ) : (
                      <Trash2 className="h-4 w-4" aria-hidden="true" />
                    )}
                  </Button>
                </div>
              </li>
            );
          })}
        </ul>
      </div>

      <AlertDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open && deletingId === null) setPendingDelete(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Attachment</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{pendingDelete?.name}&quot;?
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deletingId !== null}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                void handleConfirmDelete();
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              disabled={deletingId !== null}
            >
              {deletingId !== null ? "Deleting…" : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
