/**
 * Infrastructure management page for regions, zones, registries, providers, and modules.
 */

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { regionApi, zoneApi, registryApi, tfProviderApi } from '../api/infra';
import { gitModulesApi } from '../api/git';
import type { Region, Zone, TerraformRegistry, TerraformProvider, GitModule } from '../types';

type TabType = 'regions' | 'zones' | 'registries' | 'providers' | 'modules';

interface ConfirmModalProps {
  isOpen: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmText?: string;
  cancelText?: string;
  isDestructive?: boolean;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
  isOpen,
  title,
  message,
  onConfirm,
  onCancel,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  isDestructive = false,
}) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
          <p className="text-gray-600">{message}</p>
        </div>
        <div className="flex justify-end gap-3 px-6 py-4 bg-gray-50 rounded-b-lg">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            className={`px-4 py-2 text-white rounded-lg ${
              isDestructive
                ? 'bg-red-600 hover:bg-red-700'
                : 'bg-blue-600 hover:bg-blue-700'
            }`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export const InfraPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabType>('regions');
  const [showRegionModal, setShowRegionModal] = useState(false);
  const [showZoneModal, setShowZoneModal] = useState(false);
  const [showRegistryModal, setShowRegistryModal] = useState(false);
  const [showProviderModal, setShowProviderModal] = useState(false);
  const [editingRegion, setEditingRegion] = useState<Region | null>(null);
  const [editingZone, setEditingZone] = useState<Zone | null>(null);
  const [editingRegistry, setEditingRegistry] = useState<TerraformRegistry | null>(null);
  const [editingProvider, setEditingProvider] = useState<TerraformProvider | null>(null);
  const [confirmModal, setConfirmModal] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
    isDestructive?: boolean;
  }>({ isOpen: false, title: '', message: '', onConfirm: () => {} });

  const queryClient = useQueryClient();

  // Queries
  const { data: regionsData, isLoading: regionsLoading } = useQuery({
    queryKey: ['regions'],
    queryFn: () => regionApi.list(1, 100),
  });

  const { data: zonesData, isLoading: zonesLoading } = useQuery({
    queryKey: ['zones'],
    queryFn: () => zoneApi.list(1, 100),
  });

  const { data: registriesData, isLoading: registriesLoading } = useQuery({
    queryKey: ['tf-registries'],
    queryFn: () => registryApi.list(1, 100),
  });

  const { data: tfProvidersData, isLoading: tfProvidersLoading } = useQuery({
    queryKey: ['tf-providers'],
    queryFn: () => tfProviderApi.list(1, 100),
  });

  const { data: modulesData, isLoading: modulesLoading } = useQuery({
    queryKey: ['git-modules'],
    queryFn: () => gitModulesApi.list(),
  });

  // Sync modules mutation
  const syncModulesMutation = useMutation({
    mutationFn: gitModulesApi.sync,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['git-modules'] });
    },
  });

  // Mutations
  const createRegionMutation = useMutation({
    mutationFn: regionApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['regions'] });
      setShowRegionModal(false);
    },
  });

  const updateRegionMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof regionApi.update>[1] }) =>
      regionApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['regions'] });
      setEditingRegion(null);
      setShowRegionModal(false);
    },
  });

  const deleteRegionMutation = useMutation({
    mutationFn: regionApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['regions'] });
    },
  });

  const createZoneMutation = useMutation({
    mutationFn: zoneApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zones'] });
      setShowZoneModal(false);
    },
  });

  const updateZoneMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof zoneApi.update>[1] }) =>
      zoneApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zones'] });
      setEditingZone(null);
      setShowZoneModal(false);
    },
  });

  const deleteZoneMutation = useMutation({
    mutationFn: zoneApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zones'] });
    },
  });

  // Registry mutations
  const createRegistryMutation = useMutation({
    mutationFn: registryApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tf-registries'] });
      setShowRegistryModal(false);
    },
  });

  const updateRegistryMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof registryApi.update>[1] }) =>
      registryApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tf-registries'] });
      setEditingRegistry(null);
      setShowRegistryModal(false);
    },
  });

  const deleteRegistryMutation = useMutation({
    mutationFn: registryApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tf-registries'] });
    },
  });

  // Provider mutations
  const createProviderMutation = useMutation({
    mutationFn: tfProviderApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tf-providers'] });
      setShowProviderModal(false);
    },
  });

  const updateProviderMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Parameters<typeof tfProviderApi.update>[1] }) =>
      tfProviderApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tf-providers'] });
      setEditingProvider(null);
      setShowProviderModal(false);
    },
  });

  const deleteProviderMutation = useMutation({
    mutationFn: tfProviderApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tf-providers'] });
    },
  });

  const handleDeleteRegion = (region: Region) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Region',
      message: `Are you sure you want to delete region "${region.display_name}"? This action cannot be undone.`,
      isDestructive: true,
      onConfirm: () => {
        deleteRegionMutation.mutate(region.id);
        setConfirmModal((prev) => ({ ...prev, isOpen: false }));
      },
    });
  };

  const handleDeleteZone = (zone: Zone) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Zone',
      message: `Are you sure you want to delete zone "${zone.display_name}"? This action cannot be undone.`,
      isDestructive: true,
      onConfirm: () => {
        deleteZoneMutation.mutate(zone.id);
        setConfirmModal((prev) => ({ ...prev, isOpen: false }));
      },
    });
  };

  const handleDeleteRegistry = (registry: TerraformRegistry) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Registry',
      message: `Are you sure you want to delete registry "${registry.name}"? This action cannot be undone.`,
      isDestructive: true,
      onConfirm: () => {
        deleteRegistryMutation.mutate(registry.id);
        setConfirmModal((prev) => ({ ...prev, isOpen: false }));
      },
    });
  };

  const handleDeleteProvider = (provider: TerraformProvider) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Provider',
      message: `Are you sure you want to delete provider "${provider.name}"? This action cannot be undone.`,
      isDestructive: true,
      onConfirm: () => {
        deleteProviderMutation.mutate(provider.id);
        setConfirmModal((prev) => ({ ...prev, isOpen: false }));
      },
    });
  };

  const handleRegionSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    if (editingRegion) {
      updateRegionMutation.mutate({
        id: editingRegion.id,
        data: {
          name: formData.get('name') as string,
          display_name: formData.get('display_name') as string,
          description: formData.get('description') as string,
        },
      });
    } else {
      createRegionMutation.mutate({
        name: formData.get('name') as string,
        code: formData.get('code') as string,
        display_name: formData.get('display_name') as string,
        description: formData.get('description') as string,
      });
    }
  };

  const handleZoneSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    if (editingZone) {
      updateZoneMutation.mutate({
        id: editingZone.id,
        data: {
          name: formData.get('name') as string,
          display_name: formData.get('display_name') as string,
          description: formData.get('description') as string,
          is_default: formData.get('is_default') === 'true',
        },
      });
    } else {
      createZoneMutation.mutate({
        name: formData.get('name') as string,
        code: formData.get('code') as string,
        display_name: formData.get('display_name') as string,
        description: formData.get('description') as string,
        region_id: formData.get('region_id') as string,
        is_default: formData.get('is_default') === 'true',
      });
    }
  };

  const handleRegistrySubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    if (editingRegistry) {
      updateRegistryMutation.mutate({
        id: editingRegistry.id,
        data: {
          name: formData.get('name') as string,
          endpoint: formData.get('endpoint') as string,
          username: formData.get('username') as string || undefined,
          token: formData.get('token') as string || undefined,
          description: formData.get('description') as string,
          is_default: formData.get('is_default') === 'true',
        },
      });
    } else {
      createRegistryMutation.mutate({
        name: formData.get('name') as string,
        endpoint: formData.get('endpoint') as string,
        username: formData.get('username') as string || undefined,
        token: formData.get('token') as string || undefined,
        description: formData.get('description') as string,
        is_default: formData.get('is_default') === 'true',
      });
    }
  };

  const handleProviderSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    if (editingProvider) {
      updateProviderMutation.mutate({
        id: editingProvider.id,
        data: {
          name: formData.get('name') as string,
          namespace: formData.get('namespace') as string,
          source: formData.get('source') as string,
          version: formData.get('version') as string,
          description: formData.get('description') as string,
        },
      });
    } else {
      createProviderMutation.mutate({
        name: formData.get('name') as string,
        namespace: formData.get('namespace') as string,
        source: formData.get('source') as string,
        version: formData.get('version') as string,
        registry_id: formData.get('registry_id') as string,
        description: formData.get('description') as string,
      });
    }
  };

  const regions = regionsData?.regions || [];
  const zones = zonesData?.zones || [];
  const registries = registriesData?.registries || [];
  const tfProviders = tfProvidersData?.providers || [];
  const modules: GitModule[] = modulesData?.modules || [];

  const getTabAddLabel = () => {
    switch (activeTab) {
      case 'regions': return 'Region';
      case 'zones': return 'Zone';
      case 'registries': return 'Registry';
      case 'providers': return 'Provider';
      case 'modules': return null; // Modules come from git, no manual add
    }
  };

  const handleAddClick = () => {
    switch (activeTab) {
      case 'regions':
        setEditingRegion(null);
        setShowRegionModal(true);
        break;
      case 'zones':
        setEditingZone(null);
        setShowZoneModal(true);
        break;
      case 'registries':
        setEditingRegistry(null);
        setShowRegistryModal(true);
        break;
      case 'providers':
        setEditingProvider(null);
        setShowProviderModal(true);
        break;
      case 'modules':
        // Modules are synced from git, trigger sync instead
        syncModulesMutation.mutate();
        break;
    }
  };

  const tabs = [
    { id: 'regions' as const, name: 'Regions', count: regions.length },
    { id: 'zones' as const, name: 'Zones', count: zones.length },
    { id: 'registries' as const, name: 'Registries', count: registries.length },
    { id: 'providers' as const, name: 'Providers', count: tfProviders.length },
    { id: 'modules' as const, name: 'Modules', count: modules.length },
  ];

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Infrastructure</h1>
          <p className="text-gray-600">Manage regions, zones, and Terraform configurations</p>
        </div>
        {activeTab === 'modules' ? (
          <button
            onClick={() => syncModulesMutation.mutate()}
            disabled={syncModulesMutation.isPending}
            className="btn btn-primary"
          >
            {syncModulesMutation.isPending ? 'Syncing...' : 'Sync Modules'}
          </button>
        ) : getTabAddLabel() && (
          <button
            onClick={handleAddClick}
            className="btn btn-primary"
          >
            Add {getTabAddLabel()}
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.name}
              <span
                className={`ml-2 py-0.5 px-2.5 rounded-full text-xs ${
                  activeTab === tab.id
                    ? 'bg-blue-100 text-blue-600'
                    : 'bg-gray-100 text-gray-900'
                }`}
              >
                {tab.count}
              </span>
            </button>
          ))}
        </nav>
      </div>

      {/* Regions Tab */}
      {activeTab === 'regions' && (
        <div className="card">
          {regionsLoading ? (
            <div className="text-center py-8">Loading...</div>
          ) : regions.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No regions configured. Add a region to get started.
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Code
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Zones
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {regions.map((region) => (
                  <tr key={region.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="font-medium text-gray-900">{region.display_name}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-sm bg-gray-100 px-2 py-1 rounded">{region.code}</code>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-500 max-w-xs truncate">
                        {region.description || '-'}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">{region.zones?.length || 0}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded-full ${
                          region.status === 1
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {region.status === 1 ? 'Active' : 'Disabled'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => {
                          setEditingRegion(region);
                          setShowRegionModal(true);
                        }}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteRegion(region)}
                        className="text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Zones Tab */}
      {activeTab === 'zones' && (
        <div className="card">
          {zonesLoading ? (
            <div className="text-center py-8">Loading...</div>
          ) : zones.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No zones configured. Add a zone to get started.
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Code
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Region
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {zones.map((zone) => (
                  <tr key={zone.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="font-medium text-gray-900">
                        {zone.display_name}
                        {zone.is_default && (
                          <span className="ml-2 px-2 py-0.5 text-xs bg-blue-100 text-blue-800 rounded-full">
                            Default
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-sm bg-gray-100 px-2 py-1 rounded">{zone.code}</code>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">
                        {zone.region?.display_name || '-'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded-full ${
                          zone.status === 1
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {zone.status === 1 ? 'Active' : 'Disabled'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => {
                          setEditingZone(zone);
                          setShowZoneModal(true);
                        }}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteZone(zone)}
                        className="text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Registries Tab */}
      {activeTab === 'registries' && (
        <div className="card">
          {registriesLoading ? (
            <div className="text-center py-8">Loading...</div>
          ) : registries.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No registries configured. Add a Terraform registry to get started.
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Endpoint
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {registries.map((registry) => (
                  <tr key={registry.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="font-medium text-gray-900">
                        {registry.name}
                        {registry.is_default && (
                          <span className="ml-2 px-2 py-0.5 text-xs bg-blue-100 text-blue-800 rounded-full">
                            Default
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-sm bg-gray-100 px-2 py-1 rounded">{registry.endpoint}</code>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-500 max-w-xs truncate">
                        {registry.description || '-'}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded-full ${
                          registry.status === 1
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {registry.status === 1 ? 'Active' : 'Disabled'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => {
                          setEditingRegistry(registry);
                          setShowRegistryModal(true);
                        }}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteRegistry(registry)}
                        className="text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Providers Tab */}
      {activeTab === 'providers' && (
        <div className="card">
          {tfProvidersLoading ? (
            <div className="text-center py-8">Loading...</div>
          ) : tfProviders.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No providers configured. Add a Terraform provider (e.g., pve, aws, gcp) to get started.
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Namespace
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Source
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Registry
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {tfProviders.map((provider) => (
                  <tr key={provider.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="font-medium text-gray-900">{provider.name}</div>
                      {provider.version && (
                        <div className="text-xs text-gray-500">v{provider.version}</div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-sm bg-gray-100 px-2 py-1 rounded">{provider.namespace || '-'}</code>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">{provider.source || '-'}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">{provider.registry?.name || '-'}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded-full ${
                          provider.status === 1
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {provider.status === 1 ? 'Active' : 'Disabled'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => {
                          setEditingProvider(provider);
                          setShowProviderModal(true);
                        }}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteProvider(provider)}
                        className="text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Modules Tab */}
      {activeTab === 'modules' && (
        <div className="card">
          <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <p className="text-sm text-blue-800">
              Modules are automatically discovered from your Git Ops modules repository. 
              Click "Sync Modules" to refresh the list from the repository.
            </p>
          </div>
          {modulesLoading || syncModulesMutation.isPending ? (
            <div className="text-center py-8">
              {syncModulesMutation.isPending ? 'Syncing modules from repository...' : 'Loading...'}
            </div>
          ) : modules.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No modules found. Make sure you have configured a modules repository in Git Ops settings, 
              then click "Sync Modules" to discover available Terraform modules.
            </div>
          ) : (
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Path
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Variables
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Outputs
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {modules.map((mod, index) => (
                  <tr key={`${mod.path}-${index}`}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="font-medium text-gray-900">{mod.name}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-sm bg-gray-100 px-2 py-1 rounded">{mod.path}</code>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-500 max-w-xs truncate">
                        {mod.description || '-'}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {mod.variables && mod.variables.length > 0 ? (
                          mod.variables.slice(0, 3).map((v) => (
                            <span key={v} className="px-2 py-0.5 text-xs bg-purple-100 text-purple-700 rounded">
                              {v}
                            </span>
                          ))
                        ) : (
                          <span className="text-gray-400 text-sm">-</span>
                        )}
                        {mod.variables && mod.variables.length > 3 && (
                          <span className="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">
                            +{mod.variables.length - 3} more
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {mod.outputs && mod.outputs.length > 0 ? (
                          mod.outputs.slice(0, 3).map((o) => (
                            <span key={o} className="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded">
                              {o}
                            </span>
                          ))
                        ) : (
                          <span className="text-gray-400 text-sm">-</span>
                        )}
                        {mod.outputs && mod.outputs.length > 3 && (
                          <span className="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">
                            +{mod.outputs.length - 3} more
                          </span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Region Modal */}
      {showRegionModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4">
            <div className="p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                {editingRegion ? 'Edit Region' : 'Add Region'}
              </h3>
              <form onSubmit={handleRegionSubmit}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                    <input
                      type="text"
                      name="name"
                      defaultValue={editingRegion?.name || ''}
                      required
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="cn-north"
                    />
                  </div>
                  {!editingRegion && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Code</label>
                      <input
                        type="text"
                        name="code"
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        placeholder="cn-north"
                      />
                      <p className="text-xs text-gray-500 mt-1">Unique identifier, cannot be changed later</p>
                    </div>
                  )}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Display Name</label>
                    <input
                      type="text"
                      name="display_name"
                      defaultValue={editingRegion?.display_name || ''}
                      required
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="China North"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea
                      name="description"
                      defaultValue={editingRegion?.description || ''}
                      rows={3}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Optional description"
                    />
                  </div>
                </div>
                <div className="flex justify-end gap-3 mt-6">
                  <button
                    type="button"
                    onClick={() => {
                      setShowRegionModal(false);
                      setEditingRegion(null);
                    }}
                    className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
                  >
                    {editingRegion ? 'Update' : 'Create'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Zone Modal */}
      {showZoneModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] flex flex-col">
            <div className="p-6 overflow-y-auto flex-1">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                {editingZone ? 'Edit Zone' : 'Add Zone'}
              </h3>
              <form onSubmit={handleZoneSubmit} id="zone-form">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                    <input
                      type="text"
                      name="name"
                      defaultValue={editingZone?.name || ''}
                      required
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="zone-a"
                    />
                  </div>
                  {!editingZone && (
                    <>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Code</label>
                        <input
                          type="text"
                          name="code"
                          required
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          placeholder="zone-a"
                        />
                        <p className="text-xs text-gray-500 mt-1">Unique identifier, cannot be changed later</p>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Region</label>
                        <select
                          name="region_id"
                          required
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        >
                          <option value="">Select Region</option>
                          {regions.map((region) => (
                            <option key={region.id} value={region.id}>
                              {region.display_name}
                            </option>
                          ))}
                        </select>
                      </div>
                    </>
                  )}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Display Name</label>
                    <input
                      type="text"
                      name="display_name"
                      defaultValue={editingZone?.display_name || ''}
                      required
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Availability Zone A"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea
                      name="description"
                      defaultValue={editingZone?.description || ''}
                      rows={3}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Optional description"
                    />
                  </div>
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      name="is_default"
                      value="true"
                      defaultChecked={editingZone?.is_default || false}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <label className="ml-2 block text-sm text-gray-900">Default zone in region</label>
                  </div>
                </div>
              </form>
            </div>
            <div className="flex justify-end gap-3 px-6 py-4 bg-gray-50 rounded-b-lg border-t">
              <button
                type="button"
                onClick={() => {
                  setShowZoneModal(false);
                  setEditingZone(null);
                }}
                className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                form="zone-form"
                className="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
              >
                {editingZone ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Registry Modal */}
      {showRegistryModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] flex flex-col">
            <div className="p-6 overflow-y-auto flex-1">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                {editingRegistry ? 'Edit Registry' : 'Add Registry'}
              </h3>
              <form onSubmit={handleRegistrySubmit} id="registry-form">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                    <input
                      type="text"
                      name="name"
                      defaultValue={editingRegistry?.name || ''}
                      required
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="My Private Registry"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Endpoint</label>
                    <input
                      type="text"
                      name="endpoint"
                      defaultValue={editingRegistry?.endpoint || ''}
                      required
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="https://registry.example.com"
                    />
                    <p className="text-xs text-gray-500 mt-1">The Terraform registry server URL</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Username (Optional)</label>
                    <input
                      type="text"
                      name="username"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="username"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Token (Optional)</label>
                    <input
                      type="password"
                      name="token"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="••••••••"
                    />
                    <p className="text-xs text-gray-500 mt-1">API token for authentication</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea
                      name="description"
                      defaultValue={editingRegistry?.description || ''}
                      rows={2}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Optional description"
                    />
                  </div>
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      name="is_default"
                      value="true"
                      defaultChecked={editingRegistry?.is_default || false}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <label className="ml-2 block text-sm text-gray-900">Set as default registry</label>
                  </div>
                </div>
              </form>
            </div>
            <div className="flex justify-end gap-3 px-6 py-4 bg-gray-50 rounded-b-lg border-t">
              <button
                type="button"
                onClick={() => {
                  setShowRegistryModal(false);
                  setEditingRegistry(null);
                }}
                className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                form="registry-form"
                className="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
              >
                {editingRegistry ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Provider Modal */}
      {showProviderModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] flex flex-col">
            <div className="p-6 overflow-y-auto flex-1">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                {editingProvider ? 'Edit Provider' : 'Add Provider'}
              </h3>
              <form onSubmit={handleProviderSubmit} id="provider-form">
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                    <input
                      type="text"
                      name="name"
                      defaultValue={editingProvider?.name || ''}
                      required
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="proxmox"
                    />
                    <p className="text-xs text-gray-500 mt-1">Provider name (e.g., proxmox, aws, gcp)</p>
                  </div>
                  {!editingProvider && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Registry</label>
                      <select
                        name="registry_id"
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      >
                        <option value="">Select Registry</option>
                        {registries.map((registry) => (
                          <option key={registry.id} value={registry.id}>
                            {registry.name} ({registry.endpoint})
                          </option>
                        ))}
                      </select>
                      <p className="text-xs text-gray-500 mt-1">The registry where this provider is hosted</p>
                    </div>
                  )}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Namespace</label>
                    <input
                      type="text"
                      name="namespace"
                      defaultValue={editingProvider?.namespace || ''}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="bpg"
                    />
                    <p className="text-xs text-gray-500 mt-1">Provider namespace (e.g., hashicorp, bpg)</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Source</label>
                    <input
                      type="text"
                      name="source"
                      defaultValue={editingProvider?.source || ''}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="bpg/proxmox"
                    />
                    <p className="text-xs text-gray-500 mt-1">Full provider source path</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Version</label>
                    <input
                      type="text"
                      name="version"
                      defaultValue={editingProvider?.version || ''}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="~> 0.38.0"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                    <textarea
                      name="description"
                      defaultValue={editingProvider?.description || ''}
                      rows={2}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="Optional description"
                    />
                  </div>
                </div>
              </form>
            </div>
            <div className="flex justify-end gap-3 px-6 py-4 bg-gray-50 rounded-b-lg border-t">
              <button
                type="button"
                onClick={() => {
                  setShowProviderModal(false);
                  setEditingProvider(null);
                }}
                className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                form="provider-form"
                className="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
              >
                {editingProvider ? 'Update' : 'Create'}
              </button>
            </div>
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
        isDestructive={confirmModal.isDestructive}
      />
    </div>
  );
};

export default InfraPage;
