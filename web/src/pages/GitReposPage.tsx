/**
 * Git Repositories management page for modules and storage repos.
 */

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { gitRepoApi, nodeConfigApi } from '../api/git';
import type { GitRepository, NodeConfig, GitRepoType, GitAuthType } from '../types';

type TabType = 'modules' | 'storage' | 'configs';

interface ConfirmModalProps {
  isOpen: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmText?: string;
  isDestructive?: boolean;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
  isOpen,
  title,
  message,
  onConfirm,
  onCancel,
  confirmText = 'Confirm',
  isDestructive = false,
}) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <h3 className="text-lg font-semibold mb-2">{title}</h3>
        <p className="text-gray-600 mb-6">{message}</p>
        <div className="flex justify-end gap-3">
          <button
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className={`px-4 py-2 rounded-lg text-white ${
              isDestructive ? 'bg-red-600 hover:bg-red-700' : 'bg-blue-600 hover:bg-blue-700'
            }`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

const GitReposPage: React.FC = () => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<TabType>('storage');
  const [showRepoModal, setShowRepoModal] = useState(false);
  const [editingRepo, setEditingRepo] = useState<GitRepository | null>(null);
  const [selectedStorageRepo, setSelectedStorageRepo] = useState<string>('');
  const [notification, setNotification] = useState<{
    show: boolean;
    type: 'success' | 'error';
    message: string;
  }>({ show: false, type: 'success', message: '' });
  const [confirmModal, setConfirmModal] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
  }>({ isOpen: false, title: '', message: '', onConfirm: () => {} });

  const showNotification = (type: 'success' | 'error', message: string) => {
    setNotification({ show: true, type, message });
    setTimeout(() => setNotification({ show: false, type: 'success', message: '' }), 5000);
  };

  // Queries
  const { data: reposData, isLoading: reposLoading } = useQuery({
    queryKey: ['git-repos'],
    queryFn: () => gitRepoApi.list(1, 100),
  });

  const { data: nodeConfigsData, isLoading: nodeConfigsLoading } = useQuery({
    queryKey: ['node-configs', selectedStorageRepo],
    queryFn: () => nodeConfigApi.list(selectedStorageRepo, 1, 100),
    enabled: !!selectedStorageRepo && activeTab === 'configs',
  });

  // Mutations
  const createRepoMutation = useMutation({
    mutationFn: gitRepoApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['git-repos'] });
      setShowRepoModal(false);
    },
  });

  const updateRepoMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof gitRepoApi.update>[1] }) =>
      gitRepoApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['git-repos'] });
      setShowRepoModal(false);
      setEditingRepo(null);
    },
  });

  const deleteRepoMutation = useMutation({
    mutationFn: gitRepoApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['git-repos'] });
    },
  });

  const testConnectionMutation = useMutation({
    mutationFn: gitRepoApi.testConnection,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['git-repos'] });
      showNotification('success', 'Connection successful!');
    },
    onError: (error: Error) => {
      showNotification('error', `Connection failed: ${error.message}`);
    },
  });

  const testConnectionDirectMutation = useMutation({
    mutationFn: gitRepoApi.testConnectionDirect,
    onSuccess: () => {
      showNotification('success', 'Connection successful!');
    },
    onError: (error: Error) => {
      showNotification('error', `Connection failed: ${error.message}`);
    },
  });

  const repos = reposData?.repositories || [];
  const moduleRepos = repos.filter((r) => r.type === 'modules');
  const storageRepos = repos.filter((r) => r.type === 'storage');
  const nodeConfigs = nodeConfigsData?.node_configs || [];

  const [selectedAuthType, setSelectedAuthType] = useState<GitAuthType>('none');

  const handleRepoSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    const repoType = (activeTab === 'modules' ? 'modules' : 'storage') as GitRepoType;
    const authType = formData.get('auth_type') as GitAuthType || 'none';

    if (editingRepo) {
      updateRepoMutation.mutate({
        id: editingRepo.id,
        data: {
          name: formData.get('name') as string,
          url: formData.get('url') as string,
          branch: formData.get('branch') as string || undefined,
          auth_type: authType,
          username: formData.get('username') as string || undefined,
          token: formData.get('token') as string || undefined,
          base_path: formData.get('base_path') as string || undefined,
          description: formData.get('description') as string || undefined,
          is_default: formData.get('is_default') === 'true',
        },
      });
    } else {
      createRepoMutation.mutate({
        name: formData.get('name') as string,
        type: repoType,
        url: formData.get('url') as string,
        branch: formData.get('branch') as string || undefined,
        auth_type: authType,
        username: formData.get('username') as string || undefined,
        token: formData.get('token') as string || undefined,
        base_path: formData.get('base_path') as string || undefined,
        description: formData.get('description') as string || undefined,
        is_default: formData.get('is_default') === 'true',
      });
    }
  };

  const handleTestConnection = (e: React.MouseEvent) => {
    e.preventDefault();
    const form = document.querySelector('form') as HTMLFormElement;
    const formData = new FormData(form);
    
    testConnectionDirectMutation.mutate({
      url: formData.get('url') as string,
      branch: formData.get('branch') as string || undefined,
      auth_type: formData.get('auth_type') as GitAuthType || 'none',
      username: formData.get('username') as string || undefined,
      token: formData.get('token') as string || undefined,
    });
  };

  const handleDeleteRepo = (repo: GitRepository) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Repository',
      message: `Are you sure you want to delete "${repo.name}"? This action cannot be undone.`,
      onConfirm: () => {
        deleteRepoMutation.mutate(repo.id);
        setConfirmModal({ ...confirmModal, isOpen: false });
      },
    });
  };

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      pending: 'bg-yellow-100 text-yellow-800',
      approved: 'bg-blue-100 text-blue-800',
      provisioning: 'bg-purple-100 text-purple-800',
      active: 'bg-green-100 text-green-800',
      failed: 'bg-red-100 text-red-800',
      destroying: 'bg-orange-100 text-orange-800',
      destroyed: 'bg-gray-100 text-gray-800',
    };
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${styles[status] || styles.pending}`}>
        {status}
      </span>
    );
  };

  const renderRepoTable = (repoList: GitRepository[]) => (
    <table className="min-w-full divide-y divide-gray-200">
      <thead className="bg-gray-50">
        <tr>
          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">URL</th>
          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Branch</th>
          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
          <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
        </tr>
      </thead>
      <tbody className="bg-white divide-y divide-gray-200">
        {repoList.map((repo) => (
          <tr key={repo.id}>
            <td className="px-6 py-4 whitespace-nowrap">
              <div className="font-medium text-gray-900">
                {repo.name}
                {repo.is_default && (
                  <span className="ml-2 px-2 py-0.5 text-xs bg-blue-100 text-blue-800 rounded-full">
                    Default
                  </span>
                )}
              </div>
              <div className="text-sm text-gray-500">{repo.description}</div>
            </td>
            <td className="px-6 py-4 whitespace-nowrap">
              <code className="text-sm bg-gray-100 px-2 py-1 rounded max-w-xs truncate block">
                {repo.url}
              </code>
            </td>
            <td className="px-6 py-4 whitespace-nowrap">
              <span className="text-sm text-gray-600">{repo.branch}</span>
            </td>
            <td className="px-6 py-4 whitespace-nowrap">
              <span
                className={`px-2 py-1 text-xs font-medium rounded-full ${
                  repo.status === 1 ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                }`}
              >
                {repo.status === 1 ? 'Active' : 'Disabled'}
              </span>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button
                onClick={() => testConnectionMutation.mutate(repo.id)}
                className="text-green-600 hover:text-green-900 mr-4"
                disabled={testConnectionMutation.isPending}
              >
                Test
              </button>
              <button
                onClick={() => {
                  setEditingRepo(repo);
                  setSelectedAuthType(repo.auth_type || 'none');
                  setShowRepoModal(true);
                }}
                className="text-blue-600 hover:text-blue-900 mr-4"
              >
                Edit
              </button>
              <button
                onClick={() => handleDeleteRepo(repo)}
                className="text-red-600 hover:text-red-900"
              >
                Delete
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );

  return (
    <div className="p-6">
      {/* Notification */}
      {notification.show && (
        <div
          className={`fixed top-4 right-4 z-50 px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 ${
            notification.type === 'success'
              ? 'bg-green-50 border border-green-200 text-green-800'
              : 'bg-red-50 border border-red-200 text-red-800'
          }`}
        >
          {notification.type === 'success' ? (
            <svg className="w-5 h-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          ) : (
            <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          )}
          <span>{notification.message}</span>
          <button
            onClick={() => setNotification({ ...notification, show: false })}
            className="ml-2 text-gray-400 hover:text-gray-600"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}

      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Git Repositories</h1>
        <p className="text-gray-600 mt-1">
          Manage Terraform modules and node configuration storage repositories
        </p>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'storage' as TabType, label: 'Storage Repos', count: storageRepos.length },
            { id: 'modules' as TabType, label: 'Module Repos', count: moduleRepos.length },
            { id: 'configs' as TabType, label: 'Node Configs', count: nodeConfigs.length },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.label}
              <span className="ml-2 py-0.5 px-2.5 rounded-full text-xs bg-gray-100 text-gray-600">
                {tab.count}
              </span>
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="bg-white rounded-lg shadow">
        {(activeTab === 'storage' || activeTab === 'modules') && (
          <>
            <div className="p-4 border-b flex justify-between items-center">
              <h2 className="text-lg font-semibold">
                {activeTab === 'storage' ? 'Storage Repositories' : 'Module Repositories'}
              </h2>
              <button
                onClick={() => {
                  setEditingRepo(null);
                  setSelectedAuthType('none');
                  setShowRepoModal(true);
                }}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Add Repository
              </button>
            </div>
            <div className="p-4">
              {reposLoading ? (
                <div className="text-center py-8 text-gray-500">Loading...</div>
              ) : (activeTab === 'storage' ? storageRepos : moduleRepos).length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  No repositories configured. Add a repository to get started.
                </div>
              ) : (
                renderRepoTable(activeTab === 'storage' ? storageRepos : moduleRepos)
              )}
            </div>
          </>
        )}

        {activeTab === 'configs' && (
          <>
            <div className="p-4 border-b">
              <div className="flex items-center gap-4">
                <label className="text-sm font-medium text-gray-700">Storage Repository:</label>
                <select
                  value={selectedStorageRepo}
                  onChange={(e) => setSelectedStorageRepo(e.target.value)}
                  className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select a repository</option>
                  {storageRepos.map((repo) => (
                    <option key={repo.id} value={repo.id}>
                      {repo.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div className="p-4">
              {!selectedStorageRepo ? (
                <div className="text-center py-8 text-gray-500">
                  Select a storage repository to view node configurations.
                </div>
              ) : nodeConfigsLoading ? (
                <div className="text-center py-8 text-gray-500">Loading...</div>
              ) : nodeConfigs.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  No node configurations found in this repository.
                </div>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Name
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Path
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Commit
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Created
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {nodeConfigs.map((config: NodeConfig) => (
                      <tr key={config.id}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="font-medium text-gray-900">{config.name}</div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                            {config.path}
                          </code>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          {getStatusBadge(config.status)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <code className="text-sm text-gray-600">
                            {config.commit_sha ? config.commit_sha.substring(0, 8) : '-'}
                          </code>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {new Date(config.created_at).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </>
        )}
      </div>

      {/* Repository Modal */}
      {showRepoModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-semibold mb-4">
              {editingRepo ? 'Edit Repository' : 'Add Repository'}
            </h2>
            <form onSubmit={handleRepoSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                <input
                  type="text"
                  name="name"
                  defaultValue={editingRepo?.name || ''}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                  placeholder="My Repository"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Git URL</label>
                <input
                  type="text"
                  name="url"
                  defaultValue={editingRepo?.url || ''}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                  placeholder="https://github.com/org/repo.git"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Branch</label>
                <input
                  type="text"
                  name="branch"
                  defaultValue={editingRepo?.branch || 'main'}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                  placeholder="main"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Authentication Type</label>
                <select
                  name="auth_type"
                  value={selectedAuthType}
                  onChange={(e) => setSelectedAuthType(e.target.value as GitAuthType)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                >
                  <option value="none">None (Public Repository)</option>
                  <option value="token">Access Token / PAT</option>
                  <option value="password">Username & Password</option>
                  <option value="ssh_key">SSH Key</option>
                </select>
              </div>
              {(selectedAuthType === 'token' || selectedAuthType === 'password') && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      {selectedAuthType === 'token' ? 'Username (optional)' : 'Username'}
                    </label>
                    <input
                      type="text"
                      name="username"
                      defaultValue={editingRepo?.username || ''}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      placeholder={selectedAuthType === 'token' ? 'git (optional)' : 'git username'}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      {selectedAuthType === 'token' ? 'Access Token / PAT' : 'Password'}
                    </label>
                    <input
                      type="password"
                      name="token"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                      placeholder="••••••••"
                    />
                    {selectedAuthType === 'token' && (
                      <p className="text-xs text-gray-500 mt-1">
                        GitHub PAT, GitLab Access Token, etc.
                      </p>
                    )}
                  </div>
                </>
              )}
              {selectedAuthType === 'ssh_key' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">SSH Private Key</label>
                  <textarea
                    name="ssh_key"
                    rows={4}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                    placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                  />
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Base Path</label>
                <input
                  type="text"
                  name="base_path"
                  defaultValue={editingRepo?.base_path || '/'}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                  placeholder="/"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Base path within the repository for configurations
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <textarea
                  name="description"
                  defaultValue={editingRepo?.description || ''}
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                  placeholder="Optional description"
                />
              </div>
              <div className="flex items-center">
                <input
                  type="checkbox"
                  name="is_default"
                  value="true"
                  defaultChecked={editingRepo?.is_default || false}
                  className="h-4 w-4 text-blue-600 border-gray-300 rounded"
                />
                <label className="ml-2 text-sm text-gray-700">Set as default repository</label>
              </div>
              <div className="flex justify-between gap-3 pt-4">
                <button
                  type="button"
                  onClick={handleTestConnection}
                  disabled={testConnectionDirectMutation.isPending}
                  className="px-4 py-2 border border-green-500 text-green-600 rounded-lg hover:bg-green-50 disabled:opacity-50"
                >
                  {testConnectionDirectMutation.isPending ? 'Testing...' : 'Test Connection'}
                </button>
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={() => {
                      setShowRepoModal(false);
                      setEditingRepo(null);
                      setSelectedAuthType('none');
                    }}
                    className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={createRepoMutation.isPending || updateRepoMutation.isPending}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                  >
                    {editingRepo ? 'Save Changes' : 'Create'}
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Confirm Modal */}
      <ConfirmModal
        isOpen={confirmModal.isOpen}
        title={confirmModal.title}
        message={confirmModal.message}
        onConfirm={confirmModal.onConfirm}
        onCancel={() => setConfirmModal({ ...confirmModal, isOpen: false })}
        confirmText="Delete"
        isDestructive
      />
    </div>
  );
};

export default GitReposPage;
