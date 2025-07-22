import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiService } from '@/lib/api';
import type { GitCredential, GitCredentialType, GitCredentialFormData } from '@/types/git-credentials';
import { GitCredentialType as CredentialTypes } from '@/types/git-credentials';

interface GitCredentialFormProps {
  credential?: GitCredential;
  onSuccess: () => void;
  onCancel: () => void;
}

export const GitCredentialForm: React.FC<GitCredentialFormProps> = ({
  credential,
  onSuccess,
  onCancel,
}) => {
  const [formData, setFormData] = useState<GitCredentialFormData>({
    name: '',
    description: '',
    type: CredentialTypes.PASSWORD,
    username: '',
    password: '',
    token: '',
    private_key: '',
    public_key: '',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!credential;

  // 初始化表单数据
  useEffect(() => {
    if (credential) {
      setFormData({
        name: credential.name,
        description: credential.description,
        type: credential.type,
        username: credential.username,
        password: '',
        token: '',
        private_key: '',
        public_key: credential.public_key || '',
      });
    }
  }, [credential]);

  // 表单字段更新
  const updateField = (field: keyof GitCredentialFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  // 表单验证
  const validateForm = (): string | null => {
    if (!formData.name.trim()) return '请输入凭据名称';
    if (!formData.type) return '请选择凭据类型';
    if (!formData.username.trim()) return '请输入用户名';

    switch (formData.type) {
      case CredentialTypes.PASSWORD:
        if (!formData.password) return '请输入密码';
        break;
      case CredentialTypes.TOKEN:
        if (!formData.token) return '请输入访问令牌';
        break;
      case CredentialTypes.SSH_KEY:
        if (!formData.private_key) return '请输入私钥';
        break;
    }

    return null;
  };

  // 构建提交数据
  const buildSubmitData = () => {
    const secretData: Record<string, string> = {};

    switch (formData.type) {
      case CredentialTypes.PASSWORD:
        if (formData.password) secretData.password = formData.password;
        break;
      case CredentialTypes.TOKEN:
        if (formData.token) secretData.password = formData.token; // token存储在password字段
        break;
      case CredentialTypes.SSH_KEY:
        if (formData.private_key) secretData.private_key = formData.private_key;
        if (formData.public_key) secretData.public_key = formData.public_key;
        break;
    }

    return {
      name: formData.name.trim(),
      description: formData.description.trim(),
      type: formData.type,
      username: formData.username.trim(),
      secret_data: secretData,
    };
  };

  // 提交表单
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const submitData = buildSubmitData();

      if (isEditing && credential) {
        // 编辑模式
        await apiService.gitCredentials.update(credential.id, submitData);
      } else {
        // 创建模式
        await apiService.gitCredentials.create(submitData);
      }

      onSuccess();
    } catch (err: any) {
      setError(err.message || `${isEditing ? '更新' : '创建'}凭据失败`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{isEditing ? '编辑' : '创建'} Git 凭据</CardTitle>
        <CardDescription>
          {isEditing ? '更新现有的' : '创建新的'} Git 仓库访问凭据
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* 错误信息 */}
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-red-600 text-sm">{error}</p>
            </div>
          )}

          {/* 基本信息 */}
          <div className="space-y-4">
            <div>
              <Label htmlFor="name">凭据名称 *</Label>
              <Input
                id="name"
                type="text"
                value={formData.name}
                onChange={(e) => updateField('name', e.target.value)}
                placeholder="输入凭据名称"
                required
              />
            </div>

            <div>
              <Label htmlFor="description">描述</Label>
              <Input
                id="description"
                type="text"
                value={formData.description}
                onChange={(e) => updateField('description', e.target.value)}
                placeholder="输入凭据描述（可选）"
              />
            </div>

            <div>
              <Label htmlFor="type">凭据类型 *</Label>
              <select
                id="type"
                value={formData.type}
                onChange={(e) => updateField('type', e.target.value as GitCredentialType)}
                className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
                disabled={isEditing} // 编辑时不允许修改类型
              >
                <option value={CredentialTypes.PASSWORD}>密码认证</option>
                <option value={CredentialTypes.TOKEN}>访问令牌</option>
                <option value={CredentialTypes.SSH_KEY}>SSH 密钥</option>
              </select>
            </div>

            <div>
              <Label htmlFor="username">用户名 *</Label>
              <Input
                id="username"
                type="text"
                value={formData.username}
                onChange={(e) => updateField('username', e.target.value)}
                placeholder="输入用户名"
                required
              />
            </div>
          </div>

          {/* 凭据信息 */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium">凭据信息</h3>
            
            {formData.type === CredentialTypes.PASSWORD && (
              <div>
                <Label htmlFor="password">密码 *</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => updateField('password', e.target.value)}
                  placeholder="输入密码"
                  required
                />
              </div>
            )}

            {formData.type === CredentialTypes.TOKEN && (
              <div>
                <Label htmlFor="token">访问令牌 *</Label>
                <Input
                  id="token"
                  type="password"
                  value={formData.token}
                  onChange={(e) => updateField('token', e.target.value)}
                  placeholder="输入访问令牌"
                  required
                />
              </div>
            )}

            {formData.type === CredentialTypes.SSH_KEY && (
              <>
                <div>
                  <Label htmlFor="private_key">私钥 *</Label>
                  <textarea
                    id="private_key"
                    value={formData.private_key}
                    onChange={(e) => updateField('private_key', e.target.value)}
                    placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;..."
                    className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500 min-h-[120px] font-mono text-sm"
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="public_key">公钥</Label>
                  <textarea
                    id="public_key"
                    value={formData.public_key}
                    onChange={(e) => updateField('public_key', e.target.value)}
                    placeholder="ssh-rsa AAAAB3NzaC1yc2E..."
                    className="w-full p-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500 min-h-[80px] font-mono text-sm"
                  />
                </div>
              </>
            )}
          </div>

          {/* 操作按钮 */}
          <div className="flex justify-end space-x-3">
            <Button type="button" variant="outline" onClick={onCancel}>
              取消
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? '保存中...' : (isEditing ? '更新' : '创建')}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
};

export default GitCredentialForm; 