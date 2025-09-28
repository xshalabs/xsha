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
import { Plus, Trash2, Loader2, Settings, CircleSlash } from "lucide-react";
import { toast } from "sonner";
import { apiService } from "@/lib/api/index";
import { logError, handleApiError } from "@/lib/errors";
import type { Project } from "@/types/project";
import type { MCP } from "@/types/mcp";

interface MCPManagementSheetProps {
  project: Project;
  trigger?: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onMCPChanged?: () => void;
}

export function MCPManagementSheet({
  project,
  trigger,
  open: externalOpen,
  onOpenChange: externalOnOpenChange,
  onMCPChanged,
}: MCPManagementSheetProps) {
  const { t } = useTranslation();
  const [internalOpen, setInternalOpen] = useState(false);

  // Use external open state if provided, otherwise use internal state
  const open = externalOpen !== undefined ? externalOpen : internalOpen;
  const setOpen = externalOnOpenChange !== undefined ? externalOnOpenChange : setInternalOpen;
  const [projectMCPs, setProjectMCPs] = useState<MCP[]>([]);
  const [availableMCPs, setAvailableMCPs] = useState<MCP[]>([]);
  const [selectedMCPId, setSelectedMCPId] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  const [isAddingMCP, setIsAddingMCP] = useState(false);
  const [isRemovingMCP, setIsRemovingMCP] = useState(false);
  const [showRemoveConfirmDialog, setShowRemoveConfirmDialog] = useState(false);
  const [mcpToRemove, setMCPToRemove] = useState<MCP | null>(null);

  // Load project MCPs and available MCPs when sheet opens
  useEffect(() => {
    if (open) {
      // Reset form state when opening
      setSelectedMCPId("");
      setShowRemoveConfirmDialog(false);
      setMCPToRemove(null);

      loadProjectMCPs();
      loadAvailableMCPs();
    }
  }, [open]);

  const loadProjectMCPs = async () => {
    try {
      setIsLoading(true);
      const response = await apiService.mcp.getProjectMCPs(project.id);
      setProjectMCPs(response.mcps || []);
    } catch (error) {
      logError(error, "Failed to load project MCPs");
      toast.error(handleApiError(error));
    } finally {
      setIsLoading(false);
    }
  };

  const loadAvailableMCPs = async () => {
    try {
      const response = await apiService.mcp.list();
      setAvailableMCPs(response.mcps || []);
    } catch (error) {
      logError(error, "Failed to load available MCPs");
      toast.error(handleApiError(error));
    }
  };

  const handleAddMCP = async () => {
    if (!selectedMCPId) {
      toast.error(t("projects.mcp.select_mcp"));
      return;
    }

    try {
      setIsAddingMCP(true);
      await apiService.mcp.addToProject(project.id, {
        mcp_id: parseInt(selectedMCPId),
      });

      toast.success(t("projects.mcp.added_success"));
      setSelectedMCPId("");
      await loadProjectMCPs();
    } catch (error) {
      logError(error, "Failed to add MCP to project");
      toast.error(handleApiError(error));
    } finally {
      setIsAddingMCP(false);
    }
  };

  const handleRemoveMCP = (mcp: MCP) => {
    setMCPToRemove(mcp);
    setShowRemoveConfirmDialog(true);
  };

  const confirmRemoveMCP = async () => {
    if (!mcpToRemove) return;

    try {
      setIsRemovingMCP(true);
      await apiService.mcp.removeFromProject(project.id, mcpToRemove.id);

      toast.success(t("projects.mcp.removed_success", { name: mcpToRemove.name }));
      setShowRemoveConfirmDialog(false);
      setMCPToRemove(null);
      await loadProjectMCPs();
    } catch (error) {
      logError(error, "Failed to remove MCP from project");
      toast.error(handleApiError(error));
    } finally {
      setIsRemovingMCP(false);
    }
  };

  const handleClose = () => {
    setOpen(false);
    setSelectedMCPId("");
    setShowRemoveConfirmDialog(false);
    setMCPToRemove(null);
    onMCPChanged?.();
  };

  // Filter available MCPs to exclude those already assigned
  const unassignedMCPs = availableMCPs.filter(
    (mcp) => !projectMCPs.some((projMCP) => projMCP.id === mcp.id)
  );

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
            {t("projects.mcp.manage_title")}
          </FormSheetTitle>
          <FormSheetDescription className="text-muted-foreground text-sm">
            {t("projects.mcp.manage_description", {
              name: project.name,
            })}
          </FormSheetDescription>
        </FormSheetHeader>

        <div className="flex-1 flex flex-col space-y-4 overflow-y-auto">
          {/* Add MCP Section */}
          <div className="p-4">
            <div className="space-y-4">
              <h4 className="text-sm font-medium">{t("projects.mcp.add_mcp")}</h4>
              <div className="flex gap-2 items-center">
                <Select value={selectedMCPId} onValueChange={setSelectedMCPId}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder={t("projects.mcp.select_placeholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {unassignedMCPs.map((mcp) => (
                      <SelectItem key={mcp.id} value={mcp.id.toString()}>
                        <div className="flex items-center gap-2">
                          {mcp.enabled ? (
                            <Settings className="h-4 w-4 text-green-500" />
                          ) : (
                            <CircleSlash className="h-4 w-4 text-muted-foreground" />
                          )}
                          <span className="font-medium">{mcp.name}</span>
                          <Badge variant="outline" className="text-xs">
                            {mcp.enabled ? t("mcp.status.enabled") : t("mcp.status.disabled")}
                          </Badge>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  type="button"
                  onClick={handleAddMCP}
                  disabled={!selectedMCPId || isAddingMCP}
                  size="sm"
                  className="shrink-0"
                >
                  {isAddingMCP ? (
                    <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                  ) : (
                    <Plus className="h-4 w-4 mr-1" />
                  )}
                  {t("common.add")}
                </Button>
              </div>
            </div>
          </div>

          {/* Current MCPs Section */}
          <div className="border-t p-4 flex-1 flex flex-col">
            <h4 className="text-sm font-medium mb-4">
              {t("projects.mcp.current_mcps")} ({projectMCPs.length})
            </h4>

            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : projectMCPs.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <div className="text-muted-foreground text-sm">
                  {t("projects.mcp.no_mcps")}
                </div>
              </div>
            ) : (
              <ScrollArea className="flex-1">
                <div className="space-y-2 pr-2">
                  {projectMCPs.map((mcp) => (
                    <div
                      key={mcp.id}
                      className="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        {mcp.enabled ? (
                          <Settings className="h-4 w-4 text-green-500" />
                        ) : (
                          <CircleSlash className="h-4 w-4 text-muted-foreground" />
                        )}
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-2">
                            <div className="font-medium truncate">{mcp.name}</div>
                            <Badge variant="outline" className="text-xs shrink-0">
                              {mcp.enabled ? t("mcp.status.enabled") : t("mcp.status.disabled")}
                            </Badge>
                          </div>
                          {mcp.description && (
                            <div className="text-sm text-muted-foreground truncate">
                              {mcp.description}
                            </div>
                          )}
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveMCP(mcp)}
                        disabled={isRemovingMCP && mcpToRemove?.id === mcp.id}
                        className="shrink-0 text-destructive dark:text-white hover:text-destructive dark:hover:text-white hover:bg-destructive/10"
                      >
                        {isRemovingMCP && mcpToRemove?.id === mcp.id ? (
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

      {/* Remove MCP Confirmation Dialog */}
      <AlertDialog open={showRemoveConfirmDialog} onOpenChange={setShowRemoveConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("projects.mcp.confirm_remove_title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {mcpToRemove && (
                t("projects.mcp.confirm_remove_description", {
                  mcpName: mcpToRemove.name,
                  projectName: project.name
                })
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRemovingMCP}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmRemoveMCP}
              disabled={isRemovingMCP}
              className="bg-destructive text-destructive-foreground dark:text-white hover:bg-destructive/90"
            >
              {isRemovingMCP ? (
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