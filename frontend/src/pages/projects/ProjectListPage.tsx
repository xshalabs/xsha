import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { usePageActions } from "@/contexts/PageActionsContext";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { DataTable } from "@/components/ui/data-table/data-table";
import { Plus, FolderGit2, CheckCircle, ListFilter } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
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
  MetricCardButton,
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
} from "@/components/metric/metric-card";
import { createProjectColumns } from "@/components/data-table/projects/columns";
import { ProjectDataTableToolbar } from "@/components/data-table/projects/data-table-toolbar";
import { DataTablePaginationI18n } from "@/components/ui/data-table/data-table-pagination-i18n";
import type { Project } from "@/types/project";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const ProjectListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [projectToDelete, setProjectToDelete] = useState<number | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);
  const { setActions } = usePageActions();
  const { setItems } = useBreadcrumb();

  usePageTitle(t("common.pageTitle.projects"));

  const loadProjectsData = async () => {
    try {
      setLoading(true);
      const response = await apiService.projects.list();
      setProjects(response.projects);
    } catch (error) {
      logError(error as Error, "Failed to load projects for metrics");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProjectsData();
  }, []);

  // Initialize filters from URL params
  useEffect(() => {
    const credential = searchParams.get("credential");
    
    const filters: ColumnFiltersState = [];
    if (credential) {
      filters.push({ id: "hasCredential", value: [credential] });
    }
    
    setColumnFilters(filters);
  }, [searchParams]);

  // Set page actions (Create button in header) and clear breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      navigate("/projects/create");
    };

    setActions(
      <Button onClick={handleCreateNew} size="sm">
        <Plus className="h-4 w-4 mr-2" />
        {t("projects.create")}
      </Button>
    );

    // Clear breadcrumb items (we're at the root level)
    setItems([]);

    // Cleanup when component unmounts
    return () => {
      setActions(null);
      setItems([]);
    };
  }, [navigate, setActions, setItems, t]);

  const metrics = [
    {
      title: t("projects.metrics.total"),
      value: projects.length,
      variant: "default" as const,
      icon: FolderGit2,
      type: "info" as const,
    },
    {
      title: t("projects.metrics.withCredentials"),
      value: projects.filter(p => p.credential_id).length,
      variant: "success" as const,
      icon: CheckCircle,
      type: "filter" as const,
      filterKey: "hasCredential",
      filterValue: "true",
    },
  ];

  const icons = {
    filter: {
      active: CheckCircle,
      inactive: ListFilter,
    },
    info: {
      active: FolderGit2,
      inactive: FolderGit2,
    },
  };

  const handleEdit = (project: Project) => {
    navigate(`/projects/${project.id}/edit`);
  };

  const handleDelete = (id: number) => {
    setProjectToDelete(id);
    setDeleteDialogOpen(true);
  };

  const handleManageTasks = (project: Project) => {
    navigate(`/projects/${project.id}/tasks`);
  };

  const handleConfirmDelete = async () => {
    if (!projectToDelete) return;

    try {
      await apiService.projects.delete(projectToDelete);
      toast.success(t("projects.messages.deleteSuccess"));
      await loadProjectsData();
    } catch (error) {
      logError(error as Error, "Failed to delete project");
      toast.error(
        error instanceof Error
          ? error.message
          : t("projects.messages.deleteFailed")
      );
    } finally {
      setDeleteDialogOpen(false);
      setProjectToDelete(null);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setProjectToDelete(null);
  };



  // Handle metric card clicks for filtering
  const handleMetricClick = (metric: typeof metrics[0]) => {
    if (metric.type !== "filter") return;

    const existingFilter = columnFilters.find(
      (filter) => filter.id === metric.filterKey
    );
    
    const isFilterActive = 
      Array.isArray(existingFilter?.value) && 
      existingFilter?.value.includes(metric.filterValue);

    if (isFilterActive) {
      // Remove filter
      setColumnFilters(prev => 
        prev.filter(filter => filter.id !== metric.filterKey)
      );
      // Update URL
      const newParams = new URLSearchParams(searchParams);
      newParams.delete(metric.filterKey!);
      setSearchParams(newParams);
    } else {
      // Add filter
      setColumnFilters(prev => {
        const others = prev.filter(filter => filter.id !== metric.filterKey);
        return [...others, { id: metric.filterKey!, value: [metric.filterValue!] }];
      });
      // Update URL
      const newParams = new URLSearchParams(searchParams);
      newParams.set(metric.filterKey!, metric.filterValue!);
      setSearchParams(newParams);
    }
  };

  const columns = createProjectColumns({
    t,
    onEdit: handleEdit,
    onDelete: handleDelete,
    onManageTasks: handleManageTasks,
  });

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <SectionGroup>
          <Section>
            <div className="flex items-center justify-center h-64">
              <div className="text-gray-500">{t("common.loading")}</div>
            </div>
          </Section>
        </SectionGroup>
      </div>
    );
  }

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
          <MetricCardGroup>
            {metrics.map((metric) => {
              const existingFilter = columnFilters.find(
                (filter) => filter.id === metric.filterKey
              );
              const isFilterActive = 
                metric.type === "filter" &&
                Array.isArray(existingFilter?.value) && 
                existingFilter?.value.includes(metric.filterValue);

              const isActive = metric.type === "filter" ? isFilterActive : false;
              const IconComponent = metric.type === "filter" 
                ? icons[metric.type][isActive ? "active" : "inactive"]
                : metric.icon;

              return (
                <MetricCardButton
                  key={metric.title}
                  variant={metric.variant}
                  onClick={() => handleMetricClick(metric)}
                  disabled={metric.type === "info"}
                >
                  <MetricCardHeader className="flex justify-between items-center gap-2 w-full">
                    <MetricCardTitle className="truncate">
                      {metric.title}
                    </MetricCardTitle>
                    <IconComponent className="size-4" />
                  </MetricCardHeader>
                  <MetricCardValue>{metric.value}</MetricCardValue>
                </MetricCardButton>
              );
            })}
          </MetricCardGroup>
        </Section>
        <Section>
          <DataTable
            columns={columns}
            data={projects}
            toolbarComponent={ProjectDataTableToolbar}
            paginationComponent={DataTablePaginationI18n}
            columnFilters={columnFilters}
            setColumnFilters={setColumnFilters}
            sorting={sorting}
            setSorting={setSorting}
          />
        </Section>
      </SectionGroup>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-foreground">
              {t("projects.delete")}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
              {t("projects.messages.deleteConfirm")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              className="text-foreground hover:text-foreground"
              onClick={handleCancelDelete}
            >
              {t("common.cancel")}
            </Button>
            <Button
              variant="destructive"
              className="text-foreground"
              onClick={handleConfirmDelete}
            >
              {t("projects.delete")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ProjectListPage;
