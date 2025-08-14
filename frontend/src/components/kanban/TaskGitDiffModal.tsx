import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  File,
  FileText,
  Plus,
  Minus,
  GitBranch,
  GitCommit,
  ChevronDown,
  ChevronRight,
  Loader2,
  AlertCircle,
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
  const [expandedFiles, setExpandedFiles] = useState<Set<string>>(new Set());
  const [fileContents, setFileContents] = useState<Map<string, string>>(
    new Map()
  );
  const [loadingFiles, setLoadingFiles] = useState<Set<string>>(new Set());

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

  const toggleFileExpanded = (filePath: string) => {
    const newExpanded = new Set(expandedFiles);
    if (expandedFiles.has(filePath)) {
      newExpanded.delete(filePath);
    } else {
      newExpanded.add(filePath);
      loadFileContent(filePath);
    }
    setExpandedFiles(newExpanded);
  };

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen && task) {
      setDiffSummary(null);
      setError(null);
      setExpandedFiles(new Set());
      setFileContents(new Map());
      setLoadingFiles(new Set());
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
            <Tabs defaultValue="summary" className="w-full h-full flex flex-col">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="summary">{t("gitDiff.summary")}</TabsTrigger>
                <TabsTrigger value="files">{t("gitDiff.fileChanges")}</TabsTrigger>
              </TabsList>

              <TabsContent value="summary" className="flex-1 space-y-4">
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
              </TabsContent>

              <TabsContent value="files" className="flex-1 overflow-auto space-y-2">
                {safeFiles.length > 0 ? (
                  safeFiles.map((file) => (
                    <Card key={file.path} className="border">
                      <CardContent className="p-0">
                        <div
                          className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
                          onClick={() => toggleFileExpanded(file.path)}
                        >
                          <div className="flex items-center gap-3">
                            {expandedFiles.has(file.path) ? (
                              <ChevronDown className="w-4 h-4" />
                            ) : (
                              <ChevronRight className="w-4 h-4" />
                            )}
                            <File className="w-4 h-4 text-gray-500" />
                            <span className="font-medium">{file.path}</span>
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
                          </div>
                          <div className="flex items-center gap-2 text-sm">
                            {!file.is_binary && (
                              <>
                                <span className="text-green-600">
                                  +{file.additions}
                                </span>
                                <span className="text-red-600">
                                  -{file.deletions}
                                </span>
                              </>
                            )}
                          </div>
                        </div>

                        {expandedFiles.has(file.path) && (
                          <div className="border-t p-4">
                            {loadingFiles.has(file.path) ? (
                              <div className="flex items-center justify-center py-8">
                                <Loader2 className="w-6 h-6 animate-spin" />
                                <span className="ml-2">
                                  {t("common.loading")}
                                </span>
                              </div>
                            ) : file.is_binary ? (
                              <div className="text-center py-8 text-muted-foreground">
                                {t("gitDiff.binaryFileNote")}
                              </div>
                            ) : (
                              renderDiffContent(fileContents.get(file.path) || "")
                            )}
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  ))
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    {t("gitDiff.noChanges")}
                  </div>
                )}
              </TabsContent>
            </Tabs>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
