import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
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
import { Label } from "@/components/ui/label";
import { useTranslation } from "react-i18next";
import {
  Edit,
  Trash2,
  Play,
  CheckCircle,
  Clock,
  GitBranch,
  GitPullRequest,
  ChevronLeft,
  ChevronRight,
  Plus,
  RefreshCw,
  Filter,
  X,
  MessageSquare,
  MoreHorizontal,
} from "lucide-react";
import type { Task, TaskStatus } from "@/types/task";
import type { Project } from "@/types/project";

interface TaskListProps {
  tasks: Task[];
  projects: Project[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  statusFilter?: TaskStatus;
  projectFilter?: number;
  hideProjectFilter?: boolean;
  onPageChange: (page: number) => void;
  onStatusFilterChange: (status: TaskStatus | undefined) => void;
  onProjectFilterChange: (projectId: number | undefined) => void;
  onEdit: (task: Task) => void;
  onDelete: (id: number) => void;
  onViewConversation?: (task: Task) => void;
  onCreateNew: () => void;
}

export function TaskList({
  tasks,
  projects,
  loading,
  currentPage,
  totalPages,
  total,
  statusFilter,
  projectFilter,
  hideProjectFilter = false,
  onPageChange,
  onStatusFilterChange,
  onProjectFilterChange,
  onEdit,
  onDelete,
  onViewConversation,
  onCreateNew,
}: TaskListProps) {
  const { t } = useTranslation();
  const [showFilters, setShowFilters] = useState(false);

  const getStatusColor = (status: TaskStatus) => {
    switch (status) {
      case "todo":
        return "bg-gray-100 text-gray-800 border-gray-300";
      case "in_progress":
        return "bg-blue-100 text-blue-800 border-blue-300";
      case "done":
        return "bg-green-100 text-green-800 border-green-300";
      case "cancelled":
        return "bg-red-100 text-red-800 border-red-300";
      default:
        return "bg-gray-100 text-gray-800 border-gray-300";
    }
  };

  const getStatusIcon = (status: TaskStatus) => {
    switch (status) {
      case "todo":
        return <Clock className="w-3 h-3" />;
      case "in_progress":
        return <Play className="w-3 h-3" />;
      case "done":
        return <CheckCircle className="w-3 h-3" />;
      case "cancelled":
        return <X className="w-3 h-3" />;
      default:
        return <Clock className="w-3 h-3" />;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const handleDeleteClick = (task: Task) => {
    if (confirm(t("tasks.messages.deleteConfirm", { title: task.title }))) {
      onDelete(task.id);
    }
  };

  const handleRefresh = () => {
    // 通过重新应用当前过滤器来刷新数据
    onStatusFilterChange(statusFilter);
    onProjectFilterChange(projectFilter);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-center">
          <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-2" />
          <p>{t("common.loading")}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 顶部工具栏 */}
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

      {/* 过滤器 */}
      {showFilters && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              {t("tasks.filters.title")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* 项目筛选 */}
              {!hideProjectFilter && (
                <div className="flex flex-col gap-3">
                  <Label htmlFor="project">{t("tasks.filters.project")}</Label>
                  <Select
                    value={projectFilter?.toString() || "all"}
                    onValueChange={(value) =>
                      onProjectFilterChange(
                        value === "all" ? undefined : parseInt(value)
                      )
                    }
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder={t("common.all")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">{t("common.all")}</SelectItem>
                      {projects.map((project) => (
                        <SelectItem
                          key={project.id}
                          value={project.id.toString()}
                        >
                          {project.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}

              {/* 状态筛选 */}
              <div className="flex flex-col gap-3">
                <Label htmlFor="status">{t("tasks.filters.status")}</Label>
                <Select
                  value={statusFilter || "all"}
                  onValueChange={(value) =>
                    onStatusFilterChange(
                      value === "all" ? undefined : (value as TaskStatus)
                    )
                  }
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder={t("common.all")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("common.all")}</SelectItem>
                    <SelectItem value="todo">
                      {t("tasks.status.todo")}
                    </SelectItem>
                    <SelectItem value="in_progress">
                      {t("tasks.status.in_progress")}
                    </SelectItem>
                    <SelectItem value="done">
                      {t("tasks.status.done")}
                    </SelectItem>
                    <SelectItem value="cancelled">
                      {t("tasks.status.cancelled")}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 任务表格 */}
      <Card>
        <CardHeader>
          <CardTitle>{t("tasks.list")}</CardTitle>
        </CardHeader>
        <CardContent>
          {tasks.length === 0 ? (
            <div className="text-center py-8">
              <Clock className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">
                {t("tasks.empty.title")}
              </h3>
              <p className="text-muted-foreground mb-4">
                {t("tasks.empty.description")}
              </p>
              <Button onClick={onCreateNew}>
                <Plus className="w-4 h-4 mr-2" />
                {t("tasks.actions.create")}
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("tasks.table.title")}</TableHead>
                    {!hideProjectFilter && (
                      <TableHead>{t("tasks.table.project")}</TableHead>
                    )}
                    <TableHead>{t("tasks.table.status")}</TableHead>
                    <TableHead>{t("tasks.table.branch")}</TableHead>
                    <TableHead>{t("tasks.table.environment")}</TableHead>
                    <TableHead>{t("tasks.table.updated")}</TableHead>
                    <TableHead className="text-right">
                      {t("common.actions")}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tasks.map((task) => (
                    <TableRow key={task.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{task.title}</div>
                          <div className="text-xs text-muted-foreground">
                            {t("common.createdAt")}:{" "}
                            {formatDate(task.created_at)}
                          </div>
                        </div>
                      </TableCell>
                      {!hideProjectFilter && (
                        <TableCell>
                          {task.project?.name && (
                            <span className="text-blue-600">
                              {task.project.name}
                            </span>
                          )}
                        </TableCell>
                      )}
                      <TableCell>
                        <Badge
                          variant="outline"
                          className={`text-xs ${getStatusColor(task.status)}`}
                        >
                          {getStatusIcon(task.status)}
                          <span className="ml-1">
                            {t(`tasks.status.${task.status}`)}
                          </span>
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center space-x-2">
                          <GitBranch className="w-4 h-4 text-muted-foreground" />
                          <span className="text-sm">{task.start_branch}</span>
                          {task.has_pull_request && (
                            <Badge variant="outline" className="text-xs">
                              <GitPullRequest className="w-3 h-3 mr-1" />
                              PR
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        {task.dev_environment ? (
                          <div className="flex items-center space-x-2">
                            <div className="w-2 h-2 rounded-full bg-blue-500"></div>
                            <span className="text-sm">
                              {task.dev_environment.name}
                            </span>
                            <Badge
                              variant="outline"
                              className={`text-xs ${
                                task.dev_environment.status === "running"
                                  ? "bg-green-100 text-green-800 border-green-300"
                                  : task.dev_environment.status === "stopped"
                                  ? "bg-gray-100 text-gray-800 border-gray-300"
                                  : task.dev_environment.status === "error"
                                  ? "bg-red-100 text-red-800 border-red-300"
                                  : "bg-yellow-100 text-yellow-800 border-yellow-300"
                              }`}
                            >
                              {task.dev_environment.status}
                            </Badge>
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">
                            -
                          </span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="text-xs text-muted-foreground">
                          {formatDate(task.updated_at)}
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

                            {onViewConversation && (
                              <DropdownMenuItem
                                onClick={() => onViewConversation(task)}
                              >
                                <MessageSquare className="h-4 w-4 mr-2" />
                                {t("tasks.actions.viewConversation")}
                              </DropdownMenuItem>
                            )}

                            <DropdownMenuItem onClick={() => onEdit(task)}>
                              <Edit className="h-4 w-4 mr-2" />
                              {t("common.edit")}
                            </DropdownMenuItem>

                            <DropdownMenuItem
                              onClick={() => handleDeleteClick(task)}
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

              {/* 分页 */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between">
                  <div className="text-sm text-muted-foreground">
                    {t("common.page")} {currentPage} / {totalPages}
                  </div>
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onPageChange(currentPage - 1)}
                      disabled={currentPage <= 1}
                    >
                      <ChevronLeft className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onPageChange(currentPage + 1)}
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
