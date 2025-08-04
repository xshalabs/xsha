import React, { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  ArrowLeft,
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
import { usePageTitle } from "@/hooks/usePageTitle";
import { apiService } from "@/lib/api/index";
import { logError } from "@/lib/errors";
import type { TaskConversation, GitDiffSummary } from "@/types/task-conversation";

const getDefaultDiffSummary = (): GitDiffSummary => ({
  total_files: 0,
  total_additions: 0,
  total_deletions: 0,
  files: [],
  commits_behind: 0,
  commits_ahead: 0,
});

const TaskConversationGitDiffPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, taskId, conversationId } = useParams<{
    projectId: string;
    taskId: string;
    conversationId: string;
  }>();

  const getStatusText = (status: string) => {
    const statusMap = {
      added: t("gitDiff.status.added"),
      modified: t("gitDiff.status.modified"),
      deleted: t("gitDiff.status.deleted"),
      renamed: t("gitDiff.status.renamed"),
    } as const;
    return statusMap[status as keyof typeof statusMap] || status;
  };

  const [conversation, setConversation] = useState<TaskConversation | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [diffSummary, setDiffSummary] = useState<GitDiffSummary | null>(null);
  const [diffLoading, setDiffLoading] = useState(false);
  const [diffError, setDiffError] = useState<string | null>(null);
  const [expandedFiles, setExpandedFiles] = useState<Set<string>>(new Set());
  const [fileContents, setFileContents] = useState<Map<string, string>>(new Map());
  const [loadingFiles, setLoadingFiles] = useState<Set<string>>(new Set());

  usePageTitle(
    conversation
      ? `${t("gitDiff.title")} - ${conversation.task?.title}`
      : t("gitDiff.title")
  );

  useEffect(() => {
    const loadConversation = async () => {
      if (!conversationId) {
        logError(new Error("Conversation ID is required"), "Invalid conversation ID");
        navigate(`/projects/${projectId}/tasks/${taskId}/conversation`);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        const response = await apiService.taskConversations.get(parseInt(conversationId, 10));
        setConversation(response.data);
      } catch (error) {
        logError(error as Error, "Failed to load conversation");
        setError(error instanceof Error ? error.message : "Failed to load conversation");
      } finally {
        setLoading(false);
      }
    };

    loadConversation();
  }, [conversationId, projectId, taskId, navigate]);

  useEffect(() => {
    const loadDiffSummary = async () => {
      if (!conversation || !conversationId || !conversation.commit_hash) return;

      try {
        setDiffLoading(true);
        setDiffError(null);
        const response = await apiService.taskConversations.getGitDiff(
          parseInt(conversationId, 10),
          { include_content: false }
        );
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
        logError(error as Error, "Failed to load conversation git diff");
        setDiffError(
          error instanceof Error ? error.message : "Failed to load conversation git diff"
        );
      } finally {
        setDiffLoading(false);
      }
    };

    loadDiffSummary();
  }, [conversation, conversationId]);

  const loadFileContent = async (filePath: string) => {
    if (!conversationId || fileContents.has(filePath) || loadingFiles.has(filePath)) {
      return;
    }

    try {
      setLoadingFiles((prev) => new Set(prev).add(filePath));
      const response = await apiService.taskConversations.getGitDiffFile(
        parseInt(conversationId, 10),
        { file_path: filePath }
      );
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
    if (newExpanded.has(filePath)) {
      newExpanded.delete(filePath);
    } else {
      newExpanded.add(filePath);
      loadFileContent(filePath);
    }
    setExpandedFiles(newExpanded);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "added":
        return "text-green-600 border-green-600";
      case "modified":
        return "text-blue-600 border-blue-600";
      case "deleted":
        return "text-red-600 border-red-600";
      case "renamed":
        return "text-yellow-600 border-yellow-600";
      default:
        return "text-gray-600 border-gray-600";
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
      <div className="bg-foreground/5 border border-border rounded p-3 text-xs font-mono overflow-x-auto max-w-full">
        {lines &&
          lines.map((line, index) => {
            let className = "";
            if (line.startsWith("+")) {
              className = "text-green-600 bg-green-50";
            } else if (line.startsWith("-")) {
              className = "text-red-600 bg-red-50";
            } else if (line.startsWith("@@")) {
              className = "text-blue-600 font-bold";
            }

            return (
              <div key={index} className={`${className} px-2 py-0.5`}>
                {line || " "}
              </div>
            );
          })}
      </div>
    );
  };

  const handleGoBack = () => {
    navigate(`/projects/${projectId}/tasks/${taskId}/conversation`);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">{t("common.loading")}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error && !loading) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <p className="text-red-600 mb-4">{error}</p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  className="text-foreground hover:text-foreground"
                  onClick={handleGoBack}
                >
                  <ArrowLeft className="h-4 w-4 mr-2" />
                  {t("common.back")}
                </Button>
                <Button
                  variant="default"
                  onClick={() => window.location.reload()}
                >
                  {t("common.retry")}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!conversation && !loading) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <p className="text-muted-foreground">
                {t("taskConversation.messages.loadFailed")}
              </p>
              <Button
                variant="outline"
                onClick={handleGoBack}
                className="mt-4 text-foreground hover:text-foreground"
              >
                <ArrowLeft className="h-4 w-4 mr-2" />
                {t("common.back")}
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!conversation || !conversation.commit_hash) {
    return (
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center text-muted-foreground">
              <p className="text-lg font-medium mb-2">
                {t("taskConversation.gitDiff.noCommitHash.title")}
              </p>
              <p className="text-sm mb-4">
                {t("taskConversation.gitDiff.noCommitHash.description")}
              </p>
              <Button variant="outline" onClick={handleGoBack}>
                {t("common.back")}
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  const safeFiles =
    diffSummary && Array.isArray(diffSummary.files) ? diffSummary.files : [];

  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-6">
          <div className="min-w-0 flex-1">
            <h1 className="text-3xl font-bold text-foreground">
              {t("gitDiff.title")}
            </h1>
            <p className="mt-2 text-sm text-muted-foreground">
              {conversation.task?.title} - {t("tasks.conversation")} #{conversation.id}
            </p>
            <div className="mt-2 flex items-center text-sm text-muted-foreground">
              <GitCommit className="w-4 h-4 mr-1 flex-shrink-0" />
              <span className="font-mono text-xs sm:text-sm truncate">{conversation.commit_hash}</span>
            </div>
          </div>
          <Button
            variant="outline"
            className="text-foreground hover:text-foreground ml-4 flex-shrink-0"
            onClick={handleGoBack}
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            <span className="hidden sm:inline">{t("common.back")}</span>
          </Button>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-6 space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <GitBranch className="w-5 h-5" />
              {t("gitDiff.summary")}
            </CardTitle>
            <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1">
                <GitCommit className="w-4 h-4 flex-shrink-0" />
                <span className="font-mono text-xs sm:text-sm truncate">{conversation.commit_hash}</span>
              </div>
              {diffSummary && (
                <div className="flex flex-wrap gap-2 sm:gap-4">
                  <span className="text-green-600">
                    +{diffSummary.total_additions || 0}
                  </span>
                  <span className="text-red-600">
                    -{diffSummary.total_deletions || 0}
                  </span>
                  <span className="whitespace-nowrap">
                    {diffSummary.total_files || 0}{" "}
                    {t("gitDiff.filesChanged")}
                  </span>
                </div>
              )}
            </div>
          </CardHeader>

          <CardContent>
            {diffLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="w-6 h-6 animate-spin mr-2" />
                <span>{t("common.loading")}</span>
              </div>
            ) : diffError ? (
              <div className="flex items-center justify-center py-8 text-red-600">
                <AlertCircle className="w-5 h-5 mr-2" />
                <span>{diffError}</span>
              </div>
            ) : diffSummary ? (
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
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
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                {t("gitDiff.noData")}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <File className="w-5 h-5" />
              {t("gitDiff.fileChanges")}
            </CardTitle>
          </CardHeader>

          <CardContent>
            {diffLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="w-6 h-6 animate-spin mr-2" />
                <span>{t("common.loading")}</span>
              </div>
            ) : diffError ? (
              <div className="flex items-center justify-center py-8 text-red-600">
                <AlertCircle className="w-5 h-5 mr-2" />
                <span>{diffError}</span>
              </div>
            ) : safeFiles.length > 0 ? (
              <div className="space-y-2">
                {safeFiles.map((file) => (
                  <div key={file.path} className="border border-border rounded-lg overflow-hidden">
                    <div
                      className="flex items-center justify-between p-4 cursor-pointer hover:bg-muted/50"
                      onClick={() => toggleFileExpanded(file.path)}
                    >
                      <div className="flex items-center gap-3 min-w-0 flex-1">
                        {expandedFiles.has(file.path) ? (
                          <ChevronDown className="w-4 h-4 flex-shrink-0" />
                        ) : (
                          <ChevronRight className="w-4 h-4 flex-shrink-0" />
                        )}
                        <File className="w-4 h-4 text-gray-500 flex-shrink-0" />
                        <span
                          className="font-medium text-sm truncate"
                          title={file.path}
                        >
                          {file.path}
                        </span>
                      </div>
                      <div className="flex items-center gap-2 flex-shrink-0 ml-2">
                        <div className="hidden sm:flex items-center gap-2">
                          <Badge
                            variant="outline"
                            className={`text-xs ${getStatusColor(
                              file.status
                            )}`}
                          >
                            {getStatusIcon(file.status)}
                            <span className="ml-1 hidden md:inline">
                              {getStatusText(file.status)}
                            </span>
                          </Badge>
                          {file.is_binary && (
                            <Badge variant="outline" className="text-xs">
                              {t("gitDiff.binary")}
                            </Badge>
                          )}
                        </div>
                        {!file.is_binary && (
                          <div className="flex items-center gap-1 text-xs">
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
                    
                    {/* 在小屏幕上显示状态标签 */}
                    <div className="sm:hidden px-4 pb-3">
                      <div className="flex items-center gap-2">
                        <Badge
                          variant="outline"
                          className={`text-xs ${getStatusColor(
                            file.status
                          )}`}
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
                    </div>

                    {expandedFiles.has(file.path) && (
                      <div className="border-t border-border p-4">
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
                          renderDiffContent(
                            fileContents.get(file.path) || ""
                          )
                        )}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                {t("gitDiff.noChanges")}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default TaskConversationGitDiffPage; 