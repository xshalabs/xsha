import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '@/contexts/AuthContext';
import { apiService } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { LanguageSwitcher } from '@/components/LanguageSwitcher';
import { ROUTES } from '@/lib/constants';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';

export const DashboardPage: React.FC = () => {
  const { t } = useTranslation();
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [healthStatus, setHealthStatus] = useState<{ status: string; message?: string; lang?: string } | null>(null);

  useEffect(() => {
    // 获取后端健康状态，测试国际化对接
    const fetchHealthStatus = async () => {
      try {
        const status = await apiService.healthCheck();
        setHealthStatus(status);
      } catch (error) {
        console.error('Failed to fetch health status:', error);
      }
    };

    fetchHealthStatus();
  }, []);

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
      // 即使登出失败，也要跳转到登录页面，因为logout函数会清除本地状态
      navigate('/login');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                {t('dashboard.title')}
              </h1>
              <p className="mt-2 text-sm text-gray-600">
                {t('dashboard.welcome')}, {user}!
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <LanguageSwitcher />
              <Button 
                onClick={handleLogout}
                variant="outline"
              >
                {t('auth.logout')}
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* 系统状态卡片 - 显示后端国际化信息 */}
          {healthStatus && (
            <div className="mb-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center">
                    <span className="w-3 h-3 bg-green-400 rounded-full mr-2"></span>
                    {t('dashboard.systemStatus.title')}
                  </CardTitle>
                  <CardDescription>{t('dashboard.systemStatus.description')}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <p><strong>Status:</strong> {healthStatus.status}</p>
                    {healthStatus.message && <p><strong>Message:</strong> {healthStatus.message}</p>}
                    {healthStatus.lang && <p><strong>Backend Language:</strong> {healthStatus.lang}</p>}
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <Card 
              className="cursor-pointer hover:shadow-lg transition-shadow" 
              onClick={() => navigate(ROUTES.projects)}
            >
              <CardHeader>
                <CardTitle>{t('dashboard.projectManagement.title')}</CardTitle>
                <CardDescription>
                  {t('dashboard.projectManagement.description')}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-600">
                  {t('dashboard.projectManagement.content')}
                </p>
                <Button 
                  className="mt-4" 
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(ROUTES.projects);
                  }}
                >
                  {t('projects.title')}
                </Button>
              </CardContent>
            </Card>

            <Card 
              className="cursor-pointer hover:shadow-lg transition-shadow" 
              onClick={() => navigate(ROUTES.gitCredentials)}
            >
              <CardHeader>
                <CardTitle>{t('dashboard.gitCredentials.title')}</CardTitle>
                <CardDescription>
                  {t('dashboard.gitCredentials.description')}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-600">
                  {t('dashboard.gitCredentials.content')}
                </p>
                <Button 
                  className="mt-4" 
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(ROUTES.gitCredentials);
                  }}
                >
                  {t('dashboard.gitCredentials.manage')}
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>{t('dashboard.systemStatus.title')}</CardTitle>
                <CardDescription>
                  {t('dashboard.systemStatus.description')}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-gray-600">
                  {t('dashboard.systemStatus.content')}
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}; 