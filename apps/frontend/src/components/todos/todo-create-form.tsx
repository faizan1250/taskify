import { useGetAllCategories } from "@/api/hooks/use-category-query";
import { useCreateTodo, useGetAllTodos } from "@/api/hooks/use-todo-query";
import { useUploadTodoAttachment } from "@/api/hooks/use-todo-attachment-query";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { CalendarIcon, Upload, X } from "lucide-react";
import { type ReactNode, useState, useRef, useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";

const createTodoSchema = z.object({
  title: z.string().min(1, "Title is required").max(255, "Title too long"),
  description: z.string().optional(),
  priority: z.enum(["low", "medium", "high"]),
  categoryId: z.string().optional(),
  parentTodoId: z.string().optional(),
  dueDate: z.date().optional(),
});

type CreateTodoForm = z.infer<typeof createTodoSchema>;

interface TodoCreateFormProps {
  children: ReactNode;
  /**
   * Controlled open state for programmatic opening (e.g. the `C`
   * hotkey). Omit for the default uncontrolled trigger behavior.
   */
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  /**
   * Preselects the Parent Task field, used by the per-task "Add subtask"
   * quick action. The dialog resets back to it on every open.
   */
  defaultParentId?: string;
}

export function TodoCreateForm({
  children,
  open,
  onOpenChange,
  defaultParentId,
}: TodoCreateFormProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isControlled = open !== undefined;
  const dialogOpen = isControlled ? open : internalOpen;
  const handleOpenChange = (next: boolean) => {
    if (!isControlled) setInternalOpen(next);
    onOpenChange?.(next);
  };
  const [attachments, setAttachments] = useState<File[]>([]);
  const [failedUploads, setFailedUploads] = useState<File[]>([]);
  const [createdTodoId, setCreatedTodoId] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const createTodo = useCreateTodo();
  const uploadAttachment = useUploadTodoAttachment();
  const { data: categories } = useGetAllCategories({
    query: { page: 1, limit: 100 },
  });
  const { data: todos } = useGetAllTodos({
    query: { page: 1, limit: 100 },
  });
  const parentCandidates = (todos?.data ?? []).filter(
    (t) => t.parentTodoId == null,
  );
  const defaultParentName = defaultParentId
    ? (todos?.data ?? []).find((t) => t.id === defaultParentId)?.title
    : undefined;

  const form = useForm<CreateTodoForm>({
    resolver: zodResolver(createTodoSchema),
    defaultValues: {
      title: "",
      description: "",
      priority: "medium",
      categoryId: "none",
      parentTodoId: defaultParentId ?? "none",
    },
  });

  // Every open starts from a clean slate, preselecting the quick-action
  // parent when one was provided.
  useEffect(() => {
    if (dialogOpen) {
      form.reset({
        title: "",
        description: "",
        priority: "medium",
        categoryId: "none",
        parentTodoId: defaultParentId ?? "none",
        dueDate: undefined,
      });
      setAttachments([]);
      setFailedUploads([]);
      setCreatedTodoId(null);
    }
  }, [dialogOpen, defaultParentId, form]);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files || []);
    const validFiles = files.filter(file => {
      if (file.size > 10 * 1024 * 1024) {
        toast.error(`File ${file.name} is too large (max 10MB)`);
        return false;
      }
      return true;
    });
    
    setAttachments(prev => [...prev, ...validFiles]);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const removeAttachment = (index: number) => {
    const target = attachments[index];
    setAttachments((prev) => prev.filter((_, i) => i !== index));
    // Removing a failed file drops it from the retry set too.
    if (target && failedUploads.includes(target) && failedUploads.length === 1) {
      setFailedUploads([]);
      setCreatedTodoId(null);
    } else if (target) {
      setFailedUploads((prev) => prev.filter((f) => f !== target));
    }
  };

  const isRetrying = createdTodoId !== null && failedUploads.length > 0;
  const submitLabel = isRetrying
    ? uploadAttachment.isPending
      ? "Retrying…"
      : `Retry ${failedUploads.length} failed file${failedUploads.length === 1 ? "" : "s"}`
    : createTodo.isPending
      ? "Creating..."
      : uploadAttachment.isPending
        ? "Uploading..."
        : "Create Task";

  const onSubmit = async (data: CreateTodoForm) => {
    // Retry path: the task already exists, only the failed files remain.
    if (createdTodoId && failedUploads.length > 0) {
      const results = await Promise.allSettled(
        failedUploads.map((file) =>
          uploadAttachment.mutateAsync({ todoId: createdTodoId, file }),
        ),
      );
      const stillFailed = failedUploads.filter(
        (_, i) => results[i].status === "rejected",
      );
      setFailedUploads(stillFailed);
      if (stillFailed.length === 0) {
        toast.success("Attachments uploaded successfully!");
        form.reset({
          title: "",
          description: "",
          priority: "medium",
          categoryId: "none",
          parentTodoId: defaultParentId ?? "none",
          dueDate: undefined,
        });
        setAttachments([]);
        setCreatedTodoId(null);
        handleOpenChange(false);
      } else {
        toast.error(
          `${stillFailed.length} attachment${stillFailed.length === 1 ? "" : "s"} still failing — files are kept above for another retry.`,
        );
      }
      return;
    }

    try {
      const createdTodo = await createTodo.mutateAsync({
        body: {
          title: data.title,
          description: data.description || undefined,
          priority: data.priority,
          categoryId:
            data.categoryId === "none" ? undefined : data.categoryId || undefined,
          parentTodoId:
            data.parentTodoId === "none"
              ? undefined
              : data.parentTodoId || undefined,
          dueDate: data.dueDate?.toISOString(),
        },
      });

      // Two-phase honesty: the task exists even if uploads fail, so never
      // report "failed to create" for an upload problem — keep the dialog
      // open with the failed files listed for retry.
      if (attachments.length > 0) {
        const results = await Promise.allSettled(
          attachments.map((file) =>
            uploadAttachment.mutateAsync({ todoId: createdTodo.id, file }),
          ),
        );
        const failed = attachments.filter(
          (_, i) => results[i].status === "rejected",
        );
        if (failed.length > 0) {
          setCreatedTodoId(createdTodo.id);
          setFailedUploads(failed);
          setAttachments(failed);
          toast.success(
            `Task created — ${failed.length} attachment${failed.length === 1 ? "" : "s"} failed to upload, retry below.`,
          );
          return;
        }
      }

      toast.success("Task created successfully!");
      form.reset({
        title: "",
        description: "",
        priority: "medium",
        categoryId: "none",
        parentTodoId: defaultParentId ?? "none",
        dueDate: undefined,
      });
      setAttachments([]);
      handleOpenChange(false);
    } catch {
      toast.error("Failed to create task");
    }
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle>Create New Task</DialogTitle>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 overflow-y-auto px-1">
            {failedUploads.length > 0 && (
              <div
                role="alert"
                className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm"
              >
                <p className="font-medium text-destructive">
                  Task created, but {failedUploads.length} attachment
                  {failedUploads.length === 1 ? "" : "s"} failed to upload:
                </p>
                <ul className="mt-1 list-disc pl-5 text-muted-foreground">
                  {failedUploads.map((file) => (
                    <li key={file.name} className="truncate">
                      {file.name}
                    </li>
                  ))}
                </ul>
                <p className="mt-1 text-xs text-muted-foreground">
                  Press “Create Task” again to retry the failed files, or
                  close and attach them later.
                </p>
              </div>
            )}
            <FormField
              control={form.control}
              name="title"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title *</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Enter task title..."
                      {...field}
                      disabled={createTodo.isPending}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Enter task description..."
                      rows={3}
                      {...field}
                      disabled={createTodo.isPending}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="priority"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Priority</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      disabled={createTodo.isPending}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select priority" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="low">
                          <div className="flex items-center gap-2">
                            <div className="w-2 h-2 rounded-full bg-green-500" />
                            Low
                          </div>
                        </SelectItem>
                        <SelectItem value="medium">
                          <div className="flex items-center gap-2">
                            <div className="w-2 h-2 rounded-full bg-yellow-500" />
                            Medium
                          </div>
                        </SelectItem>
                        <SelectItem value="high">
                          <div className="flex items-center gap-2">
                            <div className="w-2 h-2 rounded-full bg-red-500" />
                            High
                          </div>
                        </SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="categoryId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Category</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                      disabled={createTodo.isPending}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select category" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="none">No Category</SelectItem>
                        {categories?.data?.map((category) => (
                          <SelectItem key={category.id} value={category.id}>
                            <div className="flex items-center gap-2">
                              <div
                                className="w-2 h-2 rounded-full"
                                style={{ backgroundColor: category.color }}
                              />
                              {category.name}
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="parentTodoId"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Parent Task</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                    disabled={createTodo.isPending}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="No parent (top-level task)" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="none">
                        No parent (top-level task)
                      </SelectItem>
                      {parentCandidates.map((todo) => (
                        <SelectItem key={todo.id} value={todo.id}>
                          {todo.title}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {defaultParentId && defaultParentName ? (
                    <p className="text-xs text-muted-foreground">
                      Creating a subtask of “{defaultParentName}”.
                    </p>
                  ) : (
                    <FormMessage />
                  )}
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="dueDate"
              render={({ field }) => (
                <FormItem className="flex flex-col">
                  <FormLabel>Due Date</FormLabel>
                  <Popover>
                    <PopoverTrigger asChild>
                      <FormControl>
                        <Button
                          variant="outline"
                          className={cn(
                            "w-full pl-3 text-left font-normal",
                            !field.value && "text-muted-foreground",
                          )}
                          disabled={createTodo.isPending}
                        >
                          {field.value ? (
                            format(field.value, "PP")
                          ) : (
                            <span>Pick a date</span>
                          )}
                          <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                        </Button>
                      </FormControl>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0" align="start">
                      <Calendar
                        mode="single"
                        selected={field.value}
                        onSelect={field.onChange}
                        disabled={(date) =>
                          date < new Date(new Date().setHours(0, 0, 0, 0))
                        }
                        initialFocus
                      />
                    </PopoverContent>
                  </Popover>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Attachments Section */}
            <div className="space-y-3">
              <Label>Attachments</Label>
              <div className="flex items-center gap-2">
                <input
                  ref={fileInputRef}
                  type="file"
                  multiple
                  onChange={handleFileSelect}
                  className="hidden"
                  disabled={createTodo.isPending}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={createTodo.isPending}
                >
                  <Upload className="h-4 w-4 mr-2" />
                  Add Files
                </Button>
                <span className="text-sm text-muted-foreground">
                  Max 10MB per file
                </span>
              </div>
              
              {attachments.length > 0 && (
                <div className="space-y-2">
                  {attachments.map((file, index) => (
                    <div key={index} className="flex items-center justify-between p-2 bg-muted rounded-md">
                      <div className="flex items-center gap-2 min-w-0 flex-1">
                        <Upload className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium truncate">{file.name}</p>
                          <p className="text-xs text-muted-foreground">
                            {(file.size / (1024 * 1024)).toFixed(2)} MB
                          </p>
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removeAttachment(index)}
                        disabled={createTodo.isPending}
                        aria-label={`Remove attachment ${file.name}`}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="flex justify-end gap-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={createTodo.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createTodo.isPending || uploadAttachment.isPending}>
                {submitLabel}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
