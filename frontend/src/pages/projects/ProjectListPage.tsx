import React, {
  useState,
  useEffect,
  useMemo,
  useCallback,
  useRef,
} from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { usePageActions } from "@/contexts/PageActionsContext";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { toast } from "sonner";

import { DataTable } from "@/components/ui/data-table/data-table";
import { Plus, Save } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { usePermissions } from "@/hooks/usePermissions";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from "@/components/content/section";
import {
  FormSheet,
  FormSheetContent,
  FormSheetHeader,
  FormSheetTitle,
  FormSheetDescription,
  FormSheetFooter,
  FormCardGroup,
} from "@/components/forms/form-sheet";
import { FormCard, FormCardContent } from "@/components/forms/form-card";
import { ProjectFormSheet } from "@/components/ProjectFormSheet";
import { AdminManagementSheet } from "@/components/projects/AdminManagementSheet";
import { NotifierManagementSheet } from "@/components/projects/NotifierManagementSheet";

import { createProjectColumns } from "@/components/data-table/projects/columns";
import { ProjectDataTableToolbar } from "@/components/data-table/projects/data-table-toolbar";
import { DataTablePaginationServer } from "@/components/ui/data-table/data-table-pagination-server";
import type {
  Project,
  ProjectListParams,
  GitProtocolType,
} from "@/types/project";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const ProjectListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreateProject, canEditProject, canDeleteProject } = usePermissions();

  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);
  const { setActions } = usePageActions();
  const { setItems } = useBreadcrumb();

  // Sheet state management
  const [isCreateSheetOpen, setIsCreateSheetOpen] = useState(false);
  const [isEditSheetOpen, setIsEditSheetOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [managingProject, setManagingProject] = useState<Project | null>(null);
  const [managingProjectForNotifiers, setManagingProjectForNotifiers] = useState<Project | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Add request deduplication
  const lastRequestRef = useRef<string>("");
  const isRequestInProgress = useRef(false);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 10;

  usePageTitle(t("common.pageTitle.projects"));

  const loadProjectsData = useCallback(
    async (page: number, filters: ColumnFiltersState, sortingState: SortingState, updateUrl = true) => {
      // Create a unique request key for deduplication
      const requestKey = JSON.stringify({ page, filters, sortingState, updateUrl });

      // Skip if same request is already in progress or just completed
      if (
        isRequestInProgress.current ||
        lastRequestRef.current === requestKey
      ) {
        return;
      }

      isRequestInProgress.current = true;
      lastRequestRef.current = requestKey;

      try {
        setLoading(true);

        // Convert DataTable filters to API parameters
        const apiParams: ProjectListParams = {
          page,
          page_size: pageSize,
        };

        // Handle column filters
        filters.forEach((filter) => {
          if (filter.id === "name" && filter.value) {
            apiParams.name = filter.value as string;
          } else if (filter.id === "protocol" && filter.value) {
            apiParams.protocol = filter.value as GitProtocolType;
          }
        });

        // Handle sorting
        if (sortingState.length > 0) {
          const sort = sortingState[0];
          apiParams.sort_by = sort.id;
          apiParams.sort_direction = sort.desc ? 'desc' : 'asc';
        }

        const response = await apiService.projects.list(apiParams);
        setProjects(response.projects);
        setTotal(response.total);
        setTotalPages(response.total_pages);
        setCurrentPage(page);

        // Update URL parameters
        if (updateUrl) {
          const params = new URLSearchParams();

          // Add filter parameters
          filters.forEach((filter) => {
            if (filter.value) {
              params.set(filter.id, String(filter.value));
            }
          });

          // Add sorting parameters
          if (sortingState.length > 0) {
            const sort = sortingState[0];
            params.set("sort_by", sort.id);
            params.set("sort_direction", sort.desc ? 'desc' : 'asc');
          }

          // Add page parameter (only if not page 1)
          if (page > 1) {
            params.set("page", String(page));
          }

          // Update URL without causing navigation
          setSearchParams(params, { replace: true });
        }
      } catch (error) {
        logError(error as Error, "Failed to load projects");
      } finally {
        setLoading(false);
        isRequestInProgress.current = false;

        // Clear the request key after a short delay to allow legitimate new requests
        setTimeout(() => {
          if (lastRequestRef.current === requestKey) {
            lastRequestRef.current = "";
          }
        }, 500); // Increase delay to prevent rapid duplicate requests
      }
    },
    [pageSize, setSearchParams]
  );

  // Initialize from URL on component mount (only once)
  const [isInitialized, setIsInitialized] = useState(false);

  // Check for action parameter to auto-open create sheet
  useEffect(() => {
    const actionParam = searchParams.get("action");
    if (actionParam === "create") {
      setIsCreateSheetOpen(true);
      // Remove action parameter from URL to keep it clean
      const newSearchParams = new URLSearchParams(searchParams);
      newSearchParams.delete("action");
      setSearchParams(newSearchParams, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  useEffect(() => {
    // Get URL params directly to avoid dependency issues
    const nameParam = searchParams.get("name");
    const protocolParam = searchParams.get("protocol");
    const pageParam = searchParams.get("page");
    const sortByParam = searchParams.get("sort_by");
    const sortDirectionParam = searchParams.get("sort_direction");

    const initialFilters: ColumnFiltersState = [];

    if (nameParam) {
      initialFilters.push({ id: "name", value: nameParam });
    }

    if (protocolParam) {
      initialFilters.push({
        id: "protocol",
        value: protocolParam as GitProtocolType,
      });
    }

    const initialPage = pageParam ? parseInt(pageParam, 10) : 1;

    // Initialize sorting from URL
    const initialSorting: SortingState = [];
    if (sortByParam) {
      initialSorting.push({
        id: sortByParam,
        desc: sortDirectionParam === 'desc'
      });
    }

    // Set state first
    setColumnFilters(initialFilters);
    setCurrentPage(initialPage);
    setSorting(initialSorting);

    // Load initial data using the unified function
    loadProjectsData(initialPage, initialFilters, initialSorting, false).then(() => {
      setIsInitialized(true);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty dependency array - only run once on mount

  // Handle filter and sorting changes (skip initial load)
  useEffect(() => {
    if (isInitialized) {
      loadProjectsData(1, columnFilters, sorting); // Reset to page 1 when filtering or sorting
    }
  }, [columnFilters, sorting, isInitialized]); // Combined effect for both filter and sorting changes

  const handlePageChange = useCallback(
    (page: number) => {
      loadProjectsData(page, columnFilters, sorting);
    },
    [columnFilters, sorting, loadProjectsData]
  );

  // Set page actions (Create button in header) and clear breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      setIsCreateSheetOpen(true);
    };

    // Only show create button if user has permission
    if (canCreateProject) {
      setActions(
        <Button onClick={handleCreateNew} size="sm">
          <Plus className="h-4 w-4 mr-2" />
          {t("projects.create")}
        </Button>
      );
    } else {
      setActions(null);
    }

    // Clear breadcrumb items (we're at the root level)
    setItems([]);

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [setActions, setItems, t, canCreateProject]);

  const handleEdit = useCallback(
    async (project: Project) => {
      try {
        // Fetch complete project details from API to ensure we have all fields
        // The list data only contains ProjectListItemResponse which may be missing fields like system_prompt, credential_id
        const response = await apiService.projects.get(project.id);
        setEditingProject(response.project);
        setIsEditSheetOpen(true);
      } catch (error) {
        logError(error as Error, "Failed to load project details for editing");
        toast.error(t("projects.messages.loadDetailsFailed"));
      }
    },
    [t]
  );

  const handleDelete = useCallback(
    async (id: number) => {
      try {
        await apiService.projects.delete(id);
        await loadProjectsData(currentPage, columnFilters, sorting);
      } catch (error) {
        // Re-throw error to let QuickActions handle the user notification
        throw error;
      }
    },
    [loadProjectsData, currentPage, columnFilters, sorting]
  );



  const handleKanban = useCallback(
    (project: Project) => {
      navigate(`/projects/${project.id}/kanban`);
    },
    [navigate]
  );

  const handleManageAdmins = useCallback(
    (project: Project) => {
      setManagingProject(project);
    },
    []
  );

  const handleManageNotifiers = useCallback(
    (project: Project) => {
      setManagingProjectForNotifiers(project);
    },
    []
  );

  // Sheet handlers
  const handleCreateProject = async (project: Project) => {
    try {
      setIsSubmitting(true);
      // Refresh the project list
      await loadProjectsData(currentPage, columnFilters, sorting);
      // Close the sheet
      setIsCreateSheetOpen(false);
      // Show success message
      toast.success(t("projects.messages.createSuccess"));
      console.log("Project created successfully:", project);
    } catch (error) {
      console.error("Failed to create project:", error);
      logError(error as Error, "Failed to create project");
      throw error;
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleUpdateProject = async (project: Project) => {
    try {
      setIsSubmitting(true);
      // Refresh the project list
      await loadProjectsData(currentPage, columnFilters, sorting);
      // Close the sheet
      setIsEditSheetOpen(false);
      setEditingProject(null);
      // Show success message
      toast.success(t("projects.messages.updateSuccess"));
      console.log("Project updated successfully:", project);
    } catch (error) {
      console.error("Failed to update project:", error);
      logError(error as Error, "Failed to update project");
      throw error;
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCloseCreateSheet = () => {
    setIsCreateSheetOpen(false);
  };

  const handleCloseEditSheet = () => {
    setIsEditSheetOpen(false);
    setEditingProject(null);
  };

  const columns = useMemo(
    () =>
      createProjectColumns({
        t,
        onEdit: handleEdit,
        onDelete: handleDelete,
        onKanban: handleKanban,
        onManageAdmins: handleManageAdmins,
        onManageNotifiers: handleManageNotifiers,
        canEditProject,
        canDeleteProject,
      }),
    [t, handleEdit, handleDelete, handleKanban, handleManageAdmins, handleManageNotifiers, canEditProject, canDeleteProject]
  );

  return (
    <div className="min-h-screen bg-background">
      <SectionGroup>
        <Section>
          <SectionHeader>
            <SectionTitle>{t("navigation.projects")}</SectionTitle>
            <SectionDescription>
              {t("projects.page_description")}
            </SectionDescription>
          </SectionHeader>
        </Section>
        <Section>
          <div className="space-y-4">
            <DataTable
              columns={columns}
              data={projects}
              toolbarComponent={ProjectDataTableToolbar}
              columnFilters={columnFilters}
              setColumnFilters={setColumnFilters}
              sorting={sorting}
              setSorting={setSorting}
              loading={loading}
            />
            <DataTablePaginationServer
              currentPage={currentPage}
              totalPages={totalPages}
              total={total}
              onPageChange={handlePageChange}
            />
          </div>
        </Section>
      </SectionGroup>

      {/* Create Project Sheet */}
      <FormSheet
        open={isCreateSheetOpen}
        onOpenChange={setIsCreateSheetOpen}
      >
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>{t("projects.create")}</FormSheetTitle>
            <FormSheetDescription>
              {t("projects.createDescription")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                <ProjectFormSheet
                  onSubmit={handleCreateProject}
                  onCancel={handleCloseCreateSheet}
                  formId="project-create-sheet-form"
                />
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="project-create-sheet-form"
              disabled={isSubmitting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSubmitting
                ? t("common.saving")
                : t("projects.create")}
            </Button>
          </FormSheetFooter>
        </FormSheetContent>
      </FormSheet>

      {/* Edit Project Sheet */}
      <FormSheet
        open={isEditSheetOpen}
        onOpenChange={setIsEditSheetOpen}
      >
        <FormSheetContent className="w-full sm:w-[600px] sm:max-w-[600px]">
          <FormSheetHeader>
            <FormSheetTitle>
              {t("projects.edit")} - {editingProject?.name || ""}
            </FormSheetTitle>
            <FormSheetDescription>
              {t("projects.editDescription")}
            </FormSheetDescription>
          </FormSheetHeader>
          <FormCardGroup className="overflow-y-auto">
            <FormCard className="border-none overflow-auto">
              <FormCardContent>
                {editingProject && (
                  <ProjectFormSheet
                    project={editingProject}
                    onSubmit={handleUpdateProject}
                    onCancel={handleCloseEditSheet}
                    formId="project-edit-sheet-form"
                  />
                )}
              </FormCardContent>
            </FormCard>
          </FormCardGroup>
          <FormSheetFooter>
            <Button
              type="submit"
              form="project-edit-sheet-form"
              disabled={isSubmitting}
            >
              <Save className="w-4 h-4 mr-2" />
              {isSubmitting
                ? t("common.saving")
                : t("common.save")}
            </Button>
          </FormSheetFooter>
        </FormSheetContent>
      </FormSheet>

      {/* Admin Management Sheet */}
      {managingProject && (
        <AdminManagementSheet
          project={managingProject}
          open={!!managingProject}
          onOpenChange={(open) => {
            if (!open) {
              setManagingProject(null);
            }
          }}
          onAdminChanged={() => {
            // Optionally reload projects data to reflect changes
            loadProjectsData(currentPage, columnFilters, sorting);
          }}
        />
      )}

      {/* Notifier Management Sheet */}
      {managingProjectForNotifiers && (
        <NotifierManagementSheet
          project={managingProjectForNotifiers}
          open={!!managingProjectForNotifiers}
          onOpenChange={(open) => {
            if (!open) {
              setManagingProjectForNotifiers(null);
            }
          }}
          onNotifierChanged={() => {
            // Optionally reload projects data to reflect changes
            loadProjectsData(currentPage, columnFilters, sorting);
          }}
        />
      )}
    </div>
  );
};

export default ProjectListPage;
