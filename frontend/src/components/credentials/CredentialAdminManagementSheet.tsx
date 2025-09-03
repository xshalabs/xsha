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
import { UserAvatar } from "@/components/ui/user-avatar";
import { Plus, Trash2, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { logError, handleApiError } from "@/lib/errors";
import type { GitCredential, MinimalAdminResponse } from "@/types/credentials";
import type { Admin } from "@/lib/api/types";

interface CredentialAdminManagementSheetProps {
  credential: GitCredential;
  trigger?: React.ReactNode;
  onAdminChanged?: () => void;
}

export function CredentialAdminManagementSheet({
  credential,
  trigger,
  onAdminChanged,
}: CredentialAdminManagementSheetProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [credentialAdmins, setCredentialAdmins] = useState<MinimalAdminResponse[]>([]);
  const [availableAdmins, setAvailableAdmins] = useState<Admin[]>([]);
  const [selectedAdminId, setSelectedAdminId] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  const [isAddingAdmin, setIsAddingAdmin] = useState(false);
  const [isRemovingAdmin, setIsRemovingAdmin] = useState(false);
  const [showAddConfirmDialog, setShowAddConfirmDialog] = useState(false);
  const [showRemoveConfirmDialog, setShowRemoveConfirmDialog] = useState(false);
  const [adminToRemove, setAdminToRemove] = useState<MinimalAdminResponse | null>(null);

  // Load credential admins and available admins when sheet opens
  useEffect(() => {
    if (open) {
      loadCredentialAdmins();
      loadAvailableAdmins();
    }
  }, [open]);

  const loadCredentialAdmins = async () => {
    try {
      setIsLoading(true);
      const response = await apiService.gitCredentials.getAdmins(credential.id);
      setCredentialAdmins(response.admins);
    } catch (error) {
      logError(error, "Failed to load credential admins");
      toast.error(handleApiError(error));
    } finally {
      setIsLoading(false);
    }
  };

  const loadAvailableAdmins = async () => {
    try {
      const response = await apiService.admin.getV1Admins();
      setAvailableAdmins(response.admins);
    } catch (error) {
      logError(error, "Failed to load available admins");
      toast.error(handleApiError(error));
    }
  };

  const handleAddAdmin = () => {
    if (!selectedAdminId) {
      toast.error(t("gitCredentials.admin.select_admin"));
      return;
    }
    setShowAddConfirmDialog(true);
  };

  const confirmAddAdmin = async () => {
    try {
      setIsAddingAdmin(true);
      await apiService.gitCredentials.addAdmin(credential.id, {
        admin_id: parseInt(selectedAdminId),
      });

      toast.success(t("gitCredentials.admin.added_success"));
      setSelectedAdminId("");
      setShowAddConfirmDialog(false);
      await loadCredentialAdmins();
    } catch (error) {
      logError(error, "Failed to add admin to credential");
      toast.error(handleApiError(error));
    } finally {
      setIsAddingAdmin(false);
    }
  };

  const handleRemoveAdmin = (admin: MinimalAdminResponse) => {
    setAdminToRemove(admin);
    setShowRemoveConfirmDialog(true);
  };

  const confirmRemoveAdmin = async () => {
    if (!adminToRemove) return;

    try {
      setIsRemovingAdmin(true);
      await apiService.gitCredentials.removeAdmin(credential.id, adminToRemove.id);

      toast.success(t("gitCredentials.admin.removed_success", { name: adminToRemove.name }));
      setShowRemoveConfirmDialog(false);
      setAdminToRemove(null);
      await loadCredentialAdmins();
    } catch (error) {
      logError(error, "Failed to remove admin from credential");
      toast.error(handleApiError(error));
    } finally {
      setIsRemovingAdmin(false);
    }
  };

  const handleClose = () => {
    setOpen(false);
    setSelectedAdminId("");
    setShowAddConfirmDialog(false);
    setShowRemoveConfirmDialog(false);
    setAdminToRemove(null);
    onAdminChanged?.();
  };

  // Filter available admins to exclude those already assigned
  const unassignedAdmins = availableAdmins.filter(
    (admin) => !credentialAdmins.some((credAdmin) => credAdmin.id === admin.id)
  );

  return (
    <FormSheet open={open} onOpenChange={setOpen}>
      {trigger && React.cloneElement(trigger as React.ReactElement<unknown>, {
        onClick: (e: React.MouseEvent) => {
          e.preventDefault();
          e.stopPropagation();
          setOpen(true);
        }
      })}
      <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
        <FormSheetHeader className="border-b">
          <FormSheetTitle className="text-foreground font-semibold">
            {t("gitCredentials.admin.manage_title")}
          </FormSheetTitle>
          <FormSheetDescription className="text-muted-foreground text-sm">
            {t("gitCredentials.admin.manage_description", {
              name: credential.name,
            })}
          </FormSheetDescription>
        </FormSheetHeader>

        <div className="flex-1 flex flex-col space-y-4 overflow-y-auto">
          {/* Add Admin Section */}
          <div className="p-4">
            <div className="space-y-4">
              <h4 className="text-sm font-medium">{t("gitCredentials.admin.add_admin")}</h4>
              <div className="flex gap-2 items-center">
                <Select value={selectedAdminId} onValueChange={setSelectedAdminId}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder={t("gitCredentials.admin.select_placeholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {unassignedAdmins.map((admin) => (
                      <SelectItem key={admin.id} value={admin.id.toString()}>
                        <div className="flex items-center gap-2">
                          <UserAvatar 
                            user={admin.username}
                            name={admin.name}
                            avatar={admin.avatar}
                            size="sm"
                          />
                          <span className="font-medium">{admin.name}</span>
                          <span className="text-muted-foreground">({admin.username})</span>
                          <Badge variant="outline" className="text-xs">
                            {admin.role}
                          </Badge>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  type="button"
                  onClick={handleAddAdmin}
                  disabled={!selectedAdminId || isAddingAdmin}
                  size="sm"
                  className="shrink-0"
                >
                  {isAddingAdmin ? (
                    <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                  ) : (
                    <Plus className="h-4 w-4 mr-1" />
                  )}
                  {t("common.add")}
                </Button>
              </div>
            </div>
          </div>

          {/* Current Admins Section */}
          <div className="border-t p-4 flex-1 flex flex-col">
            <h4 className="text-sm font-medium mb-4">
              {t("gitCredentials.admin.current_admins")} ({credentialAdmins.length})
            </h4>
            
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : credentialAdmins.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <div className="text-muted-foreground text-sm">
                  {t("gitCredentials.admin.no_admins")}
                </div>
              </div>
            ) : (
              <ScrollArea className="flex-1">
                <div className="space-y-2 pr-2">
                  {credentialAdmins.map((admin) => {
                    const isPrimaryAdmin = credential.admin_id === admin.id;
                    return (
                      <div
                        key={admin.id}
                        className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <UserAvatar 
                            user={admin.username}
                            name={admin.name}
                            avatar={admin.avatar}
                            size="sm"
                          />
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2">
                              <div className="font-medium truncate">{admin.name}</div>
                              {isPrimaryAdmin && (
                                <Badge variant="secondary" className="text-xs shrink-0">
                                  {t("gitCredentials.admin.creator")}
                                </Badge>
                              )}
                            </div>
                            <div className="text-sm text-muted-foreground truncate">
                              @{admin.username}
                              {admin.email && ` • ${admin.email}`}
                            </div>
                          </div>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemoveAdmin(admin)}
                          disabled={
                            isPrimaryAdmin || 
                            (isRemovingAdmin && adminToRemove?.id === admin.id)
                          }
                          className="shrink-0 text-destructive dark:text-white hover:text-destructive dark:hover:text-white hover:bg-destructive/10 disabled:opacity-50 disabled:cursor-not-allowed"
                          title={isPrimaryAdmin ? t("gitCredentials.admin.cannot_remove_creator") : undefined}
                        >
                          {isRemovingAdmin && adminToRemove?.id === admin.id ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4" />
                          )}
                        </Button>
                      </div>
                    );
                  })}
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

      {/* Add Admin Confirmation Dialog */}
      <AlertDialog open={showAddConfirmDialog} onOpenChange={setShowAddConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("gitCredentials.admin.confirm_add_title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {selectedAdminId && (
                t("gitCredentials.admin.confirm_add_description", {
                  adminName: availableAdmins.find(admin => admin.id.toString() === selectedAdminId)?.name,
                  credentialName: credential.name
                })
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isAddingAdmin}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmAddAdmin}
              disabled={isAddingAdmin}
              className="bg-primary text-primary-foreground hover:bg-primary/90"
            >
              {isAddingAdmin ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  {t("common.adding")}
                </>
              ) : (
                t("common.add")
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Remove Admin Confirmation Dialog */}
      <AlertDialog open={showRemoveConfirmDialog} onOpenChange={setShowRemoveConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("gitCredentials.admin.confirm_remove_title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {adminToRemove && (
                t("gitCredentials.admin.confirm_remove_description", {
                  adminName: adminToRemove.name,
                  credentialName: credential.name
                })
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRemovingAdmin}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmRemoveAdmin}
              disabled={isRemovingAdmin}
              className="bg-destructive text-destructive-foreground dark:text-white hover:bg-destructive/90"
            >
              {isRemovingAdmin ? (
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