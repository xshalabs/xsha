import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks/usePageTitle';
import { AuthProvider } from '@/contexts/AuthContext';
import { ThemeProvider } from '@/components/theme-provider';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { Layout } from '@/components/Layout';
import { LoginPage } from '@/pages/LoginPage';
import { DashboardPage } from '@/pages/DashboardPage';
import { ProjectsPage } from '@/pages/ProjectsPage';
import { TasksPage } from '@/pages/TasksPage';
import { GitCredentialsPage } from '@/pages/GitCredentialsPage';
import DevEnvironmentsPage from '@/pages/DevEnvironmentsPage';
import { AdminLogsPage } from '@/pages/AdminLogsPage';
import './App.css';

function NotFoundPage() {
  const { t } = useTranslation();
  
  // 设置页面标题
  usePageTitle('pageTitle.notFound');
  
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">404</h1>
        <p className="text-gray-600">{t('errors.pageNotFound')}</p>
      </div>
    </div>
  );
}

function App() {
  return (
    <ThemeProvider defaultTheme="system" storageKey="sleep0-ui-theme">
      <Router>
        <AuthProvider>
        <Routes>
          {/* 根路径重定向到仪表板 */}
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          
          {/* 登录页面 */}
          <Route path="/login" element={<LoginPage />} />
          
          {/* 受保护的仪表板页面 */}
          <Route 
            path="/dashboard" 
            element={
              <ProtectedRoute>
                <Layout>
                  <DashboardPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* 项目管理页面 */}
          <Route 
            path="/projects" 
            element={
              <ProtectedRoute>
                <Layout>
                  <ProjectsPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* 项目任务管理页面 */}
          <Route 
            path="/projects/:projectId/tasks" 
            element={
              <ProtectedRoute>
                <Layout>
                  <TasksPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* Git 凭据管理页面 */}
          <Route 
            path="/git-credentials" 
            element={
              <ProtectedRoute>
                <Layout>
                  <GitCredentialsPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* 开发环境管理页面 */}
          <Route 
            path="/dev-environments" 
            element={
              <ProtectedRoute>
                <Layout>
                  <DevEnvironmentsPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* 系统日志管理页面 */}
          <Route 
            path="/admin/logs" 
            element={
              <ProtectedRoute>
                <Layout>
                  <AdminLogsPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* 404页面 */}
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
        </AuthProvider>
      </Router>
    </ThemeProvider>
  );
}

export default App;
