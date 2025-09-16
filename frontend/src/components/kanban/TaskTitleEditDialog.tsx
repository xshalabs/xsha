import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
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
import { Button } from "@/components/ui/button";
import { tasksApi } from "@/lib/api/tasks";
import type { Task } from "@/types/task";

const formSchema = z.object({
  title: z
    .string()
    .min(1, "Title is required")
    .max(200, "Title must be at most 200 characters")
    .trim(),
});

type FormData = z.infer<typeof formSchema>;

interface TaskTitleEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  task: Task;
  onSuccess: (updatedTitle: string) => void;
}

export function TaskTitleEditDialog({
  open,
  onOpenChange,
  task,
  onSuccess,
}: TaskTitleEditDialogProps) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      title: "",
    },
  });

  useEffect(() => {
    if (task && open) {
      form.reset({
        title: task.title,
      });
    }
  }, [task, open, form]);

  const handleSubmit = async (data: FormData) => {
    if (!task) return;

    const trimmedTitle = data.title.trim();
    if (trimmedTitle === task.title) {
      onOpenChange(false);
      return;
    }

    setLoading(true);
    try {
      await tasksApi.update(task.project_id, task.id, { title: trimmedTitle });
      toast.success(t("tasks.messages.updateSuccess"));
      onSuccess(trimmedTitle);
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to update task title:", error);
      toast.error(t("tasks.messages.updateFailed"));
    } finally {
      setLoading(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!loading) {
      onOpenChange(newOpen);
      if (!newOpen) {
        form.reset();
      }
    }
  };

  const handleCancel = () => {
    handleOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{t("tasks.editTitle.dialogTitle")}</DialogTitle>
          <DialogDescription>
            {t("tasks.editTitle.dialogDescription")}
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="title"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("tasks.fields.title")}</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      placeholder={t("tasks.form.titlePlaceholder")}
                      disabled={loading}
                      autoFocus
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
                onClick={handleCancel}
                disabled={loading}
              >
                {t("common.cancel")}
              </Button>
              <Button type="submit" disabled={loading}>
                {loading ? t("common.processing") : t("common.save")}
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}