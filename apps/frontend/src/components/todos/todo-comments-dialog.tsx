import { type ReactNode, useState } from "react";
import { TodoCommentsPanel } from "./todo-comments-panel";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { MessageSquare } from "lucide-react";

interface TodoCommentsDialogProps {
  todoId: string;
  children: ReactNode;
}

export function TodoCommentsDialog({ todoId, children }: TodoCommentsDialogProps) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <MessageSquare className="h-5 w-5" />
            Comments
          </DialogTitle>
        </DialogHeader>

        {open && <TodoCommentsPanel todoId={todoId} />}
      </DialogContent>
    </Dialog>
  );
}
