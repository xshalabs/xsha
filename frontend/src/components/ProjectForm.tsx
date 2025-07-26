import { useState, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiService } from '@/lib/api/index';
import { logError } from '@/lib/errors';
import { useTranslation } from 'react-i18next';
import type { 
  Project, 
  CreateProjectRequest, 
  UpdateProjectRequest, 
  ProjectFormData, 
  GitProtocolType 
} from '@/types/project';

interface ProjectFormProps {
  project?: Project;
  onSubmit?: (project: Project) => void;
  onCancel?: () => void;
}

interface CredentialOption {
  id: number;
  name: string;
  type: string;
  username: string;
  is_active: boolean;
}

export function ProjectForm({ project, onSubmit, onCancel }: ProjectFormProps) {
  const { t } = useTranslation();
  const isEdit = !!project;
  
  // 表单状态
  const [formData, setFormData] = useState<ProjectFormData>({
    name: project?.name || '',
    description: project?.description || '',
    repo_url: project?.repo_url || '',
    protocol: project?.protocol || 'https',
    credential_id: project?.credential_id
  });
  
  const [credentials, setCredentials] = useState<CredentialOption[]>([]);
  const [branches, setBranches] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [credentialsLoading, setCredentialsLoading] = useState(false);
  const [branchesLoading, setBranchesLoading] = useState(false);
  const [urlParsing, setUrlParsing] = useState(false);
  const [credentialValidating, setCredentialValidating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [urlParseTimeout, setUrlParseTimeout] = useState<NodeJS.Timeout | null>(null);
  const [accessValidated, setAccessValidated] = useState(false);
  const [accessError, setAccessError] = useState<string | null>(null);

  // 加载兼容的凭据
  const loadCredentials = async (protocol: GitProtocolType) => {
    try {
      setCredentialsLoading(true);
      const response = await apiService.projects.getCompatibleCredentials(protocol);
      setCredentials(response.credentials);
    } catch (error) {
      logError(error as Error, 'Failed to load credentials');
      setCredentials([]);
    } finally {
      setCredentialsLoading(false);
    }
  };

  // 验证仓库访问权限并获取分支列表
  const validateAccessAndFetchBranches = async (repoUrl: string, credentialId?: number) => {
    if (!repoUrl.trim()) {
      setBranches([]);
      setAccessValidated(false);
      setAccessError(null);
      return;
    }

    try {
      setCredentialValidating(true);
      setBranchesLoading(true);
      setAccessError(null);
      setAccessValidated(false);

      // 首先验证访问权限
      const validateResponse = await apiService.projects.validateAccess({
        repo_url: repoUrl,
        credential_id: credentialId
      });

      if (!validateResponse.can_access) {
        setAccessError(validateResponse.error || '无法访问仓库');
        setBranches([]);
        return;
      }

      // 如果验证成功，获取分支列表
      const branchesResponse = await apiService.projects.fetchBranches({
        repo_url: repoUrl,
        credential_id: credentialId
      });

      if (branchesResponse.result.can_access) {
        setBranches(branchesResponse.result.branches);
        setAccessValidated(true);
      } else {
        setAccessError(branchesResponse.result.error_message || '获取分支列表失败');
        setBranches([]);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '验证仓库访问失败';
      setAccessError(errorMessage);
      setBranches([]);
      logError(error as Error, 'Failed to validate repository access');
    } finally {
      setCredentialValidating(false);
      setBranchesLoading(false);
    }
  };

  // 解析仓库 URL 并自动设置协议类型
  const parseRepositoryUrl = useCallback(async (url: string) => {
    if (!url.trim()) {
      return;
    }

    // 简单检查是否是 Git URL 格式
    const gitUrlPattern = /^(https?:\/\/|git@|ssh:\/\/)/;
    if (!gitUrlPattern.test(url)) {
      return;
    }

    try {
      setUrlParsing(true);
      const response = await apiService.projects.parseUrl(url);
      
      if (response.result.is_valid) {
        const detectedProtocol = response.result.protocol as GitProtocolType;
        
        // 只有当检测到的协议与当前不同时才更新
        if (detectedProtocol !== formData.protocol) {
          setFormData(prev => ({
            ...prev,
            protocol: detectedProtocol,
            credential_id: undefined // 清除之前选择的凭据
          }));
        }
      }
    } catch (error) {
      // 静默处理错误，不影响用户体验
      logError(error as Error, 'Failed to parse repository URL');
    } finally {
      setUrlParsing(false);
    }
  }, [formData.protocol]);

  useEffect(() => {
    if (formData.protocol) {
      loadCredentials(formData.protocol);
    }
  }, [formData.protocol]);

  // 监听credential变化，自动验证访问权限
  useEffect(() => {
    if (formData.repo_url && formData.credential_id) {
      validateAccessAndFetchBranches(formData.repo_url, formData.credential_id);
    } else if (formData.repo_url && !formData.credential_id) {
      // 尝试无credential访问（适用于公开仓库）
      validateAccessAndFetchBranches(formData.repo_url);
    } else {
      setBranches([]);
      setAccessValidated(false);
      setAccessError(null);
    }
  }, [formData.repo_url, formData.credential_id]);

  // 组件卸载时清理定时器
  useEffect(() => {
    return () => {
      if (urlParseTimeout) {
        clearTimeout(urlParseTimeout);
      }
    };
  }, [urlParseTimeout]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t('projects.validation.nameRequired');
    }

    if (!formData.repo_url.trim()) {
      newErrors.repo_url = t('projects.validation.repoUrlRequired');
    } else {
      // 简单的 URL 验证
      const urlPattern = /^(https?:\/\/|git@)/;
      if (!urlPattern.test(formData.repo_url)) {
        newErrors.repo_url = t('projects.validation.invalidRepoUrl');
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleInputChange = (field: keyof ProjectFormData, value: string | number | undefined) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
    
    // 清除对应字段的错误
    if (errors[field]) {
      setErrors(prev => ({
        ...prev,
        [field]: ''
      }));
    }

    // 如果是仓库 URL 字段，延时解析协议类型
    if (field === 'repo_url' && typeof value === 'string') {
      // 清除之前的定时器
      if (urlParseTimeout) {
        clearTimeout(urlParseTimeout);
      }

      // 设置新的定时器
      const timeoutId = setTimeout(() => {
        parseRepositoryUrl(value);
      }, 500); // 500ms 延时

      setUrlParseTimeout(timeoutId);
    }

    // 如果是credential字段变化，清除之前的验证状态
    if (field === 'credential_id') {
      setAccessValidated(false);
      setAccessError(null);
      setBranches([]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      setError(null);

      let result: Project;

      if (isEdit && project) {
        const updateData: UpdateProjectRequest = {
          name: formData.name,
          description: formData.description,
          repo_url: formData.repo_url,
          protocol: formData.protocol,
          credential_id: formData.credential_id
        };

        await apiService.projects.update(project.id, updateData);
        
        // 获取更新后的项目信息
        const response = await apiService.projects.get(project.id);
        result = response.project;
      } else {
        const createData: CreateProjectRequest = {
          name: formData.name,
          description: formData.description,
          repo_url: formData.repo_url,
          protocol: formData.protocol,
          credential_id: formData.credential_id
        };

        const response = await apiService.projects.create(createData);
        result = response.project;
      }

      if (onSubmit) {
        onSubmit(result);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 
        isEdit ? t('projects.messages.updateFailed') : t('projects.messages.createFailed');
      setError(errorMessage);
      logError(error as Error, `Failed to ${isEdit ? 'update' : 'create'} project`);
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    if (isEdit && project) {
      setFormData({
        name: project.name,
        description: project.description,
        repo_url: project.repo_url,
        protocol: project.protocol,
        credential_id: project.credential_id
      });
    } else {
      setFormData({
        name: '',
        description: '',
        repo_url: '',
        protocol: 'https',
        credential_id: undefined
      });
    }
    setErrors({});
    setError(null);
    setBranches([]);
    setAccessValidated(false);
    setAccessError(null);
  };

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardHeader>
        <CardTitle>{isEdit ? t('projects.edit') : t('projects.create')}</CardTitle>
        <CardDescription>
          {isEdit ? t('projects.editDescription') : t('projects.createDescription')}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* 错误提示 */}
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-red-700">{error}</p>
            </div>
          )}

          <div className="grid grid-cols-1 gap-6">
            {/* 项目名称 */}
            <div className="space-y-2">
              <Label htmlFor="name">{t('projects.name')} *</Label>
              <Input
                id="name"
                type="text"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                placeholder={t('projects.placeholders.name')}
                className={errors.name ? 'border-red-500' : ''}
              />
              {errors.name && <p className="text-sm text-red-500">{errors.name}</p>}
            </div>

            {/* 项目描述 */}
            <div className="space-y-2">
              <Label htmlFor="description">{t('projects.description')}</Label>
              <Input
                id="description"
                type="text"
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                placeholder={t('projects.placeholders.description')}
              />
            </div>

            {/* 仓库 URL */}
            <div className="space-y-2">
              <Label htmlFor="repo_url">{t('projects.repoUrl')} *</Label>
              <div className="relative">
                <Input
                  id="repo_url"
                  type="text"
                  value={formData.repo_url}
                  onChange={(e) => handleInputChange('repo_url', e.target.value)}
                  placeholder={t('projects.placeholders.repoUrl')}
                  className={errors.repo_url ? 'border-red-500' : ''}
                />
                {urlParsing && (
                  <div className="absolute right-2 top-1/2 transform -translate-y-1/2">
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                  </div>
                )}
              </div>
              {errors.repo_url && <p className="text-sm text-red-500">{errors.repo_url}</p>}
              {!errors.repo_url && formData.repo_url && (
                <p className="text-sm text-gray-500">
                  {t('projects.protocolAutoDetected')}: {formData.protocol.toUpperCase()}
                </p>
              )}
            </div>

            {/* 凭据选择 */}
            <div className="space-y-2">
              <Label htmlFor="credential_id">{t('projects.credential')}</Label>
              {credentialsLoading ? (
                <div className="text-sm text-gray-500">{t('common.loading')}</div>
              ) : (
                <select
                  id="credential_id"
                  value={formData.credential_id || ''}
                  onChange={(e) => handleInputChange('credential_id', e.target.value ? Number(e.target.value) : undefined)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">{t('projects.placeholders.selectCredential')}</option>
                  {credentials.map((credential) => (
                    <option key={credential.id} value={credential.id}>
                      {credential.name} ({credential.type} - {credential.username})
                      {!credential.is_active && ` - ${t('projects.status.inactive')}`}
                    </option>
                  ))}
                </select>
              )}
              <p className="text-sm text-gray-500">{t('projects.credentialHelp')}</p>
              
              {/* 凭据验证状态 */}
              {credentialValidating && (
                <div className="flex items-center space-x-2 text-sm text-blue-600">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                  <span>{t('projects.repository.validatingAccess')}</span>
                </div>
              )}
              
              {accessValidated && !credentialValidating && (
                <div className="text-sm text-green-600">
                  ✓ {t('projects.repository.accessValidated')}
                </div>
              )}
              
              {accessError && !credentialValidating && (
                <div className="text-sm text-red-600">
                  ✗ {accessError}
                </div>
              )}
            </div>


          </div>

          {/* 操作按钮 */}
          <div className="flex justify-end gap-4">
            <Button type="button" variant="outline" onClick={handleReset}>
              {t('common.reset')}
            </Button>
            {onCancel && (
              <Button type="button" variant="outline" onClick={onCancel}>
                {t('common.cancel')}
              </Button>
            )}
            <Button type="submit" disabled={loading}>
              {loading ? t('common.loading') : (isEdit ? t('common.save') : t('projects.create'))}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
} 