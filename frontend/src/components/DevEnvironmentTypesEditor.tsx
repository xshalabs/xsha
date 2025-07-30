import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { 
  Plus,
  Trash2,
  Save,
  X,
  Edit,
  Code,
  AlertCircle
} from "lucide-react";
import type { DevEnvironmentType } from "@/types/system-config";

interface DevEnvironmentTypesEditorProps {
  types: DevEnvironmentType[];
  onUpdate: (types: DevEnvironmentType[]) => Promise<void>;
  loading?: boolean;
  readOnly?: boolean;
}

export function DevEnvironmentTypesEditor({
  types,
  onUpdate,
  loading = false,
  readOnly = false,
}: DevEnvironmentTypesEditorProps) {
  const { t } = useTranslation();
  const [isEditing, setIsEditing] = useState(false);
  const [editingTypes, setEditingTypes] = useState<DevEnvironmentType[]>(types);
  const [saving, setSaving] = useState(false);

  const handleEdit = () => {
    setIsEditing(true);
    setEditingTypes([...types]);
  };

  const handleCancel = () => {
    setIsEditing(false);
    setEditingTypes([...types]);
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      await onUpdate(editingTypes);
      setIsEditing(false);
    } catch (error) {
      // 错误处理由父组件处理
    } finally {
      setSaving(false);
    }
  };

  const handleAddType = () => {
    setEditingTypes([
      ...editingTypes,
      { name: "", image: "" }
    ]);
  };

  const handleRemoveType = (index: number) => {
    setEditingTypes(editingTypes.filter((_, i) => i !== index));
  };

  const handleTypeChange = (index: number, field: keyof DevEnvironmentType, value: string) => {
    const newTypes = [...editingTypes];
    newTypes[index] = { ...newTypes[index], [field]: value };
    setEditingTypes(newTypes);
  };

  const validateTypes = () => {
    return editingTypes.every(type => 
      type.name.trim() !== "" && type.image.trim() !== ""
    ) && editingTypes.length > 0;
  };

  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="flex items-center gap-2 text-lg">
              <Code className="w-5 h-5" />
              {t("system-config.dev_environment_types")}
            </CardTitle>
            <p className="text-sm text-muted-foreground mt-1">
              {t("system-config.dev_environment_types_description")}
            </p>
          </div>
          
          {!readOnly && !isEditing && (
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
        {isEditing ? (
          <>
            {/* 编辑模式 */}
            <div className="space-y-4">
              {editingTypes.map((type, index) => (
                <div key={index} className="p-4 border border-gray-200 rounded-lg">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor={`name-${index}`} className="text-sm font-medium">
                        {t("system-config.env_type_name")} *
                      </Label>
                      <Input
                        id={`name-${index}`}
                        value={type.name}
                        onChange={(e) => handleTypeChange(index, "name", e.target.value)}
                        placeholder={t("system-config.env_type_name_placeholder")}
                        className="mt-1"
                      />
                    </div>
                    <div>
                      <Label htmlFor={`image-${index}`} className="text-sm font-medium">
                        {t("system-config.env_type_image")} *
                      </Label>
                      <div className="flex gap-2 mt-1">
                        <Input
                          id={`image-${index}`}
                          value={type.image}
                          onChange={(e) => handleTypeChange(index, "image", e.target.value)}
                          placeholder={t("system-config.env_type_image_placeholder")}
                          className="flex-1"
                        />
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => handleRemoveType(index)}
                          disabled={editingTypes.length <= 1}
                          className="px-3"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* 添加新类型 */}
            <Button
              type="button"
              variant="outline"
              onClick={handleAddType}
              className="w-full"
            >
              <Plus className="w-4 h-4 mr-2" />
              {t("system-config.add_env_type")}
            </Button>

            {/* 验证提示 */}
            {!validateTypes() && (
              <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-md">
                <AlertCircle className="w-4 h-4 text-red-600" />
                <p className="text-sm text-red-700">
                  {t("system-config.env_types_validation_error")}
                </p>
              </div>
            )}

            {/* 编辑操作 */}
            <div className="flex items-center justify-end space-x-2 pt-4 border-t">
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
                disabled={saving || !validateTypes()}
              >
                <Save className="w-4 h-4 mr-2" />
                {saving ? t("common.saving") : t("common.save")}
              </Button>
            </div>
          </>
        ) : (
          <>
            {/* 显示模式 */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {types.map((type, index) => (
                <div key={index} className="p-4 bg-gray-50 rounded-lg border">
                  <div className="flex items-center justify-between mb-2">
                    <Badge variant="secondary" className="font-medium">
                      {type.name}
                    </Badge>
                  </div>
                  <div className="text-sm text-muted-foreground">
                    <div className="font-mono text-xs">{type.image}</div>
                  </div>
                </div>
              ))}
            </div>

            {readOnly && (
              <div className="flex items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
                <AlertCircle className="w-4 h-4 text-yellow-600" />
                <p className="text-sm text-yellow-700">
                  {t("system-config.readonly_config")}
                </p>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
} 