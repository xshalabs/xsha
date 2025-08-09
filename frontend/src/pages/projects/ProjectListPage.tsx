import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { usePageActions } from "@/contexts/PageActionsContext";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";

import { DataTable } from "@/components/ui/data-table/data-table";
import { Plus } from "lucide-react";
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
  const [searchParams] = useSearchParams();

  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);
  const { setActions } = usePageActions();
  const { setItems } = useBreadcrumb();

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 10;

  usePageTitle(t("common.pageTitle.projects"));

  const loadProjectsData = async (
    page = currentPage,
    filters = columnFilters
  ) => {
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

      const response = await apiService.projects.list(apiParams);
      setProjects(response.projects);
      setTotal(response.total);
      setTotalPages(response.total_pages);
      setCurrentPage(page);
    } catch (error) {
      logError(error as Error, "Failed to load projects");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProjectsData().then(() => setIsInitialized(true));
  }, []);

  // Handle column filter changes (skip initial empty state)
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    if (isInitialized) {
      loadProjectsData(1, columnFilters); // Reset to page 1 when filtering
    }
  }, [columnFilters, isInitialized]);

  const handlePageChange = (page: number) => {
    loadProjectsData(page);
  };

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



  const handleEdit = (project: Project) => {
    navigate(`/projects/${project.id}/edit`);
  };

  const handleDelete = async (id: number) => {
    try {
      await apiService.projects.delete(id);
      await loadProjectsData();
    } catch (error) {
      // Re-throw error to let QuickActions handle the user notification
      throw error;
    }
  };

  const handleManageTasks = (project: Project) => {
    navigate(`/projects/${project.id}/tasks`);
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
    </div>
  );
};

export default ProjectListPage;
