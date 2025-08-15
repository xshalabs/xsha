import { useTranslation } from "react-i18next";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AlertCircle, Container, Loader2 } from "lucide-react";
import type { EnvironmentImageOption } from "@/hooks/useEnvironmentForm";

interface DockerImageSelectorProps {
  dockerImage: string;
  environmentImages: EnvironmentImageOption[];
  onDockerImageChange: (dockerImage: string) => void;
  loadingImages: boolean;
  error?: string;
  disabled?: boolean;
}

export function DockerImageSelector({
  dockerImage,
  environmentImages,
  onDockerImageChange,
  loadingImages,
  error,
  disabled = false,
}: DockerImageSelectorProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Container className="h-4 w-4 text-muted-foreground" />
        <Label className="text-sm font-medium">
          {t("devEnvironments.form.docker_image")} <span className="text-red-500">*</span>
        </Label>
        {loadingImages && (
          <Loader2 className="h-3 w-3 animate-spin text-blue-500" />
        )}
      </div>
      <Select
        value={dockerImage}
        onValueChange={onDockerImageChange}
        disabled={loadingImages || disabled}
      >
        <SelectTrigger>
          <SelectValue
            placeholder={
              loadingImages
                ? t("common.loading") + "..."
                : t("devEnvironments.form.docker_image_placeholder")
            }
          />
        </SelectTrigger>
        <SelectContent className="max-w-[400px]">
          {loadingImages ? (
            <SelectItem value="loading" disabled>
              <div className="flex items-center gap-2">
                <Loader2 className="h-3 w-3 animate-spin" />
                {t("common.loading")}...
              </div>
            </SelectItem>
          ) : environmentImages.length === 0 ? (
            <SelectItem value="empty" disabled>
              {t("devEnvironments.noImagesAvailable")}
            </SelectItem>
          ) : (
            environmentImages.map((imgOption) => (
              <SelectItem key={imgOption.image} value={imgOption.image}>
                <div className="flex items-center justify-between w-full">
                  <span className="font-medium truncate">{imgOption.name}</span>
                  <span className="text-xs text-muted-foreground ml-2">
                    {imgOption.type}
                  </span>
                </div>
              </SelectItem>
            ))
          )}
        </SelectContent>
      </Select>
      {error && (
        <p className="text-sm text-red-500 flex items-center gap-1">
          <AlertCircle className="h-3 w-3" />
          {error}
        </p>
      )}
    </div>
  );
}
