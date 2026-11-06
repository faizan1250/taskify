import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { formatDistanceToNow } from "date-fns";
import { useUser } from "@clerk/clerk-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { ErrorBoundary } from "@/components/error-boundary";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import {
  useGetCommentsByTodoId,
  useAddComment,
  useUpdateComment,
  useDeleteComment,
} from "@/api/hooks/use-comment-query";
import {
  Send,
  MessageSquare,
  MoreHorizontal,
  Edit,
  Trash2,
} from "lucide-react";
import { toast } from "sonner";

const commentSchema = z.object({
  content: z
    .string()
    .trim()
    .min(1, "Comment cannot be empty")
    .max(1000, "Comment too long"),
});

type CommentForm = z.infer<typeof commentSchema>;

interface TodoCommentsPanelProps {
  todoId: string;
}

interface Comment {
  id: string;
  content: string;
  userId: string;
  createdAt: string;
  updatedAt: string;
}

function formatRelativeTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  try {
    return formatDistanceToNow(date, { addSuffix: true });
  } catch {
    return date.toLocaleString();
  }
}

function initialsFor(userId: string): string {
  const clean = userId.replace(/^optimistic-/, "");
  return (clean.charAt(0) || "?").toUpperCase();
}

/**
 * Comment list + add/edit/delete, without any dialog chrome — shared by
 * the comments dialog and the task detail drawer tab.
 */
export function TodoCommentsPanel({ todoId }: TodoCommentsPanelProps) {
  const [editingComment, setEditingComment] = useState<string | null>(null);
  const [deleteCommentId, setDeleteCommentId] = useState<string | null>(null);

  const { user } = useUser();
  const {
    data: comments,
    isLoading,
    isError,
    refetch,
    isFetching,
  } = useGetCommentsByTodoId({ todoId });
  const addComment = useAddComment();
  const updateComment = useUpdateComment();
  const deleteComment = useDeleteComment();

  const form = useForm<CommentForm>({
    resolver: zodResolver(commentSchema),
    defaultValues: {
      content: "",
    },
  });

  const editForm = useForm<CommentForm>({
    resolver: zodResolver(commentSchema),
    defaultValues: {
      content: "",
    },
  });

  const onSubmit = async (data: CommentForm) => {
    if (addComment.isPending) return;
    await addComment.mutateAsync({
      body: {
        todoId,
        content: data.content.trim(),
      },
      userId: user?.id ?? "unknown",
    });

    toast.success("Comment added successfully!");
    form.reset();
  };

  const onEditSubmit = async (data: CommentForm) => {
    if (!editingComment) return;

    try {
      await updateComment.mutateAsync({
        commentId: editingComment,
        body: {
          content: data.content,
        },
      });

      toast.success("Comment updated successfully!");
      setEditingComment(null);
      editForm.reset();
    } catch {
      toast.error("Failed to update comment");
    }
  };

  const handleDeleteComment = async (commentId: string) => {
    try {
      await deleteComment.mutateAsync({ commentId, todoId });
      toast.success("Comment deleted successfully!");
      setDeleteCommentId(null);
    } catch {
      toast.error("Failed to delete comment");
    }
  };

  const startEditing = (comment: Comment) => {
    setEditingComment(comment.id);
    editForm.setValue("content", comment.content);
  };

  const cancelEditing = () => {
    setEditingComment(null);
    editForm.reset();
  };

  return (
    <>
      <div className="flex flex-1 flex-col min-h-0">
        <ErrorBoundary
          fallbackTitle="Comments failed to load"
          fallbackMessage="Could not load comments for this task. Please try again."
        >
          {/* Comments List */}
          <ScrollArea className="flex-1 pr-4">
            {!isLoading && !isError && (
              <p role="status" aria-live="polite" className="sr-only">
                {comments?.length
                  ? `${comments.length} comment${comments.length === 1 ? "" : "s"} on this task`
                  : "No comments on this task"}
              </p>
            )}
            <div className="space-y-4">
              {isLoading ? (
                <div className="space-y-4">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <Card key={i}>
                      <CardContent className="p-4">
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <Skeleton className="h-4 w-24" />
                            <Skeleton className="h-4 w-16" />
                          </div>
                          <Skeleton className="h-12 w-full" />
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : isError ? (
                <div className="text-center py-8">
                  <MessageSquare className="h-12 w-12 mx-auto mb-2 text-muted-foreground opacity-50" />
                  <p className="font-medium">Couldn't load comments</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Check your connection and try again — your comments are safe.
                  </p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="mt-4"
                    onClick={() => void refetch()}
                    disabled={isFetching}
                  >
                    {isFetching ? "Retrying…" : "Retry"}
                  </Button>
                </div>
              ) : comments?.length ? (
                comments.map((comment) => {
                  const isOptimistic = comment.id.startsWith("optimistic-");
                  const isOwn = comment.userId === user?.id;
                  return (
                  <Card key={comment.id} className={isOptimistic ? "opacity-70" : undefined} aria-busy={isOptimistic || undefined}>
                    <CardContent className="p-4">
                      <div className="space-y-3">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2 min-w-0">
                            {isOwn && user ? (
                              <Avatar className="h-6 w-6">
                                <AvatarImage src={user.imageUrl} alt="" />
                                <AvatarFallback className="text-[10px]">
                                  {initialsFor(user.id)}
                                </AvatarFallback>
                              </Avatar>
                            ) : (
                              <Avatar className="h-6 w-6" aria-hidden="true">
                                <AvatarFallback className="text-[10px]">
                                  {initialsFor(comment.userId)}
                                </AvatarFallback>
                              </Avatar>
                            )}
                            {isOwn && <Badge variant="secondary">You</Badge>}
                            <span
                              className="text-xs text-muted-foreground truncate"
                              title={new Date(comment.createdAt).toLocaleString()}
                            >
                              {isOptimistic
                                ? "Sending…"
                                : <>
                                    {formatRelativeTime(comment.createdAt)}
                                    {comment.createdAt !== comment.updatedAt && " • edited"}
                                  </>}
                            </span>
                          </div>

                          {comment.userId === user?.id && !isOptimistic && (
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-6 w-6" aria-label="Comment actions">
                                  <MoreHorizontal className="h-4 w-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => startEditing(comment)}>
                                  <Edit className="h-4 w-4 mr-2" />
                                  Edit
                                </DropdownMenuItem>
                                <DropdownMenuItem
                                  onClick={() => setDeleteCommentId(comment.id)}
                                  className="text-destructive focus:text-destructive"
                                >
                                  <Trash2 className="h-4 w-4 mr-2" />
                                  Delete
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          )}
                        </div>

                        {editingComment === comment.id ? (
                          <Form {...editForm}>
                            <form onSubmit={editForm.handleSubmit(onEditSubmit)} className="space-y-2">
                              <FormField
                                control={editForm.control}
                                name="content"
                                render={({ field }) => (
                                  <FormItem>
                                    <FormControl>
                                      <Textarea
                                        placeholder="Edit your comment..."
                                        rows={3}
                                        {...field}
                                        disabled={updateComment.isPending}
                                        onKeyDown={(e) => {
                                          if (
                                            (e.metaKey || e.ctrlKey) &&
                                            e.key === "Enter"
                                          ) {
                                            e.preventDefault();
                                            void editForm.handleSubmit(onEditSubmit)();
                                          }
                                        }}
                                      />
                                    </FormControl>
                                    <FormMessage />
                                  </FormItem>
                                )}
                              />
                              <div className="flex justify-end gap-2">
                                <Button
                                  type="button"
                                  variant="outline"
                                  size="sm"
                                  onClick={cancelEditing}
                                  disabled={updateComment.isPending}
                                >
                                  Cancel
                                </Button>
                                <Button
                                  type="submit"
                                  size="sm"
                                  disabled={updateComment.isPending}
                                >
                                  {updateComment.isPending ? "Updating..." : "Update"}
                                </Button>
                              </div>
                            </form>
                          </Form>
                        ) : (
                          <p className="text-sm whitespace-pre-wrap">{comment.content}</p>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                  );
                })
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <MessageSquare className="h-12 w-12 mx-auto mb-2 opacity-50" />
                  <p className="font-medium text-foreground">No discussion yet</p>
                  <p className="text-xs mt-1">
                    Decisions, blockers, and context for this task live here.
                  </p>
                </div>
              )}
            </div>
          </ScrollArea>

          {/* Add Comment Form */}
          <div className="border-t pt-4 mt-4">
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="content"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Textarea
                          placeholder="Add a comment… (⌘/Ctrl+Enter to send)"
                          rows={3}
                          {...field}
                          disabled={addComment.isPending}
                          onKeyDown={(e) => {
                            if (
                              (e.metaKey || e.ctrlKey) &&
                              e.key === "Enter"
                            ) {
                              e.preventDefault();
                              void form.handleSubmit(onSubmit)();
                            }
                          }}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="flex items-center justify-between gap-2">
                  <p className="text-[11px] text-muted-foreground">
                    Press{" "}
                    <kbd className="rounded border bg-muted px-1 font-sans text-[10px]">
                      ⌘/Ctrl+Enter
                    </kbd>{" "}
                    to send
                  </p>
                  <Button type="submit" disabled={addComment.isPending}>
                    <Send className="h-4 w-4 mr-2" />
                    {addComment.isPending ? "Adding..." : "Add Comment"}
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        </ErrorBoundary>
      </div>

      {/* Delete Comment Dialog */}
      <AlertDialog
        open={!!deleteCommentId}
        onOpenChange={(open) => !open && setDeleteCommentId(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Comment</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this comment? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteCommentId && handleDeleteComment(deleteCommentId)}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              disabled={deleteComment.isPending}
            >
              {deleteComment.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
