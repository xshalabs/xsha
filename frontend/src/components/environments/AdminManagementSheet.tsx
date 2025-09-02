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
import { logError } from "@/lib/errors";
import type { DevEnvironment } from "@/types/dev-environment";
import type { Admin } from "@/lib/api/types";

interface AdminManagementSheetProps {
  environment: DevEnvironment;
  trigger?: React.ReactNode;
  onAdminChanged?: () => void;
}

export function AdminManagementSheet({
  environment,
  trigger,
  onAdminChanged,
}: AdminManagementSheetProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [environmentAdmins, setEnvironmentAdmins] = useState<Admin[]>([]);
  const [availableAdmins, setAvailableAdmins] = useState<Admin[]>([]);
  const [selectedAdminId, setSelectedAdminId] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  const [isAddingAdmin, setIsAddingAdmin] = useState(false);

  // Load environment admins and available admins when sheet opens
  useEffect(() => {
    if (open) {
      loadEnvironmentAdmins();
      loadAvailableAdmins();
    }
  }, [open]);

  const loadEnvironmentAdmins = async () => {
    try {
      setIsLoading(true);
      const response = await apiService.devEnvironments.getAdmins(environment.id);
      setEnvironmentAdmins(response.admins);
    } catch (error) {
      logError(error, "Failed to load environment admins");
      toast.error(t("devEnvironments.admin.load_failed"));
    } finally {
      setIsLoading(false);
    }
  };

  const loadAvailableAdmins = async () => {
    try {
      const response = await apiService.admin.getAdmins({ page: 1, page_size: 100 });
      setAvailableAdmins(response.admins);
    } catch (error) {
      logError(error, "Failed to load available admins");
      toast.error(t("admin.load_failed"));
    }
  };

  const handleAddAdmin = async () => {
    if (!selectedAdminId) {
      toast.error(t("devEnvironments.admin.select_admin"));
      return;
    }

    try {
      setIsAddingAdmin(true);
      await apiService.devEnvironments.addAdmin(environment.id, {
        admin_id: parseInt(selectedAdminId),
      });

      toast.success(t("devEnvironments.admin.added_success"));
      setSelectedAdminId("");
      await loadEnvironmentAdmins();
    } catch (error) {
      logError(error, "Failed to add admin to environment");
      toast.error(t("devEnvironments.admin.add_failed"));
    } finally {
      setIsAddingAdmin(false);
    }
  };

  const handleRemoveAdmin = async (adminId: number, adminName: string) => {
    try {
      await apiService.devEnvironments.removeAdmin(environment.id, adminId);

      toast.success(t("devEnvironments.admin.removed_success", { name: adminName }));
      await loadEnvironmentAdmins();
    } catch (error) {
      logError(error, "Failed to remove admin from environment");
      toast.error(t("devEnvironments.admin.remove_failed"));
    }
  };

  const handleClose = () => {
    setOpen(false);
    setSelectedAdminId("");
    onAdminChanged?.();
  };

  // Filter available admins to exclude those already assigned
  const unassignedAdmins = availableAdmins.filter(
    (admin) => !environmentAdmins.some((envAdmin) => envAdmin.id === admin.id)
  );

  return (
    <FormSheet open={open} onOpenChange={setOpen}>
      {trigger}
      <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
        <FormSheetHeader className="border-b">
          <FormSheetTitle className="text-foreground font-semibold">
            {t("devEnvironments.admin.manage_title")}
          </FormSheetTitle>
          <FormSheetDescription className="text-muted-foreground text-sm">
            {t("devEnvironments.admin.manage_description", {
              name: environment.name,
            })}
          </FormSheetDescription>
        </FormSheetHeader>

        <div className="flex-1 flex flex-col space-y-4 overflow-y-auto">
          {/* Add Admin Section */}
          <div className="p-4">
            <div className="space-y-4">
              <h4 className="text-sm font-medium">{t("devEnvironments.admin.add_admin")}</h4>
              <div className="flex gap-2 items-center">
                <Select value={selectedAdminId} onValueChange={setSelectedAdminId}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder={t("devEnvironments.admin.select_placeholder")} />
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
              {t("devEnvironments.admin.current_admins")} ({environmentAdmins.length})
            </h4>
            
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : environmentAdmins.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <div className="text-muted-foreground text-sm">
                  {t("devEnvironments.admin.no_admins")}
                </div>
              </div>
            ) : (
              <ScrollArea className="flex-1">
                <div className="space-y-2 pr-2">
                  {environmentAdmins.map((admin) => (
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
                          <div className="font-medium truncate">{admin.name}</div>
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
                        onClick={() => handleRemoveAdmin(admin.id, admin.name)}
                        className="shrink-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                      >
                        <Trash2 className="h-4 w-4" />
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
    </FormSheet>
  );
}