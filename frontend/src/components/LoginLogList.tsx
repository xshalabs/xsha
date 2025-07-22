import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { 
  ChevronLeft, 
  ChevronRight,
  Filter,
  RefreshCw,
  Calendar,
  User,
  Shield,
  CheckCircle,
  XCircle,
  Globe
} from 'lucide-react';
import type { 
  LoginLog, 
  LoginLogListParams 
} from '@/types/admin-logs';

interface LoginLogListProps {
  logs: LoginLog[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  filters: LoginLogListParams;
  onPageChange: (page: number) => void;
  onFiltersChange: (filters: LoginLogListParams) => void;
  onRefresh: () => void;
}

export const LoginLogList: React.FC<LoginLogListProps> = ({
  logs,
  loading,
  currentPage,
  totalPages,
  total,
  filters,
  onPageChange,
  onFiltersChange,
  onRefresh,
}) => {
  const { t } = useTranslation();
  const [showFilters, setShowFilters] = useState(false);
  const [localFilters, setLocalFilters] = useState<LoginLogListParams>(filters);

  const handleFilterChange = (key: keyof LoginLogListParams, value: string | undefined) => {
    setLocalFilters(prev => ({
      ...prev,
      [key]: value === '' ? undefined : value
    }));
  };

  const applyFilters = () => {
    onFiltersChange({
      ...localFilters,
      page: 1 // 重置到第一页
    });
  };

  const resetFilters = () => {
    const emptyFilters: LoginLogListParams = {};
    setLocalFilters(emptyFilters);
    onFiltersChange(emptyFilters);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getStatusIcon = (success: boolean) => {
    return success ? 
      <CheckCircle className="w-4 h-4 text-green-500" /> : 
      <XCircle className="w-4 h-4 text-red-500" />;
  };

  const getReasonText = (reason: string) => {
    if (!reason) return '';
    const reasonKey = `adminLogs.loginLogs.reasons.${reason}`;
    const translatedReason = t(reasonKey);
    return translatedReason === reasonKey ? reason : translatedReason;
  };

  if (loading && logs.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">{t('common.loading')}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 头部操作栏 */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h2 className="text-2xl font-bold">{t('adminLogs.loginLogs.title')}</h2>
          <p className="text-gray-600">{t('adminLogs.loginLogs.description')}</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
          >
            <Filter className="w-4 h-4 mr-2" />
            {t('adminLogs.common.search')}
          </Button>
          <Button variant="outline" onClick={onRefresh} disabled={loading}>
            <RefreshCw className="w-4 h-4 mr-2" />
            {t('adminLogs.common.refresh')}
          </Button>
        </div>
      </div>

      {/* 筛选器 */}
      {showFilters && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">{t('adminLogs.loginLogs.filters.all')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <Label htmlFor="username">{t('adminLogs.loginLogs.filters.username')}</Label>
                <Input
                  id="username"
                  value={localFilters.username || ''}
                  onChange={(e) => handleFilterChange('username', e.target.value)}
                  placeholder={t('adminLogs.loginLogs.filters.username')}
                />
              </div>
            </div>

            <div className="flex gap-2 mt-4">
              <Button onClick={applyFilters}>
                {t('adminLogs.common.apply')}
              </Button>
              <Button variant="outline" onClick={resetFilters}>
                {t('adminLogs.common.reset')}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 统计信息 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center space-x-2">
              <Shield className="w-8 h-8 text-blue-500" />
              <div>
                <p className="text-2xl font-bold">{total}</p>
                <p className="text-sm text-gray-600">{t('adminLogs.common.total')}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center space-x-2">
              <CheckCircle className="w-8 h-8 text-green-500" />
              <div>
                <p className="text-2xl font-bold text-green-600">
                  {logs.filter(log => log.success).length}
                </p>
                <p className="text-sm text-gray-600">{t('adminLogs.loginLogs.status.success')}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center space-x-2">
              <XCircle className="w-8 h-8 text-red-500" />
              <div>
                <p className="text-2xl font-bold text-red-600">
                  {logs.filter(log => !log.success).length}
                </p>
                <p className="text-sm text-gray-600">{t('adminLogs.loginLogs.status.failed')}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 日志列表 */}
      <div className="space-y-2">
        {logs.length === 0 ? (
          <Card>
            <CardContent className="text-center py-8">
              <p className="text-gray-500">{t('adminLogs.loginLogs.messages.noLogs')}</p>
            </CardContent>
          </Card>
        ) : (
          logs.map((log) => (
            <Card key={log.id} className="hover:shadow-md transition-shadow">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4 flex-1">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(log.success)}
                      <span className={`font-medium ${log.success ? 'text-green-600' : 'text-red-600'}`}>
                        {log.success ? t('adminLogs.loginLogs.status.success') : t('adminLogs.loginLogs.status.failed')}
                      </span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <User className="w-4 h-4 text-gray-400" />
                      <span className="text-sm font-medium">{log.username}</span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Globe className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600">{log.ip}</span>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Calendar className="w-4 h-4 text-gray-400" />
                      <span className="text-sm text-gray-600">{formatDate(log.login_time)}</span>
                    </div>
                  </div>
                </div>

                {!log.success && log.reason && (
                  <div className="mt-2 text-sm text-red-600 bg-red-50 p-2 rounded">
                    <strong>{t('adminLogs.loginLogs.columns.reason')}:</strong> {getReasonText(log.reason)}
                  </div>
                )}

                {log.user_agent && (
                  <div className="mt-2 text-xs text-gray-500 truncate">
                    <strong>{t('adminLogs.loginLogs.columns.userAgent')}:</strong> {log.user_agent}
                  </div>
                )}
              </CardContent>
            </Card>
          ))
        )}
      </div>

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-600">
            {t('adminLogs.common.page')} {currentPage} / {totalPages}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage <= 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}; 