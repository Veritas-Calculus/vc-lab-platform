import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { providerApi, credentialApi } from '@/api/settings';
import { useIsAdmin } from '@/stores/authStore';
import type {
  ProviderConfig,
  Credential,
  CreateProviderReq,
  CreateCredentialReq,
  ProviderType,
  TestProviderConnectionReq,
  TestCredentialConnectionReq,
} from '@/types';

type TabType = 'providers' | 'credentials';

const PROVIDER_TYPES: { value: ProviderType; label: string }[] = [
  { value: 'pve', label: 'Proxmox VE' },
  { value: 'vmware', label: 'VMware vSphere' },
  { value: 'openstack', label: 'OpenStack' },
  { value: 'aws', label: 'AWS' },
  { value: 'aliyun', label: 'Aliyun' },
  { value: 'gcp', label: 'Google Cloud' },
  { value: 'azure', label: 'Azure' },
];

/**
 * Settings page for managing providers and credentials.
 */
export default function SettingsPage() {
  const queryClient = useQueryClient();
  const isAdmin = useIsAdmin();
  const [activeTab, setActiveTab] = useState<TabType>('providers');
  const [showProviderModal, setShowProviderModal] = useState(false);
  const [showCredentialModal, setShowCredentialModal] = useState(false);
  const [selectedProviderId, setSelectedProviderId] = useState<string>('');

  // Fetch providers
  const { data: providersData, isLoading: providersLoading } = useQuery({
    queryKey: ['providers'],
    queryFn: () => providerApi.list(),
  });

  // Fetch credentials
  const { data: credentialsData, isLoading: credentialsLoading } = useQuery({
    queryKey: ['credentials'],
    queryFn: () => credentialApi.list(),
  });

  // Provider mutations
  const createProviderMutation = useMutation({
    mutationFn: (data: CreateProviderReq) => providerApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['providers'] });
      setShowProviderModal(false);
    },
  });

  const testProviderConnectionMutation = useMutation({
    mutationFn: (data: TestProviderConnectionReq) => providerApi.testConnection(data),
  });

  const deleteProviderMutation = useMutation({
    mutationFn: (id: string) => providerApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['providers'] });
    },
  });

  // Credential mutations
  const createCredentialMutation = useMutation({
    mutationFn: (data: CreateCredentialReq) => credentialApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials'] });
      setShowCredentialModal(false);
      setSelectedProviderId('');
    },
  });

  const testCredentialConnectionMutation = useMutation({
    mutationFn: (data: TestCredentialConnectionReq) => credentialApi.testConnection(data),
  });

  const deleteCredentialMutation = useMutation({
    mutationFn: (id: string) => credentialApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials'] });
    },
  });

  const handleCreateProvider = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    createProviderMutation.mutate({
      name: formData.get('name') as string,
      type: formData.get('type') as ProviderType,
      endpoint: formData.get('endpoint') as string,
      description: formData.get('description') as string,
      config: formData.get('config') as string,
      is_default: formData.get('is_default') === 'true',
      credential_id: (formData.get('credential_id') as string) || undefined,
    });
  };

  const handleTestProviderConnection = (form: HTMLFormElement) => {
    const formData = new FormData(form);
    testProviderConnectionMutation.mutate({
      type: formData.get('type') as ProviderType,
      endpoint: formData.get('endpoint') as string,
      credential_id: (formData.get('credential_id') as string) || undefined,
      config: (formData.get('config') as string) || undefined,
    });
  };

  const handleCreateCredential = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const providerId = formData.get('provider_id') as string;
    const selectedProvider = providersData?.providers?.find(p => p.id === providerId);
    createCredentialMutation.mutate({
      name: formData.get('name') as string,
      type: formData.get('type') as ProviderType,
      provider_id: providerId || undefined,
      access_key: formData.get('access_key') as string,
      secret_key: formData.get('secret_key') as string,
      token: formData.get('token') as string,
      description: formData.get('description') as string,
      endpoint: selectedProvider?.endpoint || (formData.get('endpoint') as string),
    });
  };

  const handleTestCredentialConnection = (form: HTMLFormElement) => {
    const formData = new FormData(form);
    const providerId = formData.get('provider_id') as string;
    const selectedProvider = providersData?.providers?.find(p => p.id === providerId);
    testCredentialConnectionMutation.mutate({
      type: formData.get('type') as ProviderType,
      endpoint: selectedProvider?.endpoint || (formData.get('endpoint') as string),
      access_key: (formData.get('access_key') as string) || undefined,
      secret_key: (formData.get('secret_key') as string) || undefined,
      token: (formData.get('token') as string) || undefined,
    });
  };

  const handleDeleteProvider = (provider: ProviderConfig) => {
    if (confirm(`Delete provider "${provider.name}"?`)) {
      deleteProviderMutation.mutate(provider.id);
    }
  };

  const handleDeleteCredential = (credential: Credential) => {
    if (confirm(`Delete credential "${credential.name}"?`)) {
      deleteCredentialMutation.mutate(credential.id);
    }
  };

  const getProviderTypeLabel = (type: string): string => {
    return PROVIDER_TYPES.find(p => p.value === type)?.label || type;
  };

  const getStatusBadge = (status: number) => {
    if (status === 1) {
      return <span className="inline-flex px-2 py-1 text-xs font-medium rounded bg-green-100 text-green-800">Active</span>;
    }
    return <span className="inline-flex px-2 py-1 text-xs font-medium rounded bg-gray-100 text-gray-800">Disabled</span>;
  };

  if (!isAdmin) {
    return (
      <div className="p-6">
        <div className="card p-8 text-center">
          <p className="text-gray-500">You do not have permission to access settings.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-600">Manage infrastructure providers and credentials</p>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('providers')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'providers'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Providers
          </button>
          <button
            onClick={() => setActiveTab('credentials')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'credentials'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Credentials
          </button>
        </nav>
      </div>

      {/* Providers Tab */}
      {activeTab === 'providers' && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-medium text-gray-900">Infrastructure Providers</h2>
            <button onClick={() => setShowProviderModal(true)} className="btn btn-primary">
              Add Provider
            </button>
          </div>

          <div className="card overflow-hidden">
            {providersLoading ? (
              <div className="p-8 text-center">
                <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
              </div>
            ) : (
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Endpoint</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Credential</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {providersData?.providers?.map((provider) => (
                    <tr key={provider.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <span className="text-sm font-medium text-gray-900">{provider.name}</span>
                          {provider.is_default && (
                            <span className="ml-2 inline-flex px-2 py-0.5 text-xs font-medium rounded bg-blue-100 text-blue-800">Default</span>
                          )}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {getProviderTypeLabel(provider.type)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 max-w-xs truncate">
                        {provider.endpoint}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {provider.credential?.name || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(provider.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button
                          onClick={() => handleDeleteProvider(provider)}
                          className="text-red-600 hover:text-red-900"
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                  {(!providersData?.providers || providersData.providers.length === 0) && (
                    <tr>
                      <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                        No providers configured. Add a provider to get started.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Credentials Tab */}
      {activeTab === 'credentials' && (
        <div>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-medium text-gray-900">Cloud Credentials</h2>
            <button onClick={() => setShowCredentialModal(true)} className="btn btn-primary">
              Add Credential
            </button>
          </div>

          <div className="card overflow-hidden">
            {credentialsLoading ? (
              <div className="p-8 text-center">
                <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
              </div>
            ) : (
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created By</th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {credentialsData?.credentials?.map((credential) => (
                    <tr key={credential.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {credential.name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {getProviderTypeLabel(credential.type)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {credential.provider?.name || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(credential.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {credential.created_by?.username || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button
                          onClick={() => handleDeleteCredential(credential)}
                          className="text-red-600 hover:text-red-900"
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                  {(!credentialsData?.credentials || credentialsData.credentials.length === 0) && (
                    <tr>
                      <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                        No credentials configured. Add credentials to authenticate with providers.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Add Provider Modal */}
      {showProviderModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Add Provider</h3>
              <form id="providerForm" onSubmit={handleCreateProvider}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                    <input name="name" type="text" className="input" required />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Type *</label>
                    <select name="type" className="input" required>
                      {PROVIDER_TYPES.map((type) => (
                        <option key={type.value} value={type.value}>{type.label}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Endpoint URL *</label>
                    <input name="endpoint" type="url" className="input" placeholder="https://api.example.com" required />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Credential</label>
                    <select name="credential_id" className="input">
                      <option value="">-- Select Credential --</option>
                      {credentialsData?.credentials?.map((credential) => (
                        <option key={credential.id} value={credential.id}>{credential.name} ({getProviderTypeLabel(credential.type)})</option>
                      ))}
                    </select>
                    <p className="mt-1 text-sm text-gray-500">Select a credential to authenticate with this provider</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea name="description" className="input" rows={2} />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Configuration (JSON)</label>
                    <textarea name="config" className="input font-mono text-sm" rows={4} placeholder='{"node": "pve1"}' />
                  </div>
                  <div className="flex items-center">
                    <input name="is_default" type="checkbox" value="true" className="h-4 w-4 text-primary-600 rounded" />
                    <label className="ml-2 block text-sm text-gray-700">Set as default provider</label>
                  </div>
                </div>
                {/* Test Connection Result */}
                {testProviderConnectionMutation.isSuccess && (
                  <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-md">
                    <p className="text-sm text-green-700">✓ Connection successful</p>
                  </div>
                )}
                {testProviderConnectionMutation.isError && (
                  <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
                    <p className="text-sm text-red-700">✗ Connection failed: {(testProviderConnectionMutation.error as Error)?.message || 'Unknown error'}</p>
                  </div>
                )}
                <div className="mt-6 flex justify-between">
                  <button
                    type="button"
                    onClick={() => {
                      const form = document.getElementById('providerForm') as HTMLFormElement;
                      if (form) handleTestProviderConnection(form);
                    }}
                    className="btn btn-secondary"
                    disabled={testProviderConnectionMutation.isPending}
                  >
                    {testProviderConnectionMutation.isPending ? 'Testing...' : 'Test Connection'}
                  </button>
                  <div className="flex gap-3">
                    <button type="button" onClick={() => setShowProviderModal(false)} className="btn btn-secondary">
                      Cancel
                    </button>
                    <button type="submit" className="btn btn-primary" disabled={createProviderMutation.isPending}>
                      {createProviderMutation.isPending ? 'Creating...' : 'Create'}
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Add Credential Modal */}
      {showCredentialModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Add Credential</h3>
              <form id="credentialForm" onSubmit={handleCreateCredential}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                    <input name="name" type="text" className="input" required />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Type *</label>
                    <select name="type" className="input" required>
                      {PROVIDER_TYPES.map((type) => (
                        <option key={type.value} value={type.value}>{type.label}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Provider</label>
                    <select name="provider_id" className="input" value={selectedProviderId} onChange={(e) => setSelectedProviderId(e.target.value)}>
                      <option value="">-- No Provider (use custom endpoint) --</option>
                      {providersData?.providers?.map((provider) => (
                        <option key={provider.id} value={provider.id}>{provider.name} ({provider.endpoint})</option>
                      ))}
                    </select>
                    <p className="mt-1 text-sm text-gray-500">
                      {selectedProviderId ? 'Using endpoint from selected provider' : 'Select a provider or enter a custom endpoint below'}
                    </p>
                  </div>
                  {!selectedProviderId && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Endpoint URL *</label>
                      <input name="endpoint" type="url" className="input" placeholder="https://api.example.com" required />
                    </div>
                  )}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Access Key / Username</label>
                    <input name="access_key" type="text" className="input" placeholder="AKIAIOSFODNN7EXAMPLE" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Secret Key / Password</label>
                    <input name="secret_key" type="password" className="input" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Token (optional)</label>
                    <textarea name="token" className="input font-mono text-sm" rows={3} placeholder="API Token or Service Account Key" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea name="description" className="input" rows={2} />
                  </div>
                </div>
                {/* Test Connection Result */}
                {testCredentialConnectionMutation.isSuccess && (
                  <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-md">
                    <p className="text-sm text-green-700">✓ Authentication successful</p>
                  </div>
                )}
                {testCredentialConnectionMutation.isError && (
                  <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
                    <p className="text-sm text-red-700">✗ Authentication failed: {(testCredentialConnectionMutation.error as Error)?.message || 'Unknown error'}</p>
                  </div>
                )}
                <div className="mt-6 flex justify-between">
                  <button
                    type="button"
                    onClick={() => {
                      const form = document.getElementById('credentialForm') as HTMLFormElement;
                      if (form) handleTestCredentialConnection(form);
                    }}
                    className="btn btn-secondary"
                    disabled={testCredentialConnectionMutation.isPending}
                  >
                    {testCredentialConnectionMutation.isPending ? 'Testing...' : 'Test Connection'}
                  </button>
                  <div className="flex gap-3">
                    <button type="button" onClick={() => { setShowCredentialModal(false); setSelectedProviderId(''); }} className="btn btn-secondary">
                      Cancel
                    </button>
                    <button type="submit" className="btn btn-primary" disabled={createCredentialMutation.isPending}>
                      {createCredentialMutation.isPending ? 'Creating...' : 'Create'}
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
