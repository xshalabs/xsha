import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  File,
  FileText,
  Plus,
  Minus,
  GitBranch,
  GitCommit,

  Loader2,
  AlertCircle,
  FolderOpen,
} from "lucide-react";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { GitDiffSummary } from "@/lib/api/tasks";
import type { Task } from "@/types/task";

interface TaskGitDiffModalProps {
  isOpen: boolean;
  onClose: () => void;
  task: Task | null;
}

const getDefaultDiffSummary = (): GitDiffSummary => ({
  total_files: 0,
  total_additions: 0,
  total_deletions: 0,
  files: [],
  commits_behind: 0,
  commits_ahead: 0,
});

export const TaskGitDiffModal: React.FC<TaskGitDiffModalProps> = ({
  isOpen,
  onClose,
  task,
}) => {
  const { t } = useTranslation();

  const getStatusText = (status: string) => {
    const statusMap = {
      added: t("gitDiff.status.added"),
      modified: t("gitDiff.status.modified"),
      deleted: t("gitDiff.status.deleted"),
      renamed: t("gitDiff.status.renamed"),
    } as const;
    return statusMap[status as keyof typeof statusMap] || status;
  };

  const [diffSummary, setDiffSummary] = useState<GitDiffSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [fileContents, setFileContents] = useState<Map<string, string>>(
    new Map()
  );
  const [loadingFiles, setLoadingFiles] = useState<Set<string>>(new Set());
  const [selectedFile, setSelectedFile] = useState<string | null>(null);

  const loadDiffSummary = async () => {
    if (!task) return;

    try {
      setLoading(true);
      setError(null);
      const response = await apiService.tasks.getTaskGitDiff(task.id, {
        include_content: false,
      });
      const data = response.data || getDefaultDiffSummary();
      const safeDiffSummary: GitDiffSummary = {
        total_files: data.total_files || 0,
        total_additions: data.total_additions || 0,
        total_deletions: data.total_deletions || 0,
        files: Array.isArray(data.files) ? data.files : [],
        commits_behind: data.commits_behind || 0,
        commits_ahead: data.commits_ahead || 0,
      };
      setDiffSummary(safeDiffSummary);
    } catch (error) {
      logError(error as Error, "Failed to load git diff");
      setError(
        error instanceof Error ? error.message : "Failed to load git diff"
      );
    } finally {
      setLoading(false);
    }
  };

  const loadFileContent = async (filePath: string) => {
    if (!task || fileContents.has(filePath) || loadingFiles.has(filePath)) {
      return;
    }

    try {
      setLoadingFiles((prev) => new Set(prev).add(filePath));
      const response = await apiService.tasks.getTaskGitDiffFile(task.id, {
        file_path: filePath,
      });
      setFileContents((prev) =>
        new Map(prev).set(filePath, response.data.diff_content)
      );
    } catch (error) {
      logError(error as Error, `Failed to load file content for ${filePath}`);
    } finally {
      setLoadingFiles((prev) => {
        const newSet = new Set(prev);
        newSet.delete(filePath);
        return newSet;
      });
    }
  };



  const handleFileSelect = (filePath: string) => {
    setSelectedFile(filePath);
    if (!fileContents.has(filePath)) {
      loadFileContent(filePath);
    }
  };

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen && task) {
      setDiffSummary(null);
      setError(null);

      setFileContents(new Map());
      setLoadingFiles(new Set());
      setSelectedFile(null);
      loadDiffSummary();
    }
  }, [isOpen, task]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case "added":
        return "bg-green-100 text-green-800 border-green-300";
      case "modified":
        return "bg-blue-100 text-blue-800 border-blue-300";
      case "deleted":
        return "bg-red-100 text-red-800 border-red-300";
      case "renamed":
        return "bg-yellow-100 text-yellow-800 border-yellow-300";
      default:
        return "bg-gray-100 text-gray-800 border-gray-300";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "added":
        return <Plus className="w-3 h-3" />;
      case "modified":
        return <FileText className="w-3 h-3" />;
      case "deleted":
        return <Minus className="w-3 h-3" />;
      case "renamed":
        return <File className="w-3 h-3" />;
      default:
        return <File className="w-3 h-3" />;
    }
  };

  const renderDiffContent = (content: string) => {
    if (!content) return null;

    const lines = content.split("\n");
    return (
      <div className="bg-gray-50 rounded border font-mono text-sm overflow-x-auto max-h-[60vh]">
        {lines &&
          lines.map((line, index) => {
            let lineClass = "px-4 py-1 whitespace-pre";
            if (line.startsWith("+")) {
              lineClass += " bg-green-50 text-green-800";
            } else if (line.startsWith("-")) {
              lineClass += " bg-red-50 text-red-800";
            } else if (line.startsWith("@@")) {
              lineClass += " bg-blue-50 text-blue-800 font-semibold";
            }

            return (
              <div key={index} className={lineClass}>
                {line || " "}
              </div>
            );
          })}
      </div>
    );
  };

  if (!task) return null;

  const safeFiles = Array.isArray(diffSummary?.files) ? diffSummary.files : [];

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="!max-w-[95vw] !w-[95vw] max-h-[95vh] h-[95vh] flex flex-col sm:!max-w-[95vw]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <GitBranch className="w-5 h-5" />
            {t("gitDiff.title")} - {task.title}
          </DialogTitle>
          <DialogDescription>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1">
                <GitCommit className="w-4 h-4" />
                <span>
                  {task.start_branch} → {task.work_branch}
                </span>
              </div>
              {diffSummary && (
                <div className="flex gap-4">
                  <span className="text-green-600">
                    +{diffSummary.total_additions || 0}
                  </span>
                  <span className="text-red-600">
                    -{diffSummary.total_deletions || 0}
                  </span>
                  <span>
                    {diffSummary.total_files || 0} {t("gitDiff.filesChanged")}
                  </span>
                </div>
              )}
            </div>
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-hidden">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-center">
                <Loader2 className="w-8 h-8 animate-spin mx-auto mb-2" />
                <p className="text-muted-foreground">{t("common.loading")}</p>
              </div>
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-center">
                <AlertCircle className="w-8 h-8 text-red-500 mx-auto mb-2" />
                <p className="text-red-600 mb-4">{error}</p>
                <Button variant="outline" onClick={loadDiffSummary}>
                  {t("common.retry")}
                </Button>
              </div>
            </div>
          ) : !diffSummary ? (
            <div className="flex items-center justify-center py-8">
              <p className="text-muted-foreground">{t("gitDiff.noData")}</p>
            </div>
          ) : (
            <div className="space-y-6 h-full flex flex-col">
              {/* Summary section - 优先显示在最上方 */}
              <Card className="flex-shrink-0">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <GitBranch className="w-5 h-5" />
                    {t("gitDiff.summary")}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <Card>
                      <CardContent className="p-4 text-center">
                        <div className="text-2xl font-bold text-blue-600">
                          {diffSummary.total_files || 0}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {t("gitDiff.filesChanged")}
                        </div>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardContent className="p-4 text-center">
                        <div className="text-2xl font-bold text-green-600">
                          +{diffSummary.total_additions || 0}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {t("gitDiff.additions")}
                        </div>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardContent className="p-4 text-center">
                        <div className="text-2xl font-bold text-red-600">
                          -{diffSummary.total_deletions || 0}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {t("gitDiff.deletions")}
                        </div>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardContent className="p-4 text-center">
                        <div className="text-2xl font-bold text-gray-600">
                          {diffSummary.commits_ahead || 0}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {t("gitDiff.commitsAhead")}
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                </CardContent>
              </Card>

              {/* File Changes section - 左侧文件树 + 右侧内容 */}
              <Card className="flex-1 overflow-hidden">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <FolderOpen className="w-5 h-5" />
                    {t("gitDiff.fileChanges")}
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-0 h-full">
                  {safeFiles.length > 0 ? (
                    <div className="flex flex-col md:flex-row h-full">
                      {/* 左侧文件树 */}
                      <div className="w-full md:w-1/3 border-b md:border-b-0 md:border-r bg-muted/20 overflow-y-auto max-h-48 md:max-h-none">
                        <div className="p-2 md:p-4">
                          <div className="space-y-1">
                            {safeFiles.map((file) => (
                              <div
                                key={file.path}
                                className={`flex items-center gap-2 p-2 rounded cursor-pointer hover:bg-muted/50 text-sm transition-colors ${
                                  selectedFile === file.path 
                                    ? 'bg-muted border-l-2 md:border-l-2 border-l-primary' 
                                    : ''
                                }`}
                                onClick={() => handleFileSelect(file.path)}
                              >
                                <File className="w-4 h-4 text-gray-500 flex-shrink-0" />
                                <span className="font-medium truncate flex-1" title={file.path}>
                                  <span className="hidden sm:inline">{file.path}</span>
                                  <span className="sm:hidden">{file.path.split('/').pop() || file.path}</span>
                                </span>
                                <div className="flex items-center gap-1 text-xs flex-shrink-0">
                                  <Badge
                                    variant="outline"
                                    className={`text-xs px-1 py-0 ${getStatusColor(file.status)}`}
                                  >
                                    {getStatusIcon(file.status)}
                                  </Badge>
                                  {!file.is_binary && (
                                    <div className="hidden sm:flex items-center gap-1">
                                      <span className="text-green-600">
                                        +{file.additions}
                                      </span>
                                      <span className="text-red-600">
                                        -{file.deletions}
                                      </span>
                                    </div>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                      
                      {/* 右侧变化内容 */}
                      <div className="flex-1 overflow-y-auto">
                        {selectedFile ? (
                          <div className="h-full">
                            {/* 文件头部信息 */}
                            <div className="border-b p-2 md:p-4 bg-background sticky top-0 z-10">
                              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
                                <div className="flex items-center gap-2 md:gap-3 min-w-0 flex-1">
                                  <File className="w-4 h-4 md:w-5 md:h-5 text-gray-500 flex-shrink-0" />
                                  <span className="font-medium text-sm md:text-base truncate" title={selectedFile}>
                                    {selectedFile}
                                  </span>
                                </div>
                                <div className="flex items-center gap-2 flex-shrink-0">
                                  {(() => {
                                    const file = safeFiles.find(f => f.path === selectedFile);
                                    return file ? (
                                      <>
                                        <Badge
                                          variant="outline"
                                          className={`text-xs ${getStatusColor(file.status)}`}
                                        >
                                          {getStatusIcon(file.status)}
                                          <span className="ml-1">
                                            {getStatusText(file.status)}
                                          </span>
                                        </Badge>
                                        {file.is_binary && (
                                          <Badge variant="outline" className="text-xs">
                                            {t("gitDiff.binary")}
                                          </Badge>
                                        )}
                                        {!file.is_binary && (
                                          <div className="flex items-center gap-1 text-xs sm:hidden">
                                            <span className="text-green-600">
                                              +{file.additions}
                                            </span>
                                            <span className="text-red-600">
                                              -{file.deletions}
                                            </span>
                                          </div>
                                        )}
                                      </>
                                    ) : null;
                                  })()}
                                </div>
                              </div>
                            </div>
                            
                            {/* 文件内容 */}
                            <div className="p-2 md:p-4">
                              {loadingFiles.has(selectedFile) ? (
                                <div className="flex items-center justify-center py-12">
                                  <div className="text-center">
                                    <Loader2 className="w-8 h-8 animate-spin mx-auto mb-2" />
                                    <span className="text-muted-foreground">
                                      {t("common.loading")}
                                    </span>
                                  </div>
                                </div>
                              ) : (() => {
                                const file = safeFiles.find(f => f.path === selectedFile);
                                return file?.is_binary ? (
                                  <div className="text-center py-12 text-muted-foreground">
                                    <File className="w-12 h-12 mx-auto mb-2 opacity-50" />
                                    <p>{t("gitDiff.binaryFileNote")}</p>
                                  </div>
                                ) : (
                                  renderDiffContent(fileContents.get(selectedFile) || "")
                                );
                              })()}
                            </div>
                          </div>
                        ) : (
                          <div className="flex items-center justify-center h-full text-muted-foreground">
                            <div className="text-center">
                              <File className="w-12 h-12 mx-auto mb-2 opacity-50" />
                              <p>{t("gitDiff.selectFile")}</p>
                            </div>
                          </div>
                        )}
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-center justify-center h-32 text-muted-foreground">
                      {t("gitDiff.noChanges")}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
