import {
  HashRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks/usePageTitle";
import { AuthProvider } from "@/contexts/AuthContext";
import { PageTitleProvider } from "@/contexts/PageTitleContext";
import { ThemeProvider } from "@/components/ThemeProvider";
import { Toaster } from "@/components/ui/sonner";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { Layout } from "@/components/Layout";
import { LoginPage } from "@/pages/LoginPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { OperationLogsPage } from "@/pages/logs/OperationLogsPage";
import { LoginLogsPage } from "@/pages/logs/LoginLogsPage";
import { AuditStatsPage } from "@/pages/logs/AuditStatsPage";
import ProjectListPage from "@/pages/projects/ProjectListPage";
import ProjectKanbanPage from "@/pages/projects/ProjectKanbanPage";
import EnvironmentListPage from "@/pages/environments/EnvironmentListPage";
import CredentialListPage from "@/pages/credentials/CredentialListPage";
import NotifierListPage from "@/pages/notifiers/NotifierListPage";
import MCPListPage from "@/pages/mcp/MCPListPage";
import ProviderListPage from "@/pages/providers/ProviderListPage";
import SettingsPage from "@/pages/settings/Settings";
import AdminListPage from "@/pages/admin/AdminListPage";
import ChangePasswordPage from "@/pages/user/ChangePasswordPage";
import UpdateAvatarPage from "@/pages/user/UpdateAvatarPage";

import "./App.css";

function NotFoundPage() {
  const { t } = useTranslation();

  usePageTitle("common.pageTitle.notFound");

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-foreground mb-4">404</h1>
        <p className="text-muted-foreground">{t("errors.pageNotFound")}</p>
      </div>
    </div>
  );
}

function App() {
  return (
    <ThemeProvider defaultTheme="system" storageKey="xsha-ui-theme">
      <Router>
        <AuthProvider>
          <PageTitleProvider>
            <Routes>
              <Route path="/" element={<Navigate to="/dashboard" replace />} />

              <Route path="/login" element={<LoginPage />} />

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
                path="/projects/:projectId/kanban"
                element={
                  <ProtectedRoute>
                    <ProjectKanbanPage />
                  </ProtectedRoute>
                }
              />

              <Route
                path="/environments"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <EnvironmentListPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/credentials"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <CredentialListPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/notifiers"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <NotifierListPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/mcp"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <MCPListPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/providers"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <ProviderListPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/logs/operation-logs"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <OperationLogsPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/logs/login-logs"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <LoginLogsPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/logs/stats"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <AuditStatsPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/settings"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <SettingsPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/admin"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <AdminListPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/user/change-password"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <ChangePasswordPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/user/update-avatar"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <UpdateAvatarPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />

              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </PageTitleProvider>
        </AuthProvider>
      </Router>
      <Toaster />
    </ThemeProvider>
  );
}

export default App;
