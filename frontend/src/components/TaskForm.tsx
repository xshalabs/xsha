import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useTranslation } from 'react-i18next';
import { Save, X } from 'lucide-react';
import type { Task, TaskFormData } from '@/types/task';
import type { Project } from '@/types/project';
import type { DevEnvironment } from '@/types/dev-environment';
import { devEnvironmentsApi } from '@/lib/api/dev-environments';
import { projectsApi } from '@/lib/api/projects';

interface TaskFormProps {
  task?: Task;
  projects: Project[];
  defaultProjectId?: number; // 默认选择的项目ID
  loading?: boolean;
  onSubmit: (data: TaskFormData | { title: string }) => Promise<void>;
  onCancel: () => void;
}

export function TaskForm({ 
  task, 
  projects, 
  defaultProjectId,
  loading = false, 
  onSubmit, 
  onCancel 
}: TaskFormProps) {
  const { t } = useTranslation();
  const isEdit = !!task;

  const [formData, setFormData] = useState<TaskFormData>({
    title: task?.title || '',
    start_branch: task?.start_branch || 'main',
    project_id: task?.project_id || defaultProjectId || (projects[0]?.id || 0),
    dev_environment_id: task?.dev_environment_id || undefined,
    requirement_desc: '', // 仅在创建时使用
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [devEnvironments, setDevEnvironments] = useState<DevEnvironment[]>([]);
  const [loadingDevEnvs, setLoadingDevEnvs] = useState(false);
  const [availableBranches, setAvailableBranches] = useState<string[]>([]);
  const [loadingBranches, setLoadingBranches] = useState(false);
  const [branchError, setBranchError] = useState<string>('');

  // 加载开发环境列表
  useEffect(() => {
    const loadDevEnvironments = async () => {
      try {
        setLoadingDevEnvs(true);
        const response = await devEnvironmentsApi.list();
        setDevEnvironments(response.environments || []);
      } catch (error) {
        console.error('Failed to load dev environments:', error);
        setDevEnvironments([]);
      } finally {
        setLoadingDevEnvs(false);
      }
    };

    loadDevEnvironments();
  }, []);

  // 获取项目分支列表
  const fetchProjectBranches = async (projectId: number) => {
    const selectedProject = projects.find(p => p.id === projectId);
    if (!selectedProject || isEdit) return;

    try {
      setLoadingBranches(true);
      setBranchError('');
      setAvailableBranches([]);

      const response = await projectsApi.fetchBranches({
        repo_url: selectedProject.repo_url,
        credential_id: selectedProject.credential_id || undefined,
      });

      if (response.result.can_access) {
        setAvailableBranches(response.result.branches);
        // 如果当前选择的分支不在列表中，选择第一个分支或默认的main
        const currentBranch = formData.start_branch;
        if (!response.result.branches.includes(currentBranch)) {
          const defaultBranch = response.result.branches.includes('main') 
            ? 'main' 
            : response.result.branches.includes('master')
            ? 'master'
            : response.result.branches[0] || 'main';
          setFormData(prev => ({ ...prev, start_branch: defaultBranch }));
        }
      } else {
        setBranchError(response.result.error_message || t('tasks.errors.fetchBranchesFailed'));
      }
    } catch (error) {
      console.error('Failed to fetch branches:', error);
      setBranchError(t('tasks.errors.fetchBranchesFailed'));
    } finally {
      setLoadingBranches(false);
    }
  };

  // 当projects或defaultProjectId变化时更新project_id
  useEffect(() => {
    if (!isEdit && !task) {
      const newProjectId = defaultProjectId || (projects.length > 0 ? projects[0].id : 0);
      if (newProjectId && newProjectId !== formData.project_id) {
        setFormData(prev => ({ ...prev, project_id: newProjectId }));
      }
    }
  }, [projects, defaultProjectId, isEdit, task, formData.project_id]);

  // 当项目改变时获取分支
  useEffect(() => {
    if (formData.project_id && !isEdit) {
      fetchProjectBranches(formData.project_id);
    }
  }, [formData.project_id, isEdit]);

  // 表单验证
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.title.trim()) {
      newErrors.title = t('tasks.validation.titleRequired');
    }

    // 只在创建模式下验证其他字段
    if (!isEdit) {
      if (!formData.start_branch.trim()) {
        newErrors.start_branch = t('tasks.validation.branchRequired');
      }

      if (!formData.project_id) {
        newErrors.project_id = t('tasks.validation.projectRequired');
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 处理表单提交
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setSubmitting(true);
    try {
      // 在编辑模式下只发送标题字段
      const submitData = isEdit 
        ? { title: formData.title }
        : { ...formData, include_branches: true };
      
      await onSubmit(submitData);
      // 成功后由父组件处理导航和消息显示
    } catch (error) {
      console.error('Failed to submit task:', error);
      // 错误处理可以在这里添加
    } finally {
      setSubmitting(false);
    }
  };

  // 处理表单字段变化
  const handleChange = (field: keyof TaskFormData, value: string | number | undefined) => {
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

  return (
    <div className="max-w-2xl mx-auto">
      <Card>
        <CardHeader>
          <CardTitle>
            {isEdit ? t('tasks.actions.edit') : t('tasks.actions.create')}
          </CardTitle>
          <CardDescription>
            {isEdit 
              ? t('tasks.form.editDescription') 
              : t('tasks.form.createDescription')
            }
          </CardDescription>
        </CardHeader>
        
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* 任务标题 */}
            <div className="space-y-2">
              <Label htmlFor="title">
                {t('tasks.fields.title')} <span className="text-red-500">*</span>
              </Label>
              <Input
                id="title"
                type="text"
                value={formData.title}
                onChange={(e) => handleChange('title', e.target.value)}
                placeholder={t('tasks.form.titlePlaceholder')}
                className={errors.title ? 'border-red-500' : ''}
              />
              {errors.title && (
                <p className="text-sm text-red-500">{errors.title}</p>
              )}
            </div>

            {/* 任务描述 - 仅在创建模式下显示 */}
            {!isEdit && (
              <div className="space-y-2">
                <Label htmlFor="requirement_desc">
                  {t('tasks.fields.requirementDesc')}
                </Label>
                <textarea
                  id="requirement_desc"
                  value={formData.requirement_desc || ''}
                  onChange={(e) => handleChange('requirement_desc', e.target.value)}
                  placeholder={t('tasks.form.requirementDescPlaceholder')}
                  rows={4}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
                <p className="text-sm text-gray-500">
                  {t('tasks.form.requirementDescHint')}
                </p>
              </div>
            )}

            {/* 项目选择 - 仅在创建模式下显示 */}
            {!isEdit && (
              <div className="space-y-2">
                <Label htmlFor="project">
                  {t('tasks.fields.project')} <span className="text-red-500">*</span>
                </Label>
                <Select
                  value={formData.project_id.toString()}
                  onValueChange={(value) => handleChange('project_id', parseInt(value))}
                >
                  <SelectTrigger className={errors.project_id ? 'border-red-500' : ''}>
                    <SelectValue placeholder={t('tasks.form.selectProject')} />
                  </SelectTrigger>
                  <SelectContent>
                    {projects.map((project) => (
                      <SelectItem key={project.id} value={project.id.toString()}>
                        <div className="flex items-center space-x-2">
                          <span>{project.name}</span>
                          {!project.is_active && (
                            <span className="text-xs text-gray-500">({t('common.inactive')})</span>
                          )}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {errors.project_id && (
                  <p className="text-sm text-red-500">{errors.project_id}</p>
                )}
              </div>
            )}

            {/* 开发环境选择 - 仅在创建模式下显示 */}
            {!isEdit && (
              <div className="space-y-2">
                <Label htmlFor="dev_environment">
                  {t('tasks.fields.devEnvironment')}
                </Label>
                <Select
                  value={formData.dev_environment_id?.toString() || 'none'}
                  onValueChange={(value) => handleChange('dev_environment_id', value === 'none' ? undefined : parseInt(value))}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t('tasks.form.selectDevEnvironment')} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">
                      {t('tasks.form.noDevEnvironment')}
                    </SelectItem>
                    {loadingDevEnvs ? (
                      <SelectItem value="loading" disabled>
                        {t('common.loading')}...
                      </SelectItem>
                    ) : (
                      devEnvironments.map((env) => (
                        <SelectItem key={env.id} value={env.id.toString()}>
                          <div className="flex items-center space-x-2">
                            <span>{env.name}</span>
                            <span className="text-xs text-gray-500">({env.type})</span>
                          </div>
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
                <p className="text-sm text-gray-500">
                  {t('tasks.form.devEnvironmentHint')}
                </p>
              </div>
            )}

            {/* 起始分支 - 仅在创建模式下显示 */}
            {!isEdit && (
              <div className="space-y-2">
                <Label htmlFor="start_branch">
                  {t('tasks.fields.startBranch')} <span className="text-red-500">*</span>
                </Label>
                {availableBranches.length > 0 ? (
                  <Select
                    value={formData.start_branch}
                    onValueChange={(value) => handleChange('start_branch', value)}
                  >
                    <SelectTrigger className={errors.start_branch ? 'border-red-500' : ''}>
                      <SelectValue placeholder={t('tasks.form.selectBranch')} />
                    </SelectTrigger>
                    <SelectContent>
                      {availableBranches.map((branch) => (
                        <SelectItem key={branch} value={branch}>
                          {branch}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : loadingBranches ? (
                  <div className="flex items-center space-x-2 px-3 py-2 border border-gray-300 rounded-md bg-gray-50">
                    <span className="text-sm text-gray-500">{t('common.loading')}...</span>
                  </div>
                ) : (
                  <Input
                    id="start_branch"
                    type="text"
                    value={formData.start_branch}
                    onChange={(e) => handleChange('start_branch', e.target.value)}
                    placeholder={t('tasks.form.branchPlaceholder')}
                    className={errors.start_branch ? 'border-red-500' : ''}
                  />
                )}
                {errors.start_branch && (
                  <p className="text-sm text-red-500">{errors.start_branch}</p>
                )}
                {branchError && (
                  <p className="text-sm text-orange-500">{branchError}</p>
                )}
                <p className="text-sm text-gray-500">
                  {availableBranches.length > 0 
                    ? t('tasks.form.branchFromRepository')
                    : t('tasks.form.branchHint')
                  }
                </p>
              </div>
            )}

            {/* 操作按钮 */}
            <div className="flex items-center justify-end space-x-4 pt-6">
              <Button
                type="button"
                variant="outline"
                onClick={onCancel}
                disabled={submitting}
              >
                <X className="w-4 h-4 mr-2" />
                {t('common.cancel')}
              </Button>
              <Button
                type="submit"
                disabled={submitting || loading}
              >
                <Save className="w-4 h-4 mr-2" />
                {submitting 
                  ? t('common.saving') 
                  : isEdit 
                    ? t('common.save') 
                    : t('tasks.actions.create')
                }
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
} 