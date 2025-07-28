import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { useTranslation } from 'react-i18next';
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
  AlertCircle
} from 'lucide-react';
import type { 
  TaskConversationResult,
  UsageStats,
  ParsedTaskConversationResult
} from '@/types/task-conversation-result';
import { taskConversationResultsApi } from '@/lib/api';

interface TaskConversationResultProps {
  conversationId: number;
  showHeader?: boolean;
  compact?: boolean;
}

export function TaskConversationResult({ 
  conversationId, 
  showHeader = true, 
  compact = false 
}: TaskConversationResultProps) {
  const { t } = useTranslation();
  const [result, setResult] = useState<ParsedTaskConversationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 解析使用统计JSON字符串
  const parseUsage = (usageString: string): UsageStats | null => {
    try {
      return JSON.parse(usageString);
    } catch {
      return null;
    }
  };

  // 格式化持续时间
  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}min`;
  };

  // 格式化成本
  const formatCost = (cost: number): string => {
    return `$${cost.toFixed(4)}`;
  };

  // 获取成功率颜色
  const getStatusColor = (isError: boolean) => {
    return isError ? 'destructive' : 'default';
  };

  // 获取状态图标
  const getStatusIcon = (isError: boolean) => {
    return isError ? (
      <XCircle className="h-4 w-4 text-red-500" />
    ) : (
      <CheckCircle className="h-4 w-4 text-green-500" />
    );
  };

  // 加载结果数据
  const fetchResult = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await taskConversationResultsApi.getByConversationId(conversationId);
      const resultData = response.data;
      
      // 解析使用统计
      const usage = parseUsage(resultData.usage);
      setResult({
        ...resultData,
        usage: usage || {
          input_tokens: 0,
          cache_creation_input_tokens: 0,
          cache_read_input_tokens: 0,
          output_tokens: 0
        }
      });
    } catch (err: any) {
      setError(err.message || t('task_conversation.result_load_failed'));
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
          <span>{t('common.loading')}</span>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center p-6">
          <AlertCircle className="h-6 w-6 text-red-500 mr-2" />
          <span className="text-red-600">{error}</span>
          <Button 
            variant="outline" 
            size="sm" 
            className="ml-2"
            onClick={fetchResult}
          >
            {t('common.retry')}
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
          <span>{t('task_conversation.no_result')}</span>
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
            {t('task_conversation.execution_result')}
            <Badge variant={getStatusColor(result.is_error)}>
              {result.subtype}
            </Badge>
          </CardTitle>
          <CardDescription>
            {t('task_conversation.result_description', { 
              sessionId: result.session_id.slice(0, 8) 
            })}
          </CardDescription>
        </CardHeader>
      )}
      
      <CardContent className="space-y-4">
        {/* 执行结果内容 */}
        <div>
          <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            {t('task_conversation.result_content')}
          </h4>
          <div className="bg-gray-50 p-3 rounded-md">
            <pre className="text-sm whitespace-pre-wrap">{result.result}</pre>
          </div>
        </div>

        <Separator />

        {/* 性能指标 */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <Clock className="h-4 w-4" />
              {t('task_conversation.duration')}
            </div>
            <div className="text-lg font-semibold">
              {formatDuration(result.duration_ms)}
            </div>
          </div>

          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <Activity className="h-4 w-4" />
              {t('task_conversation.api_duration')}
            </div>
            <div className="text-lg font-semibold">
              {formatDuration(result.duration_api_ms)}
            </div>
          </div>

          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <TrendingUp className="h-4 w-4" />
              {t('task_conversation.turns')}
            </div>
            <div className="text-lg font-semibold">
              {result.num_turns}
            </div>
          </div>

          <div className="text-center">
            <div className="flex items-center justify-center gap-1 text-sm text-gray-600 mb-1">
              <DollarSign className="h-4 w-4" />
              {t('task_conversation.cost')}
            </div>
            <div className="text-lg font-semibold">
              {formatCost(result.total_cost_usd)}
            </div>
          </div>
        </div>

        <Separator />

        {/* 使用统计 */}
        <div>
          <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
            <Code className="h-4 w-4" />
            {t('task_conversation.usage_stats')}
          </h4>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
            <div>
              <span className="text-gray-600">{t('task_conversation.input_tokens')}:</span>
              <span className="ml-2 font-medium">{result.usage.input_tokens}</span>
            </div>
            <div>
              <span className="text-gray-600">{t('task_conversation.output_tokens')}:</span>
              <span className="ml-2 font-medium">{result.usage.output_tokens}</span>
            </div>
            <div>
              <span className="text-gray-600">{t('task_conversation.cache_read')}:</span>
              <span className="ml-2 font-medium">{result.usage.cache_read_input_tokens}</span>
            </div>
            {result.usage.service_tier && (
              <div>
                <span className="text-gray-600">{t('task_conversation.service_tier')}:</span>
                <span className="ml-2 font-medium">{result.usage.service_tier}</span>
              </div>
            )}
          </div>
        </div>

        {/* 会话信息 */}
        <div className="pt-2 text-xs text-gray-500">
          <div>
            {t('task_conversation.session_id')}: {result.session_id}
          </div>
          <div>
            {t('task_conversation.created_at')}: {new Date(result.created_at).toLocaleString()}
          </div>
        </div>
      </CardContent>
    </Card>
  );
} 