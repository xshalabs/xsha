import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks/usePageTitle';
// import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { AdminOperationLogList } from '@/components/AdminOperationLogList';
import { LoginLogList } from '@/components/LoginLogList';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import { 
  FileText, 
  Shield, 
  // Activity,
  TrendingUp
} from 'lucide-react';
import type { 
  AdminOperationLog,
  LoginLog,
  AdminOperationLogListParams,
  LoginLogListParams,
  AdminOperationStatsResponse
} from '@/types/admin-logs';

type TabType = 'operationLogs' | 'loginLogs' | 'stats';

export const AdminLogsPage: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<TabType>('operationLogs');
  
  // 设置页面标题
  usePageTitle('common.pageTitle.adminLogs');

  // 操作日志状态
  const [operationLogs, setOperationLogs] = useState<AdminOperationLog[]>([]);
  const [operationLoading, setOperationLoading] = useState(false);
  const [operationCurrentPage, setOperationCurrentPage] = useState(1);
  const [operationTotalPages, setOperationTotalPages] = useState(1);
  const [operationTotal, setOperationTotal] = useState(0);
  const [operationFilters, setOperationFilters] = useState<AdminOperationLogListParams>({});
  
  // 登录日志状态
  const [loginLogs, setLoginLogs] = useState<LoginLog[]>([]);
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginCurrentPage, setLoginCurrentPage] = useState(1);
  const [loginTotalPages, setLoginTotalPages] = useState(1);
  const [loginTotal, setLoginTotal] = useState(0);
  const [loginFilters, setLoginFilters] = useState<LoginLogListParams>({});

  // 统计数据状态
  const [stats, setStats] = useState<AdminOperationStatsResponse | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);

  const pageSize = 20;

  // 加载操作日志
  const loadOperationLogs = async (params?: AdminOperationLogListParams) => {
    try {
      setOperationLoading(true);
      const response = await apiService.adminLogs.getOperationLogs({
        page: operationCurrentPage,
        page_size: pageSize,
        ...operationFilters,
        ...params,
      });
      
      setOperationLogs(response.logs);
      setOperationTotal(response.total);
      setOperationTotalPages(response.total_pages);
      if (params?.page) {
        setOperationCurrentPage(params.page);
      }
    } catch (err: any) {
      logError(err, 'Failed to load operation logs');
      console.error('Failed to load operation logs:', err);
    } finally {
      setOperationLoading(false);
    }
  };

  // 加载登录日志
  const loadLoginLogs = async (params?: LoginLogListParams) => {
    try {
      setLoginLoading(true);
      const response = await apiService.adminLogs.getLoginLogs({
        page: loginCurrentPage,
        page_size: pageSize,
        ...loginFilters,
        ...params,
      });
      
      setLoginLogs(response.logs);
      setLoginTotal(response.total);
      setLoginTotalPages(response.total_pages);
      if (params?.page) {
        setLoginCurrentPage(params.page);
      }
    } catch (err: any) {
      logError(err, 'Failed to load login logs');
      console.error('Failed to load login logs:', err);
    } finally {
      setLoginLoading(false);
    }
  };

  // 加载统计数据
  const loadStats = async () => {
    try {
      setStatsLoading(true);
      const response = await apiService.adminLogs.getOperationStats();
      setStats(response);
    } catch (err: any) {
      logError(err, 'Failed to load stats');
      console.error('Failed to load stats:', err);
    } finally {
      setStatsLoading(false);
    }
  };

  // 处理操作日志页面变化
  const handleOperationPageChange = (page: number) => {
    loadOperationLogs({ ...operationFilters, page });
  };

  // 处理操作日志筛选变化
  const handleOperationFiltersChange = (filters: AdminOperationLogListParams) => {
    setOperationFilters(filters);
    loadOperationLogs({ ...filters, page: 1 });
  };

  // 处理登录日志页面变化
  const handleLoginPageChange = (page: number) => {
    loadLoginLogs({ ...loginFilters, page });
  };

  // 处理登录日志筛选变化
  const handleLoginFiltersChange = (filters: LoginLogListParams) => {
    setLoginFilters(filters);
    loadLoginLogs({ ...filters, page: 1 });
  };

  // 查看操作日志详情
  const handleViewOperationDetail = async (id: number) => {
    try {
      const response = await apiService.adminLogs.getOperationLog(id);
      // 简化的详情展示 - 使用国际化
      const logInfo = [
        `${t('adminLogs.operationLogs.columns.id')}: ${response.log.id}`,
        `${t('adminLogs.operationLogs.columns.operation')}: ${response.log.operation}`,
        `${t('adminLogs.operationLogs.columns.resource')}: ${response.log.resource || 'N/A'}`,
        `${t('adminLogs.operationLogs.columns.username')}: ${response.log.user || 'N/A'}`,
        `${t('adminLogs.operationLogs.columns.description')}: ${response.log.details || 'N/A'}`,
        `${t('adminLogs.operationLogs.columns.time')}: ${new Date(response.log.timestamp).toLocaleString()}`
      ].join('\n\n');
      
      alert(logInfo);
    } catch (err: any) {
      logError(err, 'Failed to load operation log detail');
      console.error('Failed to load operation log detail:', err);
    }
  };

  // 刷新当前标签页数据
  // const handleRefresh = () => {
  //   switch (activeTab) {
  //     case 'operationLogs':
  //       loadOperationLogs();
  //       break;
  //     case 'loginLogs':
  //       loadLoginLogs();
  //       break;
  //     case 'stats':
  //       loadStats();
  //       break;
  //   }
  // };

  // 初始加载数据
  useEffect(() => {
    switch (activeTab) {
      case 'operationLogs':
        if (operationLogs.length === 0) {
          loadOperationLogs();
        }
        break;
      case 'loginLogs':
        if (loginLogs.length === 0) {
          loadLoginLogs();
        }
        break;
      case 'stats':
        if (!stats) {
          loadStats();
        }
        break;
    }
  }, [activeTab]);

  const renderTabContent = () => {
    switch (activeTab) {
      case 'operationLogs':
        return (
          <AdminOperationLogList
            logs={operationLogs}
            loading={operationLoading}
            currentPage={operationCurrentPage}
            totalPages={operationTotalPages}
            total={operationTotal}
            filters={operationFilters}
            onPageChange={handleOperationPageChange}
            onFiltersChange={handleOperationFiltersChange}
            onRefresh={() => loadOperationLogs()}
            onViewDetail={handleViewOperationDetail}
          />
        );
      
      case 'loginLogs':
        return (
          <LoginLogList
            logs={loginLogs}
            loading={loginLoading}
            currentPage={loginCurrentPage}
            totalPages={loginTotalPages}
            total={loginTotal}
            filters={loginFilters}
            onPageChange={handleLoginPageChange}
            onFiltersChange={handleLoginFiltersChange}
            onRefresh={() => loadLoginLogs()}
          />
        );
      
      case 'stats':
        return (
          <div className="space-y-6">
            <div>
              <h2 className="text-2xl font-bold">{t('adminLogs.stats.title')}</h2>
              <p className="text-gray-600">{t('adminLogs.stats.description')}</p>
            </div>

            {statsLoading ? (
              <div className="flex items-center justify-center h-64">
                <div className="text-gray-500">{t('common.loading')}</div>
              </div>
            ) : stats ? (
              <div className="space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle>{t('adminLogs.stats.operationStats')}</CardTitle>
                    <CardDescription>
                      {t('adminLogs.stats.timeRange')}: {stats.start_time} ~ {stats.end_time}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
                      {Object.entries(stats.operation_stats).map(([operation, count]) => (
                        <div key={operation} className="text-center p-4 bg-gray-50 rounded-lg">
                          <div className="text-2xl font-bold text-blue-600">{count}</div>
                          <div className="text-sm text-gray-600">
                            {t(`adminLogs.operationLogs.operations.${operation}`)}
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>{t('adminLogs.stats.resourceStats')}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      {Object.entries(stats.resource_stats).map(([resource, count]) => (
                        <div key={resource} className="text-center p-4 bg-gray-50 rounded-lg">
                          <div className="text-2xl font-bold text-green-600">{count}</div>
                          <div className="text-sm text-gray-600 capitalize">{resource}</div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500">{t('adminLogs.stats.noStatsAvailable')}</p>
              </div>
            )}
          </div>
        );
      
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                {t('adminLogs.title')}
              </h1>
              <p className="mt-2 text-sm text-gray-600">
                {t('adminLogs.description')}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* 标签导航 */}
          <div className="border-b border-gray-200 mb-6">
            <nav className="-mb-px flex space-x-8">
              <button
                onClick={() => setActiveTab('operationLogs')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'operationLogs'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <FileText className="w-4 h-4 inline mr-2" />
                {t('adminLogs.operationLogs.title')}
              </button>
              
              <button
                onClick={() => setActiveTab('loginLogs')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'loginLogs'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <Shield className="w-4 h-4 inline mr-2" />
                {t('adminLogs.loginLogs.title')}
              </button>
              
              <button
                onClick={() => setActiveTab('stats')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'stats'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <TrendingUp className="w-4 h-4 inline mr-2" />
                {t('adminLogs.stats.title')}
              </button>
            </nav>
          </div>

          {/* 标签内容 */}
          {renderTabContent()}
        </div>
      </div>
    </div>
  );
}; 