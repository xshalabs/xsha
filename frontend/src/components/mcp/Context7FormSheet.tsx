import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, Globe, Key } from "lucide-react";
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
import { apiService } from "@/lib/api/index";
import {
  generateContext7Config,
  type Context7Config,
} from "@/lib/mcp/templateGenerators";
import { logError, handleApiError } from "@/lib/errors";
import { toast } from "sonner";

interface Context7FormSheetProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function Context7FormSheet({
  isOpen,
  onClose,
  onSuccess,
}: Context7FormSheetProps) {
  const { t } = useTranslation();

  const [formData, setFormData] = useState<Context7Config & { name: string }>({
    name: "context7",
    url: "https://mcp.context7.com/mcp",
    apiKey: "",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Reset form when sheet opens
  useEffect(() => {
    if (isOpen) {
      setFormData({
        name: "context7",
        url: "https://mcp.context7.com/mcp",
        apiKey: "",
      });
      setError("");
      setErrors({});
    }
  }, [isOpen]);

  const handleInputChange = (field: keyof typeof formData, value: string) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear field error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({
        ...prev,
        [field]: "",
      }));
    }

    // Clear general error
    if (error) {
      setError("");
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t("mcp.templates.context7.form.fields.name.required");
    }

    if (!formData.url.trim()) {
      newErrors.url = t("mcp.templates.context7.form.fields.url.required");
    } else {
      try {
        new URL(formData.url);
      } catch {
        newErrors.url = t("mcp.templates.context7.form.fields.url.invalid");
      }
    }

    if (!formData.apiKey.trim()) {
      newErrors.apiKey = t(
        "mcp.templates.context7.form.fields.apiKey.required"
      );
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);

      // Generate the configuration JSON
      const config = generateContext7Config({
        url: formData.url.trim(),
        apiKey: formData.apiKey.trim(),
      });

      const payload = {
        name: formData.name.trim(),
        description: "",
        config,
        enabled: true,
      };

      await apiService.mcp.create(payload);
      toast.success(t("mcp.createSuccess"));
      onSuccess();
      onClose();
    } catch (error) {
      logError(error, "Failed to create Context7 MCP configuration");
      const errorMessage = handleApiError(error);
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <FormSheet open={isOpen} onOpenChange={onClose}>
      <FormSheetContent className="w-[600px] sm:max-w-[600px]">
        <FormSheetHeader>
          <FormSheetTitle className="flex items-center gap-2">
            <Globe className="h-5 w-5" />
            {t("mcp.templates.context7.form.title")}
          </FormSheetTitle>
          <FormSheetDescription>
            {t("mcp.templates.context7.form.description")}
          </FormSheetDescription>
        </FormSheetHeader>

        <FormCardGroup className="overflow-y-auto">
          <FormCard className="border-none overflow-auto">
            <FormCardContent>
              <form onSubmit={handleSubmit} className="my-4 space-y-6">
                {error && (
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}

                <div className="space-y-6">
                  {/* Basic Information */}
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="name">
                        {t("mcp.templates.context7.form.fields.name.label")}{" "}
                        <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="name"
                        type="text"
                        placeholder={t(
                          "mcp.templates.context7.form.fields.name.placeholder"
                        )}
                        value={formData.name}
                        onChange={(e) =>
                          handleInputChange("name", e.target.value)
                        }
                        className={
                          errors.name
                            ? "border-red-500 focus-visible:ring-red-500"
                            : ""
                        }
                      />
                      {errors.name && (
                        <p className="text-sm text-red-500 flex items-center gap-1">
                          <AlertCircle className="h-3 w-3" />
                          {errors.name}
                        </p>
                      )}
                    </div>
                  </div>

                  {/* Context7 Configuration */}
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="url">
                        {t("mcp.templates.context7.form.fields.url.label")}{" "}
                        <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="url"
                        type="url"
                        placeholder={t(
                          "mcp.templates.context7.form.fields.url.placeholder"
                        )}
                        value={formData.url}
                        onChange={(e) =>
                          handleInputChange("url", e.target.value)
                        }
                        className={
                          errors.url
                            ? "border-red-500 focus-visible:ring-red-500"
                            : ""
                        }
                      />
                      {errors.url && (
                        <p className="text-sm text-red-500 flex items-center gap-1">
                          <AlertCircle className="h-3 w-3" />
                          {errors.url}
                        </p>
                      )}
                      <p className="text-xs text-muted-foreground">
                        {t("mcp.templates.context7.form.fields.url.help")}
                      </p>
                    </div>

                    <div className="space-y-2">
                      <Label
                        htmlFor="apiKey"
                        className="flex items-center gap-1"
                      >
                        <Key className="h-3 w-3" />
                        {t(
                          "mcp.templates.context7.form.fields.apiKey.label"
                        )}{" "}
                        <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="apiKey"
                        type="password"
                        placeholder={t(
                          "mcp.templates.context7.form.fields.apiKey.placeholder"
                        )}
                        value={formData.apiKey}
                        onChange={(e) =>
                          handleInputChange("apiKey", e.target.value)
                        }
                        className={
                          errors.apiKey
                            ? "border-red-500 focus-visible:ring-red-500"
                            : ""
                        }
                      />
                      {errors.apiKey && (
                        <p className="text-sm text-red-500 flex items-center gap-1">
                          <AlertCircle className="h-3 w-3" />
                          {errors.apiKey}
                        </p>
                      )}
                      <p className="text-xs text-muted-foreground">
                        {t("mcp.templates.context7.form.fields.apiKey.help")}
                      </p>
                    </div>
                  </div>
                </div>
              </form>
            </FormCardContent>
          </FormCard>
        </FormCardGroup>

        <FormSheetFooter>
          <Button type="submit" disabled={loading} onClick={handleSubmit}>
            {loading
              ? t("mcp.form.submitting")
              : t("mcp.templates.context7.form.submit")}
          </Button>
        </FormSheetFooter>
      </FormSheetContent>
    </FormSheet>
  );
}
