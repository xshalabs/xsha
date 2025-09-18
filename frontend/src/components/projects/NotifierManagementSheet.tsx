import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
  FormSheet,
  FormSheetContent,
  FormSheetHeader,
  FormSheetTitle,
  FormSheetDescription,
  FormSheetFooter,
} from "@/components/forms/form-sheet";
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Plus, Trash2, Loader2, Bell, BellOff } from "lucide-react";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { logError, handleApiError } from "@/lib/errors";
import type { Project } from "@/types/project";
import type { Notifier } from "@/types/notifier";

interface NotifierManagementSheetProps {
  project: Project;
  trigger?: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onNotifierChanged?: () => void;
}

export function NotifierManagementSheet({
  project,
  trigger,
  open: externalOpen,
  onOpenChange: externalOnOpenChange,
  onNotifierChanged,
}: NotifierManagementSheetProps) {
  const { t } = useTranslation();
  const [internalOpen, setInternalOpen] = useState(false);

  // Use external open state if provided, otherwise use internal state
  const open = externalOpen !== undefined ? externalOpen : internalOpen;
  const setOpen = externalOnOpenChange !== undefined ? externalOnOpenChange : setInternalOpen;
  const [projectNotifiers, setProjectNotifiers] = useState<Notifier[]>([]);
  const [availableNotifiers, setAvailableNotifiers] = useState<Notifier[]>([]);
  const [selectedNotifierId, setSelectedNotifierId] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  const [isAddingNotifier, setIsAddingNotifier] = useState(false);
  const [isRemovingNotifier, setIsRemovingNotifier] = useState(false);
  const [showRemoveConfirmDialog, setShowRemoveConfirmDialog] = useState(false);
  const [notifierToRemove, setNotifierToRemove] = useState<Notifier | null>(null);

  // Load project notifiers and available notifiers when sheet opens
  useEffect(() => {
    if (open) {
      // Reset form state when opening
      setSelectedNotifierId("");
      setShowRemoveConfirmDialog(false);
      setNotifierToRemove(null);

      loadProjectNotifiers();
      loadAvailableNotifiers();
    }
  }, [open]);

  const loadProjectNotifiers = async () => {
    try {
      setIsLoading(true);
      const response = await apiService.notifiers.getProjectNotifiers(project.id);
      setProjectNotifiers(response.data);
    } catch (error) {
      logError(error, "Failed to load project notifiers");
      toast.error(handleApiError(error));
    } finally {
      setIsLoading(false);
    }
  };

  const loadAvailableNotifiers = async () => {
    try {
      const response = await apiService.notifiers.list();
      setAvailableNotifiers(response.data);
    } catch (error) {
      logError(error, "Failed to load available notifiers");
      toast.error(handleApiError(error));
    }
  };

  const handleAddNotifier = async () => {
    if (!selectedNotifierId) {
      toast.error(t("projects.notifier.select_notifier"));
      return;
    }

    try {
      setIsAddingNotifier(true);
      await apiService.notifiers.addToProject(project.id, {
        notifier_id: parseInt(selectedNotifierId),
      });

      toast.success(t("projects.notifier.added_success"));
      setSelectedNotifierId("");
      await loadProjectNotifiers();
    } catch (error) {
      logError(error, "Failed to add notifier to project");
      toast.error(handleApiError(error));
    } finally {
      setIsAddingNotifier(false);
    }
  };

  const handleRemoveNotifier = (notifier: Notifier) => {
    setNotifierToRemove(notifier);
    setShowRemoveConfirmDialog(true);
  };

  const confirmRemoveNotifier = async () => {
    if (!notifierToRemove) return;

    try {
      setIsRemovingNotifier(true);
      await apiService.notifiers.removeFromProject(project.id, notifierToRemove.id);

      toast.success(t("projects.notifier.removed_success", { name: notifierToRemove.name }));
      setShowRemoveConfirmDialog(false);
      setNotifierToRemove(null);
      await loadProjectNotifiers();
    } catch (error) {
      logError(error, "Failed to remove notifier from project");
      toast.error(handleApiError(error));
    } finally {
      setIsRemovingNotifier(false);
    }
  };

  const handleClose = () => {
    setOpen(false);
    setSelectedNotifierId("");
    setShowRemoveConfirmDialog(false);
    setNotifierToRemove(null);
    onNotifierChanged?.();
  };

  // Filter available notifiers to exclude those already assigned
  const unassignedNotifiers = availableNotifiers.filter(
    (notifier) => !projectNotifiers.some((projNotifier) => projNotifier.id === notifier.id)
  );

  // Helper function to get notifier type display name
  const getNotifierTypeDisplayName = (type: string) => {
    return t(`notifiers.type.${type}`, { defaultValue: type });
  };

  return (
    <FormSheet open={open} onOpenChange={setOpen}>
      {trigger && React.cloneElement(trigger as React.ReactElement<any>, {
        onClick: (e: React.MouseEvent) => {
          e.preventDefault();
          e.stopPropagation();
          if (externalOnOpenChange) {
            externalOnOpenChange(true);
          } else {
            setOpen(true);
          }
        }
      })}
      <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
        <FormSheetHeader className="border-b">
          <FormSheetTitle className="text-foreground font-semibold">
            {t("projects.notifier.manage_title")}
          </FormSheetTitle>
          <FormSheetDescription className="text-muted-foreground text-sm">
            {t("projects.notifier.manage_description", {
              name: project.name,
            })}
          </FormSheetDescription>
        </FormSheetHeader>

        <div className="flex-1 flex flex-col space-y-4 overflow-y-auto">
          {/* Add Notifier Section */}
          <div className="p-4">
            <div className="space-y-4">
              <h4 className="text-sm font-medium">{t("projects.notifier.add_notifier")}</h4>
              <div className="flex gap-2 items-center">
                <Select value={selectedNotifierId} onValueChange={setSelectedNotifierId}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder={t("projects.notifier.select_placeholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {unassignedNotifiers.map((notifier) => (
                      <SelectItem key={notifier.id} value={notifier.id.toString()}>
                        <div className="flex items-center gap-2">
                          {notifier.is_enabled ? (
                            <Bell className="h-4 w-4 text-green-500" />
                          ) : (
                            <BellOff className="h-4 w-4 text-muted-foreground" />
                          )}
                          <span className="font-medium">{notifier.name}</span>
                          <Badge variant="outline" className="text-xs">
                            {getNotifierTypeDisplayName(notifier.type)}
                          </Badge>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  type="button"
                  onClick={handleAddNotifier}
                  disabled={!selectedNotifierId || isAddingNotifier}
                  size="sm"
                  className="shrink-0"
                >
                  {isAddingNotifier ? (
                    <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                  ) : (
                    <Plus className="h-4 w-4 mr-1" />
                  )}
                  {t("common.add")}
                </Button>
              </div>
            </div>
          </div>

          {/* Current Notifiers Section */}
          <div className="border-t p-4 flex-1 flex flex-col">
            <h4 className="text-sm font-medium mb-4">
              {t("projects.notifier.current_notifiers")} ({projectNotifiers.length})
            </h4>

            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : projectNotifiers.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <div className="text-muted-foreground text-sm">
                  {t("projects.notifier.no_notifiers")}
                </div>
              </div>
            ) : (
              <ScrollArea className="flex-1">
                <div className="space-y-2 pr-2">
                  {projectNotifiers.map((notifier) => (
                    <div
                      key={notifier.id}
                      className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        {notifier.is_enabled ? (
                          <Bell className="h-4 w-4 text-green-500" />
                        ) : (
                          <BellOff className="h-4 w-4 text-muted-foreground" />
                        )}
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-2">
                            <div className="font-medium truncate">{notifier.name}</div>
                            <Badge variant="outline" className="text-xs shrink-0">
                              {getNotifierTypeDisplayName(notifier.type)}
                            </Badge>
                            {!notifier.is_enabled && (
                              <Badge variant="secondary" className="text-xs shrink-0">
                                {t("notifiers.status.disabled")}
                              </Badge>
                            )}
                          </div>
                          {notifier.description && (
                            <div className="text-sm text-muted-foreground truncate">
                              {notifier.description}
                            </div>
                          )}
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveNotifier(notifier)}
                        disabled={isRemovingNotifier && notifierToRemove?.id === notifier.id}
                        className="shrink-0 text-destructive dark:text-white hover:text-destructive dark:hover:text-white hover:bg-destructive/10"
                      >
                        {isRemovingNotifier && notifierToRemove?.id === notifier.id ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            )}
          </div>
        </div>

        <FormSheetFooter className="border-t">
          <Button variant="outline" onClick={handleClose}>
            {t("common.close")}
          </Button>
        </FormSheetFooter>
      </FormSheetContent>

      {/* Remove Notifier Confirmation Dialog */}
      <AlertDialog open={showRemoveConfirmDialog} onOpenChange={setShowRemoveConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("projects.notifier.confirm_remove_title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {notifierToRemove && (
                t("projects.notifier.confirm_remove_description", {
                  notifierName: notifierToRemove.name,
                  projectName: project.name
                })
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRemovingNotifier}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmRemoveNotifier}
              disabled={isRemovingNotifier}
              className="bg-destructive text-destructive-foreground dark:text-white hover:bg-destructive/90"
            >
              {isRemovingNotifier ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  {t("common.removing")}
                </>
              ) : (
                t("common.remove")
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </FormSheet>
  );
}