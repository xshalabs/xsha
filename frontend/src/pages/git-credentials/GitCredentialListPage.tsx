import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { usePageTitle } from '@/hooks/usePageTitle';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { GitCredentialList } from '@/components/GitCredentialList';
import { apiService } from '@/lib/api/index';
import type { GitCredential, GitCredentialListParams } from '@/types/git-credentials';
import { GitCredentialType } from '@/types/git-credentials';
import { Plus, RefreshCw } from 'lucide-react';

const GitCredentialListPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  
  const [credentials, setCredentials] = useState<GitCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [typeFilter, setTypeFilter] = useState<GitCredentialType | undefined>();
  const [error, setError] = useState<string | null>(null);

  const pageSize = 10;

  usePageTitle(t('common.pageTitle.gitCredentials'));

  // 加载凭据列表
  const loadCredentials = async (params?: GitCredentialListParams) => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiService.gitCredentials.list({
        page: currentPage,
        page_size: pageSize,
        type: typeFilter,
        ...params,
      });
      
      setCredentials(response.credentials);
      setTotal(response.total);
      setTotalPages(response.total_pages);
    } catch (err: any) {
      setError(err.message || t('gitCredentials.messages.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 初始加载和依赖更新
  useEffect(() => {
    loadCredentials();
  }, [currentPage, typeFilter]);

  // 刷新列表
  const handleRefresh = () => {
    setCurrentPage(1);
    loadCredentials({ page: 1 });
  };

  // 编辑凭据
  const handleEdit = (credential: GitCredential) => {
    navigate(`/git-credentials/${credential.id}/edit`);
  };

  // 删除凭据
  const handleDelete = async (id: number) => {
    if (!confirm(t('gitCredentials.messages.deleteConfirm'))) return;
    
    try {
      await apiService.gitCredentials.delete(id);
      loadCredentials();
    } catch (err: any) {
      setError(err.message || t('gitCredentials.messages.deleteFailed'));
    }
  };

  // 页面变更
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 类型筛选变更
  const handleTypeFilterChange = (type: GitCredentialType | undefined) => {
    setTypeFilter(type);
    setCurrentPage(1);
  };

  const handleCreateNew = () => {
    navigate('/git-credentials/create');
  };

  return (
    <div className="container mx-auto p-6">
      <div className="max-w-6xl mx-auto">
        {/* 页面标题和操作按钮 */}
        <div className="mb-6">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold">{t('gitCredentials.title')}</h1>
              <p className="text-muted-foreground mt-1">{t('gitCredentials.subtitle', 'Manage your Git repository access credentials')}</p>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={handleRefresh}
                disabled={loading}
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                {t('gitCredentials.refresh')}
              </Button>
              <Button onClick={handleCreateNew}>
                <Plus className="w-4 h-4 mr-2" />
                {t('gitCredentials.create')}
              </Button>
            </div>
          </div>
        </div>

        {/* 错误信息 */}
        {error && (
          <Card className="mb-6 border-red-200 bg-red-50">
            <CardContent className="pt-6">
              <p className="text-red-600">{error}</p>
            </CardContent>
          </Card>
        )}

        {/* 统计信息 */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>{t('gitCredentials.statistics.title', 'Statistics')}</CardTitle>
            <CardDescription>{t('gitCredentials.statistics.description', 'Overview of credential usage')}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-primary">{total}</div>
                <div className="text-sm text-muted-foreground">{t('gitCredentials.statistics.total')}</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-accent">
                  {credentials.filter(c => c.is_active).length}
                </div>
                <div className="text-sm text-muted-foreground">{t('gitCredentials.statistics.active')}</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-chart-1">
                  {credentials.filter(c => c.type === GitCredentialType.SSH_KEY).length}
                </div>
                <div className="text-sm text-muted-foreground">{t('gitCredentials.statistics.sshKeys')}</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-chart-2">
                  {credentials.filter(c => c.type === GitCredentialType.TOKEN).length}
                </div>
                <div className="text-sm text-muted-foreground">{t('gitCredentials.statistics.tokens')}</div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 凭据列表 */}
        <GitCredentialList
          credentials={credentials}
          loading={loading}
          currentPage={currentPage}
          totalPages={totalPages}
          total={total}
          typeFilter={typeFilter}
          onPageChange={handlePageChange}
          onTypeFilterChange={handleTypeFilterChange}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onRefresh={handleRefresh}
        />
      </div>
    </div>
  );
};

export default GitCredentialListPage; 