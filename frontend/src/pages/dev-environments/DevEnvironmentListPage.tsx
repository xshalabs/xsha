import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { useBreadcrumb } from "@/contexts/BreadcrumbContext";
import { usePageActions } from "@/contexts/PageActionsContext";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Plus, HardDrive, Cpu, Server } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { toast } from "sonner";
import {
  Section,
  SectionDescription,
  SectionGroup,
  SectionHeader,
  SectionTitle,
} from "@/components/content/section";
import {
  MetricCardGroup,
  MetricCardHeader,
  MetricCardTitle,
  MetricCardValue,
  MetricCardButton,
} from "@/components/metric/metric-card";
import { DataTable } from "@/components/ui/data-table/data-table";
import { DataTablePaginationSimple } from "@/components/ui/data-table/data-table-pagination";
import { createDevEnvironmentColumns } from "@/components/data-table/dev-environments/columns";
import { DevEnvironmentDataTableToolbar } from "@/components/data-table/dev-environments/data-table-toolbar";
import type {
  DevEnvironment,
  DevEnvironmentDisplay,
  DevEnvironmentListParams,
} from "@/types/dev-environment";
import type { ColumnFiltersState, SortingState } from "@tanstack/react-table";

const DevEnvironmentListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { setItems } = useBreadcrumb();
  const { setActions } = usePageActions();
  
  usePageTitle(t("navigation.dev_environments"));

  // Set page actions (Create button in header) and clear breadcrumb
  useEffect(() => {
    const handleCreateNew = () => {
      navigate("/dev-environments/create");
    };

    setActions(
      <Button onClick={handleCreateNew} size="sm">
        <Plus className="h-4 w-4 mr-2" />
        {t("devEnvironments.create")}
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

  const [environments, setEnvironments] = useState<DevEnvironmentDisplay[]>([]);
  const [stats, setStats] = useState<Record<string, number>>({});
  const [loading, setLoading] = useState(false);
  const [listParams, setListParams] = useState<DevEnvironmentListParams>({
    page: 1,
    page_size: 10,
  });
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [environmentToDelete, setEnvironmentToDelete] = useState<number | null>(
    null
  );
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [sorting, setSorting] = useState<SortingState>([]);

  const transformEnvironments = (
    envs: DevEnvironment[]
  ): DevEnvironmentDisplay[] => {
    return envs.map((env) => {
      let envVarsMap: Record<string, string> = {};
      try {
        if (env.env_vars) {
          envVarsMap = JSON.parse(env.env_vars);
        }
      } catch (error) {
        console.warn(
          "Failed to parse env_vars for environment:",
          env.id,
          error
        );
      }

      return {
        ...env,
        env_vars_map: envVarsMap,
      };
    });
  };

  const fetchEnvironments = async (
    params: DevEnvironmentListParams = listParams
  ) => {
    setLoading(true);
    try {
      const response = await apiService.devEnvironments.list(params);
      const transformedEnvironments = transformEnvironments(
        response.environments
      );
      setEnvironments(transformedEnvironments);
      setTotalPages(response.total_pages);
      setTotal(response.total);
    } catch (error) {
      logError(error as Error, "Failed to fetch environments");
      toast.error(
        error instanceof Error
          ? error.message
          : t("devEnvironments.fetch_failed")
      );
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await apiService.devEnvironments.getStats();
      setStats(response.stats);
    } catch (error) {
      logError(error as Error, "Failed to fetch environment stats");
    }
  };

  const formatMemory = (mb: number) => {
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(1)} GB`;
    }
    return `${mb} MB`;
  };

  const formatCPU = (cores: number) => {
    return cores === 1 ? `${cores} core` : `${cores} cores`;
  };

  const metrics = [
    {
      title: t("devEnvironments.stats.total"),
      value: stats.total || 0,
      variant: "default" as const,
      icon: Server,
    },
    {
      title: t("devEnvironments.stats.total_cpu"),
      value: formatCPU((stats.total_cpu || 0) / 100),
      variant: "ghost" as const,
      icon: Cpu,
    },
    {
      title: t("devEnvironments.stats.total_memory"),
      value: formatMemory(stats.total_memory || 0),
      variant: "ghost" as const,
      icon: HardDrive,
    },
  ];

  const handleDeleteEnvironment = (id: number) => {
    setEnvironmentToDelete(id);
    setDeleteDialogOpen(true);
  };



  const handleConfirmDelete = async () => {
    if (!environmentToDelete) return;

    try {
      await apiService.devEnvironments.delete(environmentToDelete);
      toast.success(t("devEnvironments.delete_success"));
      await fetchEnvironments();
      await fetchStats();
    } catch (error) {
      logError(error as Error, "Failed to delete environment");
      toast.error(
        error instanceof Error
          ? error.message
          : t("devEnvironments.delete_failed")
      );
    } finally {
      setDeleteDialogOpen(false);
      setEnvironmentToDelete(null);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setEnvironmentToDelete(null);
  };

  const handleEditEnvironment = (environment: DevEnvironmentDisplay) => {
    navigate(`/dev-environments/${environment.id}/edit`);
  };

  const columns = createDevEnvironmentColumns({
    onEdit: handleEditEnvironment,
    onDelete: handleDeleteEnvironment,
  });

  useEffect(() => {
    fetchEnvironments();
    fetchStats();
  }, []);

  return (
    <div className="min-h-screen bg-background">
      <SectionGroup>
          <Section>
            <SectionHeader>
              <SectionTitle>{t("navigation.dev_environments")}</SectionTitle>
              <SectionDescription>
                {t("devEnvironments.page_description")}
              </SectionDescription>
            </SectionHeader>
            <MetricCardGroup>
              {metrics.map((metric) => {
                const Icon = metric.icon;
                return (
                  <MetricCardButton
                    key={metric.title}
                    variant={metric.variant}
                  >
                    <MetricCardHeader className="flex justify-between items-center gap-2 w-full">
                      <MetricCardTitle className="truncate">
                        {metric.title}
                      </MetricCardTitle>
                      <Icon className="size-4" />
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
            data={environments}
            toolbarComponent={DevEnvironmentDataTableToolbar}
            paginationComponent={DataTablePaginationSimple}
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
              {t("devEnvironments.delete_confirm_title")}
            </DialogTitle>
            <DialogDescription className="text-muted-foreground">
              {t("devEnvironments.delete_confirm")}
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
              {t("common.confirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default DevEnvironmentListPage;