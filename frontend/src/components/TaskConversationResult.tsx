import { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useTranslation } from "react-i18next";
import {
  CheckCircle,
  XCircle,
  Clock,
  DollarSign,
  MessageSquare,
  Activity,
  TrendingUp,
  Info,
  Code,
  Loader2,
  AlertCircle,
  Copy,
} from "lucide-react";
import { toast } from "sonner";
import type {
  TaskConversationResult,
  UsageStats,
  ParsedTaskConversationResult,
} from "@/types/task-conversation-result";
import { formatToLocal } from "@/lib/timezone";
import { taskConversationResultsApi } from "@/lib/api";

interface TaskConversationResultProps {
  conversationId: number;
  showHeader?: boolean;
  compact?: boolean;
}

export function TaskConversationResult({
  conversationId,
  showHeader = true,
  compact = false,
}: TaskConversationResultProps) {
  const { t } = useTranslation();
  const [result, setResult] = useState<ParsedTaskConversationResult | null>(
    null
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const parseUsage = (usageString: string): UsageStats | null => {
    try {
      return JSON.parse(usageString);
    } catch {
      return null;
    }
  };

  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}min`;
  };

  const formatCost = (cost: number): string => {
    return `$${cost.toFixed(4)}`;
  };

  const formatTokens = (tokens: number): string => {
    if (tokens >= 1000000) {
      return `${(tokens / 1000000).toFixed(1)}M`;
    }
    if (tokens >= 1000) {
      return `${(tokens / 1000).toFixed(1)}K`;
    }
    return tokens.toString();
  };

  const getStatusColor = (isError: boolean) => {
    return isError ? "destructive" : "default";
  };

  const getStatusIcon = (isError: boolean) => {
    return isError ? (
      <XCircle className="h-4 w-4 text-red-500" />
    ) : (
      <CheckCircle className="h-4 w-4 text-green-500" />
    );
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast.success(t("common.copied_to_clipboard"));
    } catch (err) {
      toast.error(t("common.copy_failed"));
    }
  };

  const fetchResult = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await taskConversationResultsApi.getByConversationId(
        conversationId
      );
      const resultData = response.data;

      const usage = parseUsage(resultData.usage);
      setResult({
        ...resultData,
        usage: usage || {
          input_tokens: 0,
          cache_creation_input_tokens: 0,
          cache_read_input_tokens: 0,
          output_tokens: 0,
        },
      });
    } catch (err: any) {
      setError(err.message || t("taskConversations.result_load_failed"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (conversationId) {
      fetchResult();
    }
  }, [conversationId]);

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center p-6">
          <Loader2 className="h-6 w-6 animate-spin mr-2" />
          <span>{t("common.loading")}</span>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center p-6">
          <AlertCircle className="h-6 w-6 text-muted-foreground mr-2" />
          <span className="text-muted-foreground">{error}</span>
          <Button
            variant="outline"
            size="sm"
            className="ml-2"
            onClick={fetchResult}
          >
            {t("common.retry")}
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (!result) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center p-6 text-gray-500">
          <Info className="h-6 w-6 mr-2" />
          <span>{t("taskConversations.no_result")}</span>
        </CardContent>
      </Card>
    );
  }

  if (compact) {
    return (
      <div className="flex items-center gap-2 text-sm">
        {getStatusIcon(result.is_error)}
        <Badge variant={getStatusColor(result.is_error)}>
          {result.subtype}
        </Badge>
        <span className="text-gray-600">
          {formatDuration(result.duration_ms)}
        </span>
        <span className="text-gray-600">
          {formatCost(result.total_cost_usd)}
        </span>
      </div>
    );
  }

  return (
    <Card>
      {showHeader && (
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {getStatusIcon(result.is_error)}
            {t("taskConversations.execution_result")}
            <Badge variant={getStatusColor(result.is_error)}>
              {result.subtype}
            </Badge>
          </CardTitle>
          <CardDescription>
            {t("taskConversations.result.description", {
              sessionId: result.session_id.slice(0, 8),
            })}
          </CardDescription>
        </CardHeader>
      )}

      <CardContent className="space-y-4">
        <div>
          <div className="flex items-center justify-between mb-2">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <MessageSquare className="h-4 w-4" />
              {t("taskConversations.result.content")}
            </h4>
            <Button
              variant="outline"
              size="sm"
              onClick={() => copyToClipboard(result.result)}
              className="h-7 px-2 text-foreground hover:text-foreground"
            >
              <Copy className="h-3 w-3 mr-1" />
              {t("common.copy")}
            </Button>
          </div>
          <div className="bg-foreground/5 p-3 rounded-md max-h-[400px] overflow-auto">
            <pre className="text-sm whitespace-pre-wrap">{result.result}</pre>
          </div>
        </div>

        <Separator />

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <Clock className="h-4 w-4" />
              {t("taskConversations.result.duration")}
            </div>
            <div className="text-lg font-semibold">
              {formatDuration(result.duration_ms)}
            </div>
          </div>

          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <Activity className="h-4 w-4" />
              {t("taskConversations.result.api_duration")}
            </div>
            <div className="text-lg font-semibold">
              {formatDuration(result.duration_api_ms)}
            </div>
          </div>

          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <TrendingUp className="h-4 w-4" />
              {t("taskConversations.result.turns")}
            </div>
            <div className="text-lg font-semibold">{result.num_turns}</div>
          </div>

          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <DollarSign className="h-4 w-4" />
              {t("taskConversations.result.cost")}
            </div>
            <div className="text-lg font-semibold">
              {formatCost(result.total_cost_usd)}
            </div>
          </div>
        </div>

        <Separator />

        <div>
          <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
            <Code className="h-4 w-4" />
            {t("taskConversations.result.usage_stats")}
          </h4>
          <div className="grid grid-cols-2 md:grid-cols-2 gap-3 text-sm">
            <div>
              <span className="text-gray-600">
                {t("taskConversations.result.input_tokens")}:
              </span>
              <span className="ml-2 font-medium">
                {formatTokens(result.usage.input_tokens)}
              </span>
              <span className="ml-1 text-xs text-gray-500">
                ({result.usage.input_tokens.toLocaleString()})
              </span>
            </div>
            <div>
              <span className="text-gray-600">
                {t("taskConversations.result.output_tokens")}:
              </span>
              <span className="ml-2 font-medium">
                {result.usage.output_tokens}
              </span>
            </div>
            <div>
              <span className="text-gray-600">
                {t("taskConversations.result.cache_read")}:
              </span>
              <span className="ml-2 font-medium">
                {result.usage.cache_read_input_tokens}
              </span>
            </div>
            {result.usage.service_tier && (
              <div>
                <span className="text-gray-600">
                  {t("taskConversations.result.service_tier")}:
                </span>
                <span className="ml-2 font-medium">
                  {result.usage.service_tier}
                </span>
              </div>
            )}
          </div>
        </div>

        <div className="pt-2 text-xs text-gray-500">
          <div>
            {t("taskConversations.result.session_id")}: {result.session_id}
          </div>
          <div>
            {t("taskConversations.result.created_at")}:{" "}
            {formatToLocal(result.created_at)}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
