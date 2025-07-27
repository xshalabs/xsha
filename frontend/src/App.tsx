import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { usePageTitle } from '@/hooks/usePageTitle';
import { AuthProvider } from '@/contexts/AuthContext';
import { ThemeProvider } from '@/components/theme-provider';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { Layout } from '@/components/Layout';
import { LoginPage } from '@/pages/LoginPage';
import { DashboardPage } from '@/pages/DashboardPage';
import { AdminLogsPage } from '@/pages/AdminLogsPage';

// 项目管理页面
import ProjectListPage from '@/pages/projects/ProjectListPage';
import ProjectCreatePage from '@/pages/projects/ProjectCreatePage';
import ProjectEditPage from '@/pages/projects/ProjectEditPage';

// 开发环境页面
import DevEnvironmentListPage from '@/pages/dev-environments/DevEnvironmentListPage';
import DevEnvironmentCreatePage from '@/pages/dev-environments/DevEnvironmentCreatePage';
import DevEnvironmentEditPage from '@/pages/dev-environments/DevEnvironmentEditPage';

// Git凭据页面
import GitCredentialListPage from '@/pages/git-credentials/GitCredentialListPage';
import GitCredentialCreatePage from '@/pages/git-credentials/GitCredentialCreatePage';
import GitCredentialEditPage from '@/pages/git-credentials/GitCredentialEditPage';

// 任务管理页面
import TaskListPage from '@/pages/tasks/TaskListPage';
import TaskCreatePage from '@/pages/tasks/TaskCreatePage';
import TaskEditPage from '@/pages/tasks/TaskEditPage';
import TaskConversationPage from '@/pages/tasks/TaskConversationPage';

import './App.css';

function NotFoundPage() {
  const { t } = useTranslation();
  
  // 设置页面标题
  usePageTitle('common.pageTitle.notFound');
  
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-foreground mb-4">404</h1>
        <p className="text-muted-foreground">{t('errors.pageNotFound')}</p>
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
          
          {/* 项目管理页面组 */}
          <Route 
            path="/projects" 
            element={
              <ProtectedRoute>
                <Layout>
                  <ProjectListPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/projects/create" 
            element={
              <ProtectedRoute>
                <Layout>
                  <ProjectCreatePage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/projects/:id/edit" 
            element={
              <ProtectedRoute>
                <Layout>
                  <ProjectEditPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* 开发环境管理页面组 */}
          <Route 
            path="/dev-environments" 
            element={
              <ProtectedRoute>
                <Layout>
                  <DevEnvironmentListPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/dev-environments/create" 
            element={
              <ProtectedRoute>
                <Layout>
                  <DevEnvironmentCreatePage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/dev-environments/:id/edit" 
            element={
              <ProtectedRoute>
                <Layout>
                  <DevEnvironmentEditPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* Git 凭据管理页面组 */}
          <Route 
            path="/git-credentials" 
            element={
              <ProtectedRoute>
                <Layout>
                  <GitCredentialListPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/git-credentials/create" 
            element={
              <ProtectedRoute>
                <Layout>
                  <GitCredentialCreatePage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/git-credentials/:id/edit" 
            element={
              <ProtectedRoute>
                <Layout>
                  <GitCredentialEditPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          
          {/* 项目任务管理页面组 */}
          <Route 
            path="/projects/:projectId/tasks" 
            element={
              <ProtectedRoute>
                <Layout>
                  <TaskListPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/projects/:projectId/tasks/create" 
            element={
              <ProtectedRoute>
                <Layout>
                  <TaskCreatePage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/projects/:projectId/tasks/:taskId/edit" 
            element={
              <ProtectedRoute>
                <Layout>
                  <TaskEditPage />
                </Layout>
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/projects/:projectId/tasks/:taskId/conversation" 
            element={
              <ProtectedRoute>
                <Layout>
                  <TaskConversationPage />
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
