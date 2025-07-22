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

interface TaskFormProps {
  task?: Task;
  projects: Project[];
  loading?: boolean;
  onSubmit: (data: TaskFormData) => Promise<void>;
  onCancel: () => void;
}

export function TaskForm({ 
  task, 
  projects, 
  loading = false, 
  onSubmit, 
  onCancel 
}: TaskFormProps) {
  const { t } = useTranslation();
  const isEdit = !!task;

  const [formData, setFormData] = useState<TaskFormData>({
    title: task?.title || '',
    description: task?.description || '',
    start_branch: task?.start_branch || 'main',
    project_id: task?.project_id || (projects[0]?.id || 0),
    dev_environment_id: task?.dev_environment_id || undefined,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [devEnvironments, setDevEnvironments] = useState<DevEnvironment[]>([]);
  const [loadingDevEnvs, setLoadingDevEnvs] = useState(false);

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

  // 表单验证
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.title.trim()) {
      newErrors.title = t('tasks.validation.titleRequired');
    }

    if (!formData.start_branch.trim()) {
      newErrors.start_branch = t('tasks.validation.branchRequired');
    }

    if (!formData.project_id) {
      newErrors.project_id = t('tasks.validation.projectRequired');
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
      await onSubmit(formData);
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

            {/* 任务描述 */}
            <div className="space-y-2">
              <Label htmlFor="description">
                {t('tasks.fields.description')}
              </Label>
              <textarea
                id="description"
                value={formData.description}
                onChange={(e) => handleChange('description', e.target.value)}
                placeholder={t('tasks.form.descriptionPlaceholder')}
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>

            {/* 项目选择 */}
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

            {/* 开发环境选择 */}
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
                          <span className={`text-xs px-2 py-1 rounded ${
                            env.status === 'running' ? 'bg-green-100 text-green-800' :
                            env.status === 'stopped' ? 'bg-gray-100 text-gray-800' :
                            env.status === 'error' ? 'bg-red-100 text-red-800' :
                            'bg-yellow-100 text-yellow-800'
                          }`}>
                            {env.status}
                          </span>
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

            {/* 起始分支 */}
            <div className="space-y-2">
              <Label htmlFor="start_branch">
                {t('tasks.fields.startBranch')} <span className="text-red-500">*</span>
              </Label>
              <Input
                id="start_branch"
                type="text"
                value={formData.start_branch}
                onChange={(e) => handleChange('start_branch', e.target.value)}
                placeholder={t('tasks.form.branchPlaceholder')}
                className={errors.start_branch ? 'border-red-500' : ''}
              />
              {errors.start_branch && (
                <p className="text-sm text-red-500">{errors.start_branch}</p>
              )}
              <p className="text-sm text-gray-500">
                {t('tasks.form.branchHint')}
              </p>
            </div>

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