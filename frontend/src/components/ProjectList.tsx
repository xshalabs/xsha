import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { apiService } from '@/lib/api';
import { logError } from '@/lib/errors';
import { useTranslation } from 'react-i18next';
import type { Project, ProjectListParams, GitProtocolType } from '@/types/project';

interface ProjectListProps {
  onEdit?: (project: Project) => void;
  onDelete?: (id: number) => void;
  onUse?: (id: number) => void;
  onCreateNew?: () => void;
}

export function ProjectList({ onEdit, onDelete, onUse, onCreateNew }: ProjectListProps) {
  const { t } = useTranslation();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [totalPages, setTotalPages] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [protocolFilter, setProtocolFilter] = useState<GitProtocolType | undefined>();

  const pageSize = 20;

  const loadProjects = async (page = 1, protocol?: GitProtocolType) => {
    try {
      setLoading(true);
      setError(null);
      
      const params: ProjectListParams = {
        page,
        page_size: pageSize,
        ...(protocol && { protocol })
      };
      
      const response = await apiService.projects.list(params);
      setProjects(response.projects);
      setTotalPages(response.total_pages);
      setTotal(response.total);
      setCurrentPage(page);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('projects.messages.loadFailed');
      setError(errorMessage);
      logError(error as Error, 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProjects(1, protocolFilter);
  }, [protocolFilter]);

  const handleRefresh = () => {
    loadProjects(currentPage, protocolFilter);
  };

  const handlePageChange = (page: number) => {
    loadProjects(page, protocolFilter);
  };

  const handleProtocolFilterChange = (protocol?: GitProtocolType) => {
    setProtocolFilter(protocol);
    setCurrentPage(1);
  };



  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getProtocolBadgeColor = (protocol: GitProtocolType) => {
    return protocol === 'https' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800';
  };

  if (loading && projects.length === 0) {
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
          <h2 className="text-2xl font-bold">{t('projects.title')}</h2>
          <p className="text-gray-600">{t('projects.description')}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleRefresh} disabled={loading}>
            {t('projects.refresh')}
          </Button>
          {onCreateNew && (
            <Button onClick={onCreateNew}>
              {t('projects.create')}
            </Button>
          )}
        </div>
      </div>

      {/* 筛选器 */}
      <div className="flex flex-wrap gap-2">
        <Button
          variant={protocolFilter === undefined ? "default" : "outline"}
          size="sm"
          onClick={() => handleProtocolFilterChange(undefined)}
        >
          {t('projects.filter.all')}
        </Button>
        <Button
          variant={protocolFilter === 'https' ? "default" : "outline"}
          size="sm"
          onClick={() => handleProtocolFilterChange('https')}
        >
          HTTPS
        </Button>
        <Button
          variant={protocolFilter === 'ssh' ? "default" : "outline"}
          size="sm"
          onClick={() => handleProtocolFilterChange('ssh')}
        >
          SSH
        </Button>
      </div>

      {/* 统计信息 */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">{t('projects.statistics.total')}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{total}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">HTTPS</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {projects.filter(p => p.protocol === 'https').length}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">SSH</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {projects.filter(p => p.protocol === 'ssh').length}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 错误显示 */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-red-700">{error}</p>
        </div>
      )}

      {/* 项目列表 */}
      {projects.length === 0 && !loading ? (
        <Card>
          <CardContent className="text-center py-12">
            <div className="text-gray-500 mb-4">
              {protocolFilter ? t('projects.messages.noMatchingProjects') : t('projects.messages.noProjects')}
            </div>
            {protocolFilter ? (
              <Button variant="outline" onClick={() => handleProtocolFilterChange(undefined)}>
                {t('projects.messages.clearFilter')}
              </Button>
            ) : (
              onCreateNew && (
                <div>
                  <p className="text-gray-400 mb-4">{t('projects.messages.noProjectsDesc')}</p>
                  <Button onClick={onCreateNew}>
                    {t('projects.create')}
                  </Button>
                </div>
              )
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {projects.map((project) => (
            <Card key={project.id}>
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <CardTitle className="flex items-center gap-2">
                      {project.name}
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getProtocolBadgeColor(project.protocol)}`}>
                        {project.protocol.toUpperCase()}
                      </span>
                    </CardTitle>
                    <CardDescription>{project.description}</CardDescription>
                  </div>
                  <div className="flex gap-2">
                    {onUse && (
                      <Button size="sm" variant="outline" onClick={() => onUse(project.id)}>
                        {t('projects.use')}
                      </Button>
                    )}
                    {onEdit && (
                      <Button size="sm" variant="outline" onClick={() => onEdit(project)}>
                        {t('common.edit')}
                      </Button>
                    )}
                    {onDelete && (
                      <Button 
                        size="sm" 
                        variant="outline" 
                        onClick={() => onDelete(project.id)}
                        className="text-red-600 hover:text-red-700"
                      >
                        {t('common.delete')}
                      </Button>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 text-sm">
                  <div>
                    <span className="font-medium text-gray-700">{t('projects.repoUrl')}:</span>
                    <div className="text-blue-600 truncate">{project.repo_url}</div>
                  </div>
                  <div>
                    <span className="font-medium text-gray-700">{t('projects.defaultBranch')}:</span>
                    <div>{project.default_branch || 'main'}</div>
                  </div>
                  {project.credential && (
                    <div>
                      <span className="font-medium text-gray-700">{t('projects.credential')}:</span>
                      <div>{project.credential.name}</div>
                    </div>
                  )}
                  <div>
                    <span className="font-medium text-gray-700">{t('projects.createdAt')}:</span>
                    <div>{formatDate(project.created_at)}</div>
                  </div>
                  {project.last_used && (
                    <div>
                      <span className="font-medium text-gray-700">{t('projects.lastUsed')}:</span>
                      <div>{formatDate(project.last_used)}</div>
                    </div>
                  )}
                  <div>
                    <span className="font-medium text-gray-700">{t('projects.createdBy')}:</span>
                    <div>{project.created_by}</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-600">
            {t('projects.pagination.showing')} {((currentPage - 1) * pageSize) + 1} {t('projects.pagination.to')} {Math.min(currentPage * pageSize, total)} {t('projects.pagination.of')} {total} {t('projects.pagination.items')}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage === 1}
            >
              {t('common.previous')}
            </Button>
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
              <Button
                key={page}
                variant={page === currentPage ? "default" : "outline"}
                size="sm"
                onClick={() => handlePageChange(page)}
              >
                {page}
              </Button>
            ))}
            <Button
              variant="outline"
              size="sm"
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
            >
              {t('common.next')}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
} 