import { useState, useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { usePageTitle } from "@/hooks/usePageTitle";
import { ConfigCategoryNav } from "@/components/ConfigCategoryNav";
import { ConfigCard } from "@/components/ConfigCard";
import { DevEnvironmentTypesEditor } from "@/components/DevEnvironmentTypesEditor";
import { systemConfigsApi } from "@/lib/api/system-configs";
import { Button } from "@/components/ui/button";
import { RefreshCw, Settings } from "lucide-react";
import type {
  SystemConfig,
  UpdateSystemConfigRequest,
  DevEnvironmentType,
} from "@/types/system-config";

export default function SystemConfigEditPage() {
  const { t } = useTranslation();
  usePageTitle(t("system-config.title"));

  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeCategory, setActiveCategory] = useState("development");
  const [devEnvTypes, setDevEnvTypes] = useState<DevEnvironmentType[]>([]);

  // 获取配置数据
  const fetchConfigs = async () => {
    try {
      setLoading(true);
      const response = await systemConfigsApi.list();
      setConfigs(response.configs);
    } catch (error: any) {
      toast.error(error.message || t("api.operation_failed"));
    } finally {
      setLoading(false);
    }
  };

  // 获取开发环境类型
  const fetchDevEnvTypes = async () => {
    try {
      const response = await systemConfigsApi.getDevEnvironmentTypes();
      setDevEnvTypes(response.env_types);
    } catch (error: any) {
      toast.error(error.message || t("system-config.get_types_failed"));
    }
  };

  useEffect(() => {
    fetchConfigs();
    fetchDevEnvTypes();
  }, []);

  // 按分类分组配置
  const configsByCategory = useMemo(() => {
    const grouped: Record<string, SystemConfig[]> = {};
    configs.forEach((config) => {
      const category = config.category || "general";
      if (!grouped[category]) {
        grouped[category] = [];
      }
      grouped[category].push(config);
    });
    return grouped;
  }, [configs]);

  // 生成分类导航数据
  const categories = useMemo(() => {
    const categoryKeys = Object.keys(configsByCategory);
    return categoryKeys.map((key) => ({
      key,
      name: t(`system-config.category_${key}`, key),
      description: t(`system-config.category_${key}_desc`, ""),
      icon: undefined as any, // 将在组件内部分配
      count: configsByCategory[key]?.length || 0,
    }));
  }, [configsByCategory, t]);

  // 处理配置更新
  const handleConfigUpdate = async (id: number, data: UpdateSystemConfigRequest) => {
    try {
      await systemConfigsApi.update(id, data);
      toast.success(t("system-config.update_success"));
      // 更新本地状态
      setConfigs((prev) =>
        prev.map((config) =>
          config.id === id
            ? {
                ...config,
                ...data,
                updated_at: new Date().toISOString(),
              }
            : config
        )
      );
    } catch (error: any) {
      toast.error(error.message || t("api.operation_failed"));
      throw error; // 重新抛出错误以便组件处理
    }
  };

  // 处理开发环境类型更新
  const handleDevEnvTypesUpdate = async (types: DevEnvironmentType[]) => {
    try {
      await systemConfigsApi.updateDevEnvironmentTypes({ env_types: types });
      toast.success(t("system-config.update_dev_env_types_success"));
      setDevEnvTypes(types);
    } catch (error: any) {
      toast.error(error.message || t("system-config.update_dev_env_types_failed"));
      throw error;
    }
  };

  // 刷新数据
  const handleRefresh = () => {
    fetchConfigs();
    fetchDevEnvTypes();
  };

  // 获取当前分类的配置
  const currentConfigs = configsByCategory[activeCategory] || [];

  return (
    <div className="container mx-auto py-6">
      {/* 页面头部 */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
              <Settings className="w-8 h-8" />
              {t("system-config.title")}
            </h1>
            <p className="text-muted-foreground mt-1">
              {t("system-config.edit_page_description")}
            </p>
          </div>
          <Button variant="outline" onClick={handleRefresh} disabled={loading}>
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            {t("common.refresh")}
          </Button>
        </div>
      </div>

      <div className="flex gap-6">
        {/* 侧边栏 - 分类导航 */}
        <div className="flex-shrink-0">
          <ConfigCategoryNav
            categories={categories}
            activeCategory={activeCategory}
            onCategoryChange={setActiveCategory}
          />
        </div>

        {/* 主内容区 */}
        <div className="flex-1 space-y-6">
          {/* 开发环境类型特殊处理 */}
          {activeCategory === "development" && devEnvTypes.length > 0 && (
            <DevEnvironmentTypesEditor
              types={devEnvTypes}
              onUpdate={handleDevEnvTypesUpdate}
              loading={loading}
            />
          )}

          {/* 常规配置卡片 */}
          {currentConfigs.length > 0 ? (
            <div className="space-y-4">
              {currentConfigs.map((config) => (
                <ConfigCard
                  key={config.id}
                  config={config}
                  onUpdate={handleConfigUpdate}
                  loading={loading}
                />
              ))}
            </div>
          ) : !loading ? (
            <div className="text-center py-12">
              <Settings className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                {t("system-config.no_configs")}
              </h3>
              <p className="text-sm text-muted-foreground">
                {t("system-config.no_configs_description")}
              </p>
            </div>
          ) : null}

          {/* 加载状态 */}
          {loading && currentConfigs.length === 0 && (
            <div className="space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-48 bg-gray-100 rounded-lg animate-pulse" />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
} 