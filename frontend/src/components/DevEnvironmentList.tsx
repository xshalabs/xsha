import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import {
  MoreHorizontal,
  Edit,
  Trash2,
  Monitor,
  Filter,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
} from "lucide-react";
import type {
  DevEnvironmentDisplay,
  DevEnvironmentListParams,
  DevEnvironmentType,
  DevEnvironmentTypeConfig,
} from "@/types/dev-environment";
import { devEnvironmentsApi } from "@/lib/api/dev-environments";

interface DevEnvironmentListProps {
  environments: DevEnvironmentDisplay[];
  loading: boolean;
  params: DevEnvironmentListParams;
  totalPages: number;
  total: number;
  onPageChange: (page: number) => void;
  onFiltersChange: (filters: DevEnvironmentListParams) => void;
  onRefresh: () => void;
  onEdit: (environment: DevEnvironmentDisplay) => void;
  onDelete: (id: number) => void;
}

const defaultColors = [
  "text-blue-600",
  "text-purple-600",
  "text-green-600",
  "text-orange-600",
  "text-red-600",
];

const getTypeColor = (index: number) => {
  return defaultColors[index % defaultColors.length];
};

const DevEnvironmentList: React.FC<DevEnvironmentListProps> = ({
  environments,
  loading,
  params,
  totalPages,
  total,
  onPageChange,
  onFiltersChange,
  onRefresh,
  onEdit,
  onDelete,
}) => {
  const { t } = useTranslation();
  const [showFilters, setShowFilters] = useState(false);
  const [environmentTypes, setEnvironmentTypes] = useState<
    DevEnvironmentTypeConfig[]
  >([]);
  const [typeConfigMap, setTypeConfigMap] = useState<
    Record<string, { label: string; color: string }>
  >({});
  const [localFilters, setLocalFilters] =
    useState<DevEnvironmentListParams>(params);

  const handleFilterChange = (
    key: keyof DevEnvironmentListParams,
    value: string | number | undefined
  ) => {
    setLocalFilters((prev) => ({
      ...prev,
      [key]: value === "" ? undefined : value,
    }));
  };

  const applyFilters = () => {
    onFiltersChange({
      ...localFilters,
      page: 1,
    });
  };

  useEffect(() => {
    const loadEnvironmentTypes = async () => {
      try {
        const response = await devEnvironmentsApi.getAvailableTypes();
        setEnvironmentTypes(response.types);

        const configMap: Record<string, { label: string; color: string }> = {};
        response.types.forEach((type, index) => {
          configMap[type.key] = {
            label: type.name,
            color: getTypeColor(index),
          };
        });
        setTypeConfigMap(configMap);
      } catch (error) {
        console.error("Failed to load environment types:", error);
        setTypeConfigMap({
          "claude-code": { label: "Claude Code", color: "text-blue-600" },
        });
      }
    };

    loadEnvironmentTypes();
  }, []);

  const resetFilters = () => {
    const emptyFilters: DevEnvironmentListParams = { page: 1, page_size: 10 };
    setLocalFilters(emptyFilters);
    onFiltersChange(emptyFilters);
  };

  const formatMemory = (mb: number) => {
    if (mb >= 1024) {
      return `${(mb / 1024).toFixed(1)} GB`;
    }
    return `${mb} MB`;
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div className="text-sm text-foreground">
          {t("common.total")} {total} {t("common.items")}
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="ghost"
            className="text-foreground"
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter className="w-4 h-4 mr-2" />
            {t("common.filter")}
          </Button>
          <Button
            onClick={onRefresh}
            disabled={loading}
            size="sm"
            variant="ghost"
            className="text-foreground"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            {t("common.refresh")}
          </Button>
        </div>
      </div>

      {showFilters && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              {t("dev_environments.filters.title")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex flex-col gap-3">
                <Label htmlFor="name">
                  {t("dev_environments.filters.name")}
                </Label>
                <Input
                  id="name"
                  value={localFilters.name || ""}
                  onChange={(e) => handleFilterChange("name", e.target.value)}
                  placeholder={t("dev_environments.filters.name_placeholder")}
                />
              </div>

              <div className="flex flex-col gap-3">
                <Label htmlFor="type">
                  {t("dev_environments.filters.type")}
                </Label>
                <Select
                  value={localFilters.type || "all"}
                  onValueChange={(value) =>
                    handleFilterChange(
                      "type",
                      value === "all"
                        ? undefined
                        : (value as DevEnvironmentType)
                    )
                  }
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder={t("common.all")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("common.all")}</SelectItem>
                    {environmentTypes.map((type) => (
                      <SelectItem key={type.key} value={type.key}>
                        {type.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="flex gap-2 mt-4">
              <Button onClick={applyFilters}>{t("common.apply")}</Button>
              <Button variant="outline" onClick={resetFilters}>
                {t("common.reset")}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{t("dev_environments.list")}</CardTitle>
        </CardHeader>
        <CardContent>
          {environments.length === 0 ? (
            <div className="text-center py-8">
              <Monitor className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">
                {t("dev_environments.empty.title")}
              </h3>
              <p className="text-muted-foreground">
                {t("dev_environments.empty.description")}
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("dev_environments.table.name")}</TableHead>
                    <TableHead>{t("dev_environments.table.type")}</TableHead>
                    <TableHead>
                      {t("dev_environments.table.resources")}
                    </TableHead>
                    <TableHead className="text-right">
                      {t("common.actions")}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {environments.map((environment) => (
                    <TableRow key={environment.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{environment.name}</div>
                          {environment.description && (
                            <div className="text-sm text-muted-foreground">
                              {environment.description}
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant="outline"
                          className={
                            typeConfigMap[environment.type]?.color ||
                            "text-gray-600"
                          }
                        >
                          {typeConfigMap[environment.type]?.label ||
                            environment.type}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="text-sm">
                          <div>
                            CPU: {environment.cpu_limit}{" "}
                            {t("dev_environments.stats.cores")}
                          </div>
                          <div>
                            {t("dev_environments.stats.memory")}:{" "}
                            {formatMemory(environment.memory_limit)}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="h-8 w-8 p-0">
                              <span className="sr-only">
                                {t("common.open_menu")}
                              </span>
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuLabel>
                              {t("common.actions")}
                            </DropdownMenuLabel>

                            <DropdownMenuItem
                              onClick={() => onEdit(environment)}
                            >
                              <Edit className="h-4 w-4 mr-2" />
                              {t("common.edit")}
                            </DropdownMenuItem>

                            <DropdownMenuItem
                              onClick={() => onDelete(environment.id)}
                              className="text-destructive"
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              {t("common.delete")}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {totalPages > 1 && (
                <div className="flex items-center justify-between">
                  <div className="text-sm text-muted-foreground">
                    {t("common.page")} {params.page || 1} / {totalPages}
                  </div>
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onPageChange((params.page || 1) - 1)}
                      disabled={!params.page || params.page <= 1}
                    >
                      <ChevronLeft className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onPageChange((params.page || 1) + 1)}
                      disabled={!params.page || params.page >= totalPages}
                    >
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default DevEnvironmentList;
