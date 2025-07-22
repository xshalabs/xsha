import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiService } from '@/lib/api';
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
    default_branch: project?.default_branch || 'main',
    credential_id: project?.credential_id
  });
  
  const [credentials, setCredentials] = useState<CredentialOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [credentialsLoading, setCredentialsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});

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

  useEffect(() => {
    if (formData.protocol) {
      loadCredentials(formData.protocol);
    }
  }, [formData.protocol]);

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

    if (!formData.protocol) {
      newErrors.protocol = t('projects.validation.protocolRequired');
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
  };

  const handleProtocolChange = (protocol: GitProtocolType) => {
    setFormData(prev => ({
      ...prev,
      protocol,
      credential_id: undefined // 清除之前选择的凭据
    }));
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
        // 更新项目
        const updateData: UpdateProjectRequest = {
          name: formData.name !== project.name ? formData.name : undefined,
          description: formData.description !== project.description ? formData.description : undefined,
          repo_url: formData.repo_url !== project.repo_url ? formData.repo_url : undefined,
          default_branch: formData.default_branch !== project.default_branch ? formData.default_branch : undefined,
          credential_id: formData.credential_id !== project.credential_id ? formData.credential_id : undefined
        };

        await apiService.projects.update(project.id, updateData);
        
        // 获取更新后的项目信息
        const response = await apiService.projects.get(project.id);
        result = response.project;
      } else {
        // 创建项目
        const createData: CreateProjectRequest = {
          name: formData.name,
          description: formData.description,
          repo_url: formData.repo_url,
          protocol: formData.protocol,
          default_branch: formData.default_branch,
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
        default_branch: project.default_branch,
        credential_id: project.credential_id
      });
    } else {
      setFormData({
        name: '',
        description: '',
        repo_url: '',
        protocol: 'https',
        default_branch: 'main',
        credential_id: undefined
      });
    }
    setErrors({});
    setError(null);
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
              <Input
                id="repo_url"
                type="text"
                value={formData.repo_url}
                onChange={(e) => handleInputChange('repo_url', e.target.value)}
                placeholder={t('projects.placeholders.repoUrl')}
                className={errors.repo_url ? 'border-red-500' : ''}
              />
              {errors.repo_url && <p className="text-sm text-red-500">{errors.repo_url}</p>}
            </div>

            {/* 协议选择 */}
            <div className="space-y-2">
              <Label>{t('projects.protocol')} *</Label>
              <div className="flex gap-4">
                <label className="flex items-center space-x-2">
                  <input
                    type="radio"
                    name="protocol"
                    value="https"
                    checked={formData.protocol === 'https'}
                    onChange={(e) => handleProtocolChange(e.target.value as GitProtocolType)}
                    className="radio"
                  />
                  <span>HTTPS</span>
                </label>
                <label className="flex items-center space-x-2">
                  <input
                    type="radio"
                    name="protocol"
                    value="ssh"
                    checked={formData.protocol === 'ssh'}
                    onChange={(e) => handleProtocolChange(e.target.value as GitProtocolType)}
                    className="radio"
                  />
                  <span>SSH</span>
                </label>
              </div>
              {errors.protocol && <p className="text-sm text-red-500">{errors.protocol}</p>}
            </div>

            {/* 默认分支 */}
            <div className="space-y-2">
              <Label htmlFor="default_branch">{t('projects.defaultBranch')}</Label>
              <Input
                id="default_branch"
                type="text"
                value={formData.default_branch}
                onChange={(e) => handleInputChange('default_branch', e.target.value)}
                placeholder={t('projects.placeholders.defaultBranch')}
              />
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