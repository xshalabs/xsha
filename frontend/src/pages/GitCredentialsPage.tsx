import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks/usePageTitle';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { GitCredentialList } from '@/components/GitCredentialList';
import { GitCredentialForm } from '@/components/GitCredentialForm';
import { apiService } from '@/lib/api';
import type { GitCredential, GitCredentialListParams } from '@/types/git-credentials';
import { GitCredentialType } from '@/types/git-credentials';
import { Plus, RefreshCw } from 'lucide-react';

export const GitCredentialsPage: React.FC = () => {
  const { t } = useTranslation();
  const [credentials, setCredentials] = useState<GitCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingCredential, setEditingCredential] = useState<GitCredential | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [typeFilter, setTypeFilter] = useState<GitCredentialType | undefined>();
  const [error, setError] = useState<string | null>(null);

  // 设置页面标题
  usePageTitle('common.pageTitle.gitCredentials');

  const pageSize = 10;

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

  // 创建凭据成功
  const handleCreateSuccess = () => {
    setShowCreateForm(false);
    handleRefresh();
  };

  // 编辑凭据成功
  const handleEditSuccess = () => {
    setEditingCredential(null);
    loadCredentials();
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

  // 切换凭据状态
  const handleToggle = async (id: number, isActive: boolean) => {
    try {
      await apiService.gitCredentials.toggle(id, isActive);
      loadCredentials();
    } catch (err: any) {
      setError(err.message || t('gitCredentials.messages.toggleFailed'));
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

  if (showCreateForm) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-2xl mx-auto">
          <div className="mb-6">
            <Button 
              variant="outline" 
              onClick={() => setShowCreateForm(false)}
              className="mb-4"
            >
              ← {t('gitCredentials.backToList')}
            </Button>
            <h1 className="text-2xl font-bold">{t('gitCredentials.create')}</h1>
          </div>
          <GitCredentialForm
            onSuccess={handleCreateSuccess}
            onCancel={() => setShowCreateForm(false)}
          />
        </div>
      </div>
    );
  }

  if (editingCredential) {
    return (
      <div className="container mx-auto p-6">
        <div className="max-w-2xl mx-auto">
          <div className="mb-6">
            <Button 
              variant="outline" 
              onClick={() => setEditingCredential(null)}
              className="mb-4"
            >
              ← {t('gitCredentials.backToList')}
            </Button>
            <h1 className="text-2xl font-bold">{t('gitCredentials.edit')}</h1>
          </div>
          <GitCredentialForm
            credential={editingCredential}
            onSuccess={handleEditSuccess}
            onCancel={() => setEditingCredential(null)}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      <div className="max-w-6xl mx-auto">
        {/* 页面标题和操作按钮 */}
        <div className="mb-6">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold">{t('gitCredentials.title')}</h1>
              <p className="text-gray-600 mt-1">{t('gitCredentials.subtitle', 'Manage your Git repository access credentials')}</p>
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
              <Button onClick={() => setShowCreateForm(true)}>
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
                <div className="text-2xl font-bold text-blue-600">{total}</div>
                <div className="text-sm text-gray-600">{t('gitCredentials.statistics.total')}</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-green-600">
                  {credentials.filter(c => c.is_active).length}
                </div>
                <div className="text-sm text-gray-600">{t('gitCredentials.statistics.active')}</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-orange-600">
                  {credentials.filter(c => c.type === GitCredentialType.SSH_KEY).length}
                </div>
                <div className="text-sm text-gray-600">{t('gitCredentials.statistics.sshKeys')}</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-purple-600">
                  {credentials.filter(c => c.type === GitCredentialType.TOKEN).length}
                </div>
                <div className="text-sm text-gray-600">{t('gitCredentials.statistics.tokens')}</div>
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
          onEdit={setEditingCredential}
          onDelete={handleDelete}
          onToggle={handleToggle}
          onRefresh={handleRefresh}
        />
      </div>
    </div>
  );
};

export default GitCredentialsPage; 