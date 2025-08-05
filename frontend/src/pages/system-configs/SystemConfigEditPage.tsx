import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { usePageTitle } from "@/hooks/usePageTitle";
import { systemConfigsApi } from "@/lib/api/system-configs";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { RefreshCw, Save } from "lucide-react";
import type { SystemConfig, ConfigUpdateItem } from "@/types/system-config";

export default function SystemConfigEditPage() {
  const { t } = useTranslation();
  usePageTitle(t("systemConfig.title"));

  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [formData, setFormData] = useState<{ [key: string]: string }>({});
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const fetchConfigs = async () => {
    try {
      setLoading(true);
      const response = await systemConfigsApi.listAll();
      setConfigs(response.configs);

      const initialFormData: { [key: string]: string } = {};
      response.configs.forEach((config) => {
        initialFormData[config.config_key] = config.config_value;
      });
      setFormData(initialFormData);
    } catch (error: any) {
      toast.error(error.message || t("api.operation_failed"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConfigs();
  }, []);

  const handleInputChange = (configKey: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      [configKey]: value,
    }));
  };

  const handleSwitchChange = (configKey: string, checked: boolean) => {
    setFormData((prev) => ({
      ...prev,
      [configKey]: checked ? "true" : "false",
    }));
  };

  const renderFormField = (config: SystemConfig) => {
    switch (config.form_type) {
      case 'switch':
        return (
          <div className="flex items-center space-x-2">
            <Switch
              id={config.config_key}
              checked={formData[config.config_key] === "true"}
              onCheckedChange={(checked) =>
                handleSwitchChange(config.config_key, checked)
              }
              disabled={!config.is_editable}
            />
            <Label htmlFor={config.config_key} className="text-sm">
              {formData[config.config_key] === "true" 
                ? t("common.enabled") 
                : t("common.disabled")}
            </Label>
          </div>
        );
      
      case 'textarea':
        return (
          <Textarea
            id={config.config_key}
            value={formData[config.config_key] || ""}
            onChange={(e) =>
              handleInputChange(config.config_key, e.target.value)
            }
            disabled={!config.is_editable}
            rows={4}
            className="resize-none"
            placeholder={getPlaceholder(config.config_key)}
          />
        );
      
      case 'number':
        return (
          <Input
            id={config.config_key}
            type="number"
            value={formData[config.config_key] || ""}
            onChange={(e) =>
              handleInputChange(config.config_key, e.target.value)
            }
            disabled={!config.is_editable}
            placeholder={getPlaceholder(config.config_key)}
          />
        );
      
      case 'password':
        return (
          <Input
            id={config.config_key}
            type="password"
            value={formData[config.config_key] || ""}
            onChange={(e) =>
              handleInputChange(config.config_key, e.target.value)
            }
            disabled={!config.is_editable}
            placeholder={getPlaceholder(config.config_key)}
          />
        );
      
      case 'input':
      default:
        return (
          <Input
            id={config.config_key}
            value={formData[config.config_key] || ""}
            onChange={(e) =>
              handleInputChange(config.config_key, e.target.value)
            }
            disabled={!config.is_editable}
            placeholder={getPlaceholder(config.config_key)}
          />
        );
    }
  };

  const getPlaceholder = (configKey: string) => {
    switch (configKey) {
      case "git_proxy_http":
      case "git_proxy_https":
        return t("systemConfig.proxy_url_placeholder");
      case "git_proxy_no_proxy":
        return t("systemConfig.no_proxy_placeholder");
      default:
        return "";
    }
  };

  const getConfigLabel = (config: SystemConfig) => {
    const configKeyMap = {
      'admin_user': t("systemConfig.admin_user"),
      'admin_password': t("systemConfig.admin_password"),
      'dev_environment_images': t("systemConfig.dev_environment_images"),
      'git_proxy_enabled': t("systemConfig.git_proxy_enabled"),
      'git_proxy_http': t("systemConfig.git_proxy_http"),
      'git_proxy_https': t("systemConfig.git_proxy_https"),
      'git_proxy_no_proxy': t("systemConfig.git_proxy_no_proxy"),
    } as const;
    
    return configKeyMap[config.config_key as keyof typeof configKeyMap] || config.config_key;
  };

  const getConfigDescription = (config: SystemConfig) => {
    const configDescMap = {
      'admin_user': t("systemConfig.admin_user_desc"),
      'admin_password': t("systemConfig.admin_password_desc"), 
      'dev_environment_images': t("systemConfig.dev_environment_images_description"),
      'git_proxy_enabled': t("systemConfig.git_proxy_enabled_desc"),
      'git_proxy_http': t("systemConfig.git_proxy_http_desc"),
      'git_proxy_https': t("systemConfig.git_proxy_https_desc"),
      'git_proxy_no_proxy': t("systemConfig.git_proxy_no_proxy_desc"),
    } as const;
    
    return configDescMap[config.config_key as keyof typeof configDescMap] || config.description;
  };

  const getCategoryName = (category: string) => {
    const categoryMap = {
      'auth': t("systemConfig.categories.auth"),
      'dev_environment': t("systemConfig.categories.devEnvironment"),
      'git': t("systemConfig.categories.git"),
    } as const;
    
    return categoryMap[category as keyof typeof categoryMap] || category;
  };

  const getCategoryDescription = (category: string) => {
    const descMap = {
      'auth': t("systemConfig.categoryDescriptions.auth"),
      'dev_environment': t("systemConfig.categoryDescriptions.devEnvironment"),
      'git': t("systemConfig.categoryDescriptions.git"),
    } as const;
    
    return descMap[category as keyof typeof descMap] || "";
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setSaving(true);

      const updateItems: ConfigUpdateItem[] = configs
        .filter((config) => config.is_editable)
        .map((config) => ({
          config_key: config.config_key,
          config_value: formData[config.config_key] !== undefined ? formData[config.config_key] : config.config_value,
        }));

      await systemConfigsApi.batchUpdate({ configs: updateItems });
      toast.success(t("systemConfig.update_success"));

      await fetchConfigs();
    } catch (error: any) {
      toast.error(error.message || t("api.operation_failed"));
    } finally {
      setSaving(false);
    }
  };

  const handleRefresh = () => {
    fetchConfigs();
  };

  const configsByCategory = configs.reduce((acc, config) => {
    const category = config.category || "general";
    if (!acc[category]) {
      acc[category] = [];
    }
    acc[category].push(config);
    return acc;
  }, {} as Record<string, SystemConfig[]>);

  const categories = Object.keys(configsByCategory).sort();

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center h-64">
            <RefreshCw className="w-8 h-8 animate-spin" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">
              {t("systemConfig.title")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {t("systemConfig.edit_page_description")}
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="text-foreground hover:text-foreground"
              onClick={handleRefresh}
              disabled={loading}
            >
              <RefreshCw
                className={`w-4 h-4 mr-2 ${loading ? "animate-spin" : ""}`}
              />
              {t("common.refresh")}
            </Button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <form onSubmit={handleSubmit} className="space-y-6">
          {categories.map((category) => (
            <Card key={category}>
              <CardHeader>
                <CardTitle className="capitalize">
                  {getCategoryName(category)}
                </CardTitle>
                <CardDescription>
                  {getCategoryDescription(category)}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {configsByCategory[category].map((config, index) => (
                  <div key={config.id} className="space-y-2">
                    {index > 0 && <Separator />}
                    <div className="pt-4">
                      <Label
                        htmlFor={config.config_key}
                        className="text-sm font-medium"
                      >
                        {getConfigLabel(config)}
                        {!config.is_editable && (
                          <span className="ml-2 text-xs text-muted-foreground">
                            ({t("systemConfig.readonly")})
                          </span>
                        )}
                      </Label>
                      {getConfigDescription(config) && (
                        <p className="text-xs text-muted-foreground mt-1">
                          {getConfigDescription(config)}
                        </p>
                      )}
                      <div className="mt-2">
                        {renderFormField(config)}
                      </div>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          ))}

          <div className="flex justify-end">
            <Button type="submit" disabled={saving}>
              <Save
                className={`w-4 h-4 mr-2 ${saving ? "animate-spin" : ""}`}
              />
              {saving ? t("common.saving") : t("common.save")}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
