import { SelectTrigger } from "@radix-ui/react-select";
import type { Table } from "@tanstack/react-table";
import { CheckCircle2, Trash2 } from "lucide-react";
import { useState } from "react";

import { useTranslation } from "react-i18next";

import {
  DataTableActionBar,
  DataTableActionBarAction,
  DataTableActionBarSelection,
} from "@/components/ui/data-table/data-table-action-bar";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
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
import type { Task, TaskStatus } from "@/types/task";

interface TaskDataTableActionBarProps {
  table: Table<Task>;
  onBatchUpdateStatus?: (taskIds: number[], status: TaskStatus) => void;
  onBatchDelete?: (taskIds: number[]) => void;
}

export function TaskDataTableActionBar({
  table,
  onBatchUpdateStatus,
  onBatchDelete,
}: TaskDataTableActionBarProps) {
  const { t } = useTranslation();
  const rows = table.getFilteredSelectedRowModel().rows;
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const handleStatusUpdate = (status: TaskStatus) => {
    const taskIds = rows.map(row => row.original.id);
    if (onBatchUpdateStatus && taskIds.length > 0) {
      onBatchUpdateStatus(taskIds, status);
      table.resetRowSelection();
    }
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = () => {
    const taskIds = rows.map(row => row.original.id);
    if (onBatchDelete && taskIds.length > 0) {
      onBatchDelete(taskIds);
      table.resetRowSelection();
      setDeleteDialogOpen(false);
    }
  };

  return (
    <DataTableActionBar table={table} visible={rows.length > 0}>
      <DataTableActionBarSelection table={table} />
      <Separator
        orientation="vertical"
        className="hidden data-[orientation=vertical]:h-5 sm:block"
      />
      <div className="flex items-center gap-1.5">
        <Select onValueChange={handleStatusUpdate}>
          <SelectTrigger asChild>
            <DataTableActionBarAction size="icon" tooltip={t("tasks.batch.updateStatus")}>
              <CheckCircle2 />
            </DataTableActionBarAction>
          </SelectTrigger>
          <SelectContent align="center">
            <SelectGroup>
              <SelectItem value="todo">{t("tasks.status.todo")}</SelectItem>
              <SelectItem value="in_progress">{t("tasks.status.in_progress")}</SelectItem>
              <SelectItem value="done">{t("tasks.status.done")}</SelectItem>
              <SelectItem value="cancelled">{t("tasks.status.cancelled")}</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <DataTableActionBarAction 
          size="icon" 
          tooltip={t("tasks.batch.delete")}
          onClick={handleDeleteClick}
        >
          <Trash2 />
        </DataTableActionBarAction>
      </div>

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("tasks.batch.deleteConfirmTitle")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("tasks.batch.deleteConfirmDescription", { count: rows.length })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-white shadow-sm hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60"
              onClick={handleDeleteConfirm}
            >
              {t("common.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </DataTableActionBar>
  );
}
