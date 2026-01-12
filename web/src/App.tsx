import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import Layout from '@/components/Layout';
import LoginPage from '@/pages/LoginPage';
import DashboardPage from '@/pages/DashboardPage';
import UsersPage from '@/pages/UsersPage';
import RolesPage from '@/pages/RolesPage';
import ResourcesPage from '@/pages/ResourcesPage';
import ResourceRequestsPage from '@/pages/ResourceRequestsPage';
import InfraPage from '@/pages/InfraPage';
import GitReposPage from '@/pages/GitReposPage';
import SettingsPage from '@/pages/SettingsPage';
import SSHKeysPage from '@/pages/SSHKeysPage';
import IPAMPage from '@/pages/IPAMPage';
import VMTemplatesPage from '@/pages/VMTemplatesPage';
import NotFoundPage from '@/pages/NotFoundPage';

/**
 * Protected route wrapper component.
 * Redirects to login if user is not authenticated.
 */
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

/**
 * Main application component.
 */
function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<LoginPage />} />

      {/* Protected routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="users" element={<UsersPage />} />
        <Route path="roles" element={<RolesPage />} />
        <Route path="resources" element={<ResourcesPage />} />
        <Route path="resource-requests" element={<ResourceRequestsPage />} />
        <Route path="infra" element={<InfraPage />} />
        <Route path="git-repos" element={<GitReposPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="ssh-keys" element={<SSHKeysPage />} />
        <Route path="ipam" element={<IPAMPage />} />
        <Route path="vm-templates" element={<VMTemplatesPage />} />
      </Route>

      {/* 404 */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}

export default App;
