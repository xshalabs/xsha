import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { 
  Edit, 
  Save, 
  X, 
  Lock,
  Settings,
  AlertCircle
} from "lucide-react";
import type {
  SystemConfig,
  UpdateSystemConfigRequest,
} from "@/types/system-config";

interface ConfigCardProps {
  config: SystemConfig;
  onUpdate: (id: number, data: UpdateSystemConfigRequest) => Promise<void>;
  loading?: boolean;
}

export function ConfigCard({
  config,
  onUpdate,
  loading = false,
}: ConfigCardProps) {
  const { t } = useTranslation();
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    config_value: config.config_value,
    description: config.description,
  });
  const [saving, setSaving] = useState(false);

  const handleEdit = () => {
    setIsEditing(true);
    setFormData({
      config_value: config.config_value,
      description: config.description,
    });
  };

  const handleCancel = () => {
    setIsEditing(false);
    setFormData({
      config_value: config.config_value,
      description: config.description,
    });
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      
      // 只发送有更改的字段
      const updateData: UpdateSystemConfigRequest = {};
      if (formData.config_value !== config.config_value) {
        updateData.config_value = formData.config_value;
      }
      if (formData.description !== config.description) {
        updateData.description = formData.description;
      }

      await onUpdate(config.id, updateData);
      setIsEditing(false);
    } catch (error) {
      // 错误处理由父组件处理
    } finally {
      setSaving(false);
    }
  };

  const handleChange = (field: string, value: string) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 格式化配置值显示
  const formatConfigValue = (value: string) => {
    try {
      // 尝试解析JSON并格式化显示
      const parsed = JSON.parse(value);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return value;
    }
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case 'development': return 'bg-blue-100 text-blue-800';
      case 'database': return 'bg-green-100 text-green-800';
      case 'auth': return 'bg-red-100 text-red-800';
      case 'system': return 'bg-purple-100 text-purple-800';
      case 'ui': return 'bg-orange-100 text-orange-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="flex items-center gap-2 text-lg">
              <Settings className="w-5 h-5" />
              {config.config_key}
              {!config.is_editable && (
                <Lock className="w-4 h-4 text-muted-foreground" />
              )}
            </CardTitle>
            <div className="flex items-center gap-2 mt-2">
              <Badge className={getCategoryColor(config.category)}>
                {config.category}
              </Badge>
              {!config.is_editable && (
                <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                  <Lock className="w-3 h-3 mr-1" />
                  {t("system-config.readonly")}
                </Badge>
              )}
            </div>
          </div>
          
          {config.is_editable && !isEditing && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleEdit}
              disabled={loading}
            >
              <Edit className="w-4 h-4 mr-2" />
              {t("common.edit")}
            </Button>
          )}
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Configuration Value */}
        <div>
          <Label className="text-sm font-medium">
            {t("system-config.config_value")}
          </Label>
          {isEditing ? (
            <Textarea
              value={formData.config_value}
              onChange={(e) => handleChange("config_value", e.target.value)}
              placeholder={t("system-config.config_value_placeholder")}
              rows={6}
              className="mt-2 font-mono text-sm"
            />
          ) : (
            <div className="mt-2 p-3 bg-gray-50 rounded-md border">
              <pre className="whitespace-pre-wrap text-sm font-mono">
                {formatConfigValue(config.config_value)}
              </pre>
            </div>
          )}
        </div>

        {/* Description */}
        <div>
          <Label className="text-sm font-medium">
            {t("system-config.description")}
          </Label>
          {isEditing ? (
            <Textarea
              value={formData.description}
              onChange={(e) => handleChange("description", e.target.value)}
              placeholder={t("system-config.description_placeholder")}
              rows={2}
              className="mt-2"
            />
          ) : (
            <div className="mt-2 p-3 bg-gray-50 rounded-md border min-h-[60px]">
              <p className="text-sm text-muted-foreground">
                {config.description || t("system-config.no_description")}
              </p>
            </div>
          )}
        </div>

        {/* 只读配置提示 */}
        {!config.is_editable && (
          <div className="flex items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
            <AlertCircle className="w-4 h-4 text-yellow-600" />
            <p className="text-sm text-yellow-700">
              {t("system-config.readonly_config")}
            </p>
          </div>
        )}

        {/* Edit Actions */}
        {isEditing && (
          <div className="flex items-center justify-end space-x-2 pt-2 border-t">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleCancel}
              disabled={saving}
            >
              <X className="w-4 h-4 mr-2" />
              {t("common.cancel")}
            </Button>
            <Button
              type="button"
              size="sm"
              onClick={handleSave}
              disabled={saving}
            >
              <Save className="w-4 h-4 mr-2" />
              {saving ? t("common.saving") : t("common.save")}
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
} 