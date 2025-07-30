import { useState, useEffect, forwardRef, useImperativeHandle } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
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
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import {
  MoreHorizontal,
  Edit,
  Trash2,
  FolderOpen,
  Filter,
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  Settings,
} from "lucide-react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import { ROUTES } from "@/lib/constants";
import type { Project, ProjectListParams } from "@/types/project";

interface ProjectListProps {
  onEdit?: (project: Project) => void;
  onDelete?: (id: number) => void;
  onCreateNew?: () => void;
}

export interface ProjectListRef {
  refreshData: () => void;
}

export const ProjectList = forwardRef<ProjectListRef, ProjectListProps>(
  ({ onEdit, onDelete, onCreateNew }, ref) => {
    const { t } = useTranslation();
    const navigate = useNavigate();
    const [projects, setProjects] = useState<Project[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [totalPages, setTotalPages] = useState(0);
    const [currentPage, setCurrentPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [showFilters, setShowFilters] = useState(false);
    const [localFilters, setLocalFilters] = useState<ProjectListParams>({
      page: 1,
      page_size: 20,
    });

    const pageSize = 20;

    const loadProjects = async (params: ProjectListParams) => {
      try {
        setLoading(true);
        setError(null);

        const response = await apiService.projects.list(params);
        setProjects(response.projects);
        setTotalPages(response.total_pages);
        setTotal(response.total);
        setCurrentPage(params.page || 1);
      } catch (error) {
        const errorMessage =
          error instanceof Error
            ? error.message
            : t("projects.messages.loadFailed");
        setError(errorMessage);
        logError(error as Error, "Failed to load projects");
      } finally {
        setLoading(false);
      }
    };

    useImperativeHandle(ref, () => ({
      refreshData: () => {
        loadProjects(localFilters);
      },
    }));

    useEffect(() => {
      loadProjects(localFilters);
    }, []);

    const handlePageChange = (page: number) => {
      const newFilters = { ...localFilters, page };
      setLocalFilters(newFilters);
      loadProjects(newFilters);
    };

    const handleFilterChange = (
      key: keyof ProjectListParams,
      value: string | number | undefined
    ) => {
      setLocalFilters((prev) => ({
        ...prev,
        [key]: value === "" ? undefined : value,
      }));
    };

    const applyFilters = () => {
      const filtersWithPage = { ...localFilters, page: 1 };
      setLocalFilters(filtersWithPage);
      loadProjects(filtersWithPage);
    };

    const resetFilters = () => {
      const emptyFilters: ProjectListParams = { page: 1, page_size: pageSize };
      setLocalFilters(emptyFilters);
      loadProjects(emptyFilters);
    };

    const handleRefresh = () => {
      loadProjects(localFilters);
    };

    const handleTasksManagement = (projectId: number) => {
      navigate(`${ROUTES.projects}/${projectId}/tasks`);
    };

    if (loading && projects.length === 0) {
      return (
        <div className="flex items-center justify-center h-64">
          <div className="text-gray-500">{t("common.loading")}</div>
        </div>
      );
    }

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
              onClick={handleRefresh}
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
                {t("projects.filter.title")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 gap-4">
                <div className="flex flex-col gap-3">
                  <Label htmlFor="name">{t("projects.name")}</Label>
                  <Input
                    id="name"
                    value={localFilters.name || ""}
                    onChange={(e) => handleFilterChange("name", e.target.value)}
                    placeholder={t("projects.filter.name_placeholder")}
                  />
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

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <p className="text-red-700">{error}</p>
          </div>
        )}

        <Card>
          <CardHeader>
            <CardTitle>{t("projects.list")}</CardTitle>
          </CardHeader>
          <CardContent>
            {projects.length === 0 ? (
              <div className="text-center py-8">
                <FolderOpen className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-semibold mb-2">
                  {localFilters.name
                    ? t("projects.messages.noMatchingProjects")
                    : t("projects.messages.noProjects")}
                </h3>
                <p className="text-muted-foreground mb-4">
                  {localFilters.name
                    ? t("projects.messages.clearFilter")
                    : t("projects.messages.noProjectsDesc")}
                </p>
                {localFilters.name ? (
                  <Button variant="outline" onClick={resetFilters}>
                    {t("projects.messages.clearFilter")}
                  </Button>
                ) : (
                  onCreateNew && (
                    <Button onClick={onCreateNew}>
                      {t("projects.create")}
                    </Button>
                  )
                )}
              </div>
            ) : (
              <div className="space-y-4">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("projects.name")}</TableHead>
                      <TableHead>{t("projects.repoUrl")}</TableHead>
                      <TableHead>{t("projects.credential")}</TableHead>
                      <TableHead className="text-center">{t("projects.taskCount")}</TableHead>
                      <TableHead className="text-right">
                        {t("common.actions")}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {projects.map((project) => (
                      <TableRow key={project.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{project.name}</div>
                            {project.description && (
                              <div className="text-sm text-muted-foreground">
                                {project.description}
                              </div>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-blue-600 truncate max-w-xs">
                            {project.repo_url}
                          </div>
                        </TableCell>
                        <TableCell>
                          {project.credential ? (
                            <div className="text-sm">
                              {project.credential.name}
                            </div>
                          ) : (
                            <span className="text-muted-foreground">-</span>
                          )}
                        </TableCell>
                        <TableCell className="text-center">
                          <span className="inline-flex items-center justify-center min-w-[2rem] h-6 px-2 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
                            {project.task_count ?? 0}
                          </span>
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
                                onClick={() =>
                                  handleTasksManagement(project.id)
                                }
                              >
                                <Settings className="h-4 w-4 mr-2" />
                                {t("projects.tasksManagement")}
                              </DropdownMenuItem>

                              {onEdit && (
                                <DropdownMenuItem
                                  onClick={() => onEdit(project)}
                                >
                                  <Edit className="h-4 w-4 mr-2" />
                                  {t("common.edit")}
                                </DropdownMenuItem>
                              )}

                              {onDelete && (
                                <DropdownMenuItem
                                  onClick={() => onDelete(project.id)}
                                  className="text-destructive"
                                >
                                  <Trash2 className="h-4 w-4 mr-2" />
                                  {t("common.delete")}
                                </DropdownMenuItem>
                              )}
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
                      {t("common.page")} {currentPage} / {totalPages}
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePageChange(currentPage - 1)}
                        disabled={currentPage <= 1}
                      >
                        <ChevronLeft className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePageChange(currentPage + 1)}
                        disabled={currentPage >= totalPages}
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
  }
);
