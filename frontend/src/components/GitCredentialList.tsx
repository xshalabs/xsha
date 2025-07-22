import React from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import type { GitCredential, GitCredentialType } from '@/types/git-credentials';
import { GitCredentialType as CredentialTypes } from '@/types/git-credentials';
import { 
  Edit, 
  Trash2, 
  Eye, 
  EyeOff, 
  Key, 
  Shield, 
  User, 
  Clock, 
  ChevronLeft, 
  ChevronRight,
  Filter,
  X
} from 'lucide-react';

interface GitCredentialListProps {
  credentials: GitCredential[];
  loading: boolean;
  currentPage: number;
  totalPages: number;
  total: number;
  typeFilter?: GitCredentialType;
  onPageChange: (page: number) => void;
  onTypeFilterChange: (type: GitCredentialType | undefined) => void;
  onEdit: (credential: GitCredential) => void;
  onDelete: (id: number) => void;
  onToggle: (id: number, isActive: boolean) => void;
  onRefresh: () => void;
}

export const GitCredentialList: React.FC<GitCredentialListProps> = ({
  credentials,
  loading,
  currentPage,
  totalPages,
  total,
  typeFilter,
  onPageChange,
  onTypeFilterChange,
  onEdit,
  onDelete,
  onToggle,
  onRefresh: _onRefresh,
}) => {
  const { t } = useTranslation();

  // 获取凭据类型图标
  const getTypeIcon = (type: GitCredentialType) => {
    switch (type) {
      case CredentialTypes.PASSWORD:
        return <Key className="w-4 h-4" />;
      case CredentialTypes.TOKEN:
        return <Shield className="w-4 h-4" />;
      case CredentialTypes.SSH_KEY:
        return <User className="w-4 h-4" />;
      default:
        return <Key className="w-4 h-4" />;
    }
  };

  // 获取凭据类型名称
  const getTypeName = (type: GitCredentialType) => {
    switch (type) {
      case CredentialTypes.PASSWORD:
        return t('gitCredentials.filter.password');
      case CredentialTypes.TOKEN:
        return t('gitCredentials.filter.token');
      case CredentialTypes.SSH_KEY:
        return t('gitCredentials.filter.sshKey');
      default:
        return t('gitCredentials.unknown', 'Unknown');
    }
  };

  // 格式化时间
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  // 计算分页信息
  const startItem = (currentPage - 1) * 10 + 1;
  const endItem = Math.min(currentPage * 10, total);

  return (
    <div className="space-y-4">
      {/* 筛选器 */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Filter className="w-4 h-4" />
              <span className="text-sm font-medium">{t('gitCredentials.filterTitle', 'Filter')}:</span>
            </div>
            <div className="flex gap-2">
              <Button
                variant={!typeFilter ? "default" : "outline"}
                size="sm"
                onClick={() => onTypeFilterChange(undefined)}
              >
                {t('gitCredentials.filter.all')}
              </Button>
              <Button
                variant={typeFilter === CredentialTypes.PASSWORD ? "default" : "outline"}
                size="sm"
                onClick={() => onTypeFilterChange(CredentialTypes.PASSWORD)}
              >
                {t('gitCredentials.filter.password')}
              </Button>
              <Button
                variant={typeFilter === CredentialTypes.TOKEN ? "default" : "outline"}
                size="sm"
                onClick={() => onTypeFilterChange(CredentialTypes.TOKEN)}
              >
                {t('gitCredentials.filter.token')}
              </Button>
              <Button
                variant={typeFilter === CredentialTypes.SSH_KEY ? "default" : "outline"}
                size="sm"
                onClick={() => onTypeFilterChange(CredentialTypes.SSH_KEY)}
              >
                {t('gitCredentials.filter.sshKey')}
              </Button>
            </div>
            {typeFilter && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onTypeFilterChange(undefined)}
                className="text-gray-500"
              >
                <X className="w-4 h-4" />
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 凭据列表 */}
      {loading ? (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div className="space-y-2">
                    <div className="h-4 bg-gray-200 rounded w-48"></div>
                    <div className="h-3 bg-gray-200 rounded w-32"></div>
                  </div>
                  <div className="flex gap-2">
                    <div className="h-8 bg-gray-200 rounded w-16"></div>
                    <div className="h-8 bg-gray-200 rounded w-16"></div>
                    <div className="h-8 bg-gray-200 rounded w-16"></div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : credentials.length === 0 ? (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center py-8">
              <Key className="w-12 h-12 mx-auto text-gray-400 mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">
                {typeFilter ? t('gitCredentials.messages.noMatchingCredentials') : t('gitCredentials.messages.noCredentials')}
              </h3>
              <p className="text-gray-600 mb-4">
                {typeFilter 
                  ? t('gitCredentials.noMatchingType', `No ${getTypeName(typeFilter)} credentials found`, { type: getTypeName(typeFilter) })
                  : t('gitCredentials.messages.noCredentialsDesc')
                }
              </p>
              {typeFilter && (
                <Button 
                  variant="outline" 
                  onClick={() => onTypeFilterChange(undefined)}
                >
                  {t('gitCredentials.messages.clearFilter')}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {credentials.map((credential) => (
            <Card key={credential.id} className={credential.is_active ? '' : 'opacity-60'}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <div className="flex items-center gap-2">
                        {getTypeIcon(credential.type)}
                        <h3 className="font-medium text-gray-900">{credential.name}</h3>
                      </div>
                      <span className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full">
                        {getTypeName(credential.type)}
                      </span>
                      <div className="flex items-center gap-1">
                        {credential.is_active ? (
                          <Eye className="w-4 h-4 text-green-500" />
                        ) : (
                          <EyeOff className="w-4 h-4 text-gray-400" />
                        )}
                        <span className={`text-xs ${credential.is_active ? 'text-green-600' : 'text-gray-500'}`}>
                          {credential.is_active ? t('gitCredentials.active') : t('gitCredentials.inactive')}
                        </span>
                      </div>
                    </div>
                    
                    {credential.description && (
                      <p className="text-gray-600 text-sm mb-2">{credential.description}</p>
                    )}
                    
                    <div className="flex items-center gap-4 text-sm text-gray-500">
                      <div className="flex items-center gap-1">
                        <User className="w-3 h-3" />
                        <span>{credential.username}</span>
                      </div>
                      <div className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        <span>{t('gitCredentials.createdAt')}: {formatDate(credential.created_at)}</span>
                      </div>
                      {credential.last_used && (
                        <div className="flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          <span>{t('gitCredentials.lastUsed')}: {formatDate(credential.last_used)}</span>
                        </div>
                      )}
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onToggle(credential.id, !credential.is_active)}
                      title={credential.is_active ? t('gitCredentials.deactivateTooltip', 'Deactivate credential') : t('gitCredentials.activateTooltip', 'Activate credential')}
                    >
                      {credential.is_active ? (
                        <EyeOff className="w-4 h-4" />
                      ) : (
                        <Eye className="w-4 h-4" />
                      )}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onEdit(credential)}
                      title={t('gitCredentials.edit')}
                    >
                      <Edit className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onDelete(credential.id)}
                      className="text-red-600 hover:text-red-700 hover:border-red-300"
                      title={t('gitCredentials.delete')}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 分页信息和控制 */}
      {!loading && credentials.length > 0 && (
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">
                {t('gitCredentials.pagination.showing')} {startItem} {t('gitCredentials.pagination.to')} {endItem} {t('gitCredentials.pagination.of')} {total} {t('gitCredentials.pagination.items')}
              </div>
              
              {totalPages > 1 && (
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange(currentPage - 1)}
                    disabled={currentPage <= 1}
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </Button>
                  
                  <div className="flex items-center gap-1">
                    {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                      let pageNum;
                      if (totalPages <= 5) {
                        pageNum = i + 1;
                      } else if (currentPage <= 3) {
                        pageNum = i + 1;
                      } else if (currentPage >= totalPages - 2) {
                        pageNum = totalPages - 4 + i;
                      } else {
                        pageNum = currentPage - 2 + i;
                      }
                      
                      return (
                        <Button
                          key={pageNum}
                          variant={currentPage === pageNum ? "default" : "outline"}
                          size="sm"
                          onClick={() => onPageChange(pageNum)}
                          className="w-8 h-8 p-0"
                        >
                          {pageNum}
                        </Button>
                      );
                    })}
                  </div>
                  
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onPageChange(currentPage + 1)}
                    disabled={currentPage >= totalPages}
                  >
                    <ChevronRight className="w-4 h-4" />
                  </Button>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default GitCredentialList; 