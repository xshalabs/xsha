import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus, Settings, Monitor } from "lucide-react";
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { toast } from "sonner";
import type {
  DevEnvironment,
  DevEnvironmentDisplay,
  DevEnvironmentListParams,
} from "@/types/dev-environment";
import DevEnvironmentList from "@/components/DevEnvironmentList";

const DevEnvironmentListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  usePageTitle(t("navigation.dev_environments"));

  const [environments, setEnvironments] = useState<DevEnvironmentDisplay[]>([]);
  const [loading, setLoading] = useState(false);
  const [listParams, setListParams] = useState<DevEnvironmentListParams>({
    page: 1,
    page_size: 10,
  });
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);

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
      console.error("Failed to fetch environments:", error);
      toast.error(t("dev_environments.fetch_failed"));
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteEnvironment = async (id: number) => {
    if (!confirm(t("dev_environments.delete_confirm"))) {
      return;
    }

    try {
      await apiService.devEnvironments.delete(id);
      toast.success(t("dev_environments.delete_success"));
      await fetchEnvironments();
    } catch (error) {
      console.error("Failed to delete environment:", error);
      toast.error(t("dev_environments.delete_failed"));
    }
  };

  const handleEditEnvironment = (environment: DevEnvironmentDisplay) => {
    navigate(`/dev-environments/${environment.id}/edit`);
  };

  const handlePageChange = (page: number) => {
    const newParams = { ...listParams, page };
    setListParams(newParams);
    fetchEnvironments(newParams);
  };

  const handleFiltersChange = (filters: DevEnvironmentListParams) => {
    setListParams(filters);
    fetchEnvironments(filters);
  };

  useEffect(() => {
    fetchEnvironments();
  }, []);

  const totalCpu = environments.reduce((sum, env) => sum + env.cpu_limit, 0);
  const totalMemory = environments.reduce(
    (sum, env) => sum + env.memory_limit,
    0
  );

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("navigation.dev_environments")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("dev_environments.page_description")}
            </p>
          </div>
          <div className="flex gap-2">
            <Button onClick={() => navigate("/dev-environments/create")}>
              <Plus className="h-4 w-4 mr-2" />
              {t("dev_environments.create")}
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {t("dev_environments.stats.total")}
              </CardTitle>
              <Monitor className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{total}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {t("dev_environments.stats.cpu_total")}
              </CardTitle>
              <Settings className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalCpu.toFixed(1)}</div>
              <p className="text-xs text-muted-foreground">
                {t("dev_environments.stats.cores")}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {t("dev_environments.stats.memory_total")}
              </CardTitle>
              <Settings className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {(totalMemory / 1024).toFixed(1)}
              </div>
              <p className="text-xs text-muted-foreground">GB</p>
            </CardContent>
          </Card>
        </div>

        <DevEnvironmentList
          environments={environments}
          loading={loading}
          params={listParams}
          totalPages={totalPages}
          total={total}
          onPageChange={handlePageChange}
          onFiltersChange={handleFiltersChange}
          onRefresh={() => fetchEnvironments()}
          onEdit={handleEditEnvironment}
          onDelete={handleDeleteEnvironment}
        />
      </div>
    </div>
  );
};

export default DevEnvironmentListPage;
