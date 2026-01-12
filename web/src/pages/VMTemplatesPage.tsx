import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { vmTemplateApi } from '@/api/templates';
import { zoneApi } from '@/api/infra';
import { providerApi } from '@/api/settings';
import { useIsAdmin } from '@/stores/authStore';
import type { VMTemplate, CreateVMTemplateReq, Zone } from '@/types';

const OS_TYPES = [
  { value: 'linux', label: 'Linux' },
  { value: 'windows', label: 'Windows' },
  { value: 'bsd', label: 'BSD' },
  { value: 'other', label: 'Other' },
];

const OS_FAMILIES = [
  { value: 'debian', label: 'Debian/Ubuntu' },
  { value: 'rhel', label: 'RHEL/CentOS' },
  { value: 'alpine', label: 'Alpine' },
  { value: 'arch', label: 'Arch Linux' },
  { value: 'windows-server', label: 'Windows Server' },
  { value: 'windows-desktop', label: 'Windows Desktop' },
  { value: 'other', label: 'Other' },
];

/**
 * VM Templates management page.
 */
export default function VMTemplatesPage() {
  const queryClient = useQueryClient();
  const isAdmin = useIsAdmin();
  const [showModal, setShowModal] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<VMTemplate | null>(null);
  const [filterProvider, setFilterProvider] = useState<string>('');

  // Fetch templates
  const { data, isLoading } = useQuery({
    queryKey: ['vm-templates', filterProvider],
    queryFn: () => vmTemplateApi.list({ provider: filterProvider || undefined }),
  });

  // Fetch providers for dropdown
  const { data: providersData } = useQuery({
    queryKey: ['settings-providers'],
    queryFn: () => providerApi.list(),
  });

  // Fetch zones for dropdown
  const { data: zonesData } = useQuery({
    queryKey: ['zones', 'all'],
    queryFn: () => zoneApi.listAll(),
  });

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateVMTemplateReq) => vmTemplateApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vm-templates'] });
      setShowModal(false);
      setEditingTemplate(null);
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateVMTemplateReq> }) => 
      vmTemplateApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vm-templates'] });
      setShowModal(false);
      setEditingTemplate(null);
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => vmTemplateApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vm-templates'] });
    },
  });

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data: CreateVMTemplateReq = {
      name: formData.get('name') as string,
      template_name: formData.get('template_name') as string,
      provider: formData.get('provider') as string,
      os_type: formData.get('os_type') as string,
      os_family: formData.get('os_family') as string,
      os_version: formData.get('os_version') as string,
      zone_id: formData.get('zone_id') as string || undefined,
      min_cpu: parseInt(formData.get('min_cpu') as string) || 1,
      min_memory_mb: parseInt(formData.get('min_memory_mb') as string) || 512,
      min_disk_gb: parseInt(formData.get('min_disk_gb') as string) || 10,
      default_user: formData.get('default_user') as string,
      cloud_init: formData.get('cloud_init') === 'true',
      description: formData.get('description') as string,
    };

    if (editingTemplate) {
      updateMutation.mutate({ id: editingTemplate.id, data });
    } else {
      createMutation.mutate(data);
    }
  };

  const handleDelete = (template: VMTemplate) => {
    if (confirm(`Delete template "${template.name}"?`)) {
      deleteMutation.mutate(template.id);
    }
  };

  const handleEdit = (template: VMTemplate) => {
    setEditingTemplate(template);
    setShowModal(true);
  };

  const getStatusBadge = (status: number) => {
    if (status === 1) {
      return <span className="inline-flex px-2 py-1 text-xs font-medium rounded bg-green-100 text-green-800">Active</span>;
    }
    return <span className="inline-flex px-2 py-1 text-xs font-medium rounded bg-gray-100 text-gray-800">Disabled</span>;
  };

  const getProviderLabel = (provider: string): string => {
    const labels: Record<string, string> = {
      pve: 'Proxmox VE',
      vmware: 'VMware',
      openstack: 'OpenStack',
      aws: 'AWS',
      aliyun: 'Aliyun',
      gcp: 'Google Cloud',
      azure: 'Azure',
    };
    return labels[provider] || provider;
  };

  if (!isAdmin) {
    return (
      <div className="p-6">
        <div className="card p-8 text-center">
          <p className="text-gray-500">You do not have permission to manage VM templates.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">VM Templates</h1>
        <p className="text-gray-600">Manage VM templates for cloud-init provisioning</p>
      </div>

      {/* Filters and Actions */}
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center space-x-4">
          <select
            value={filterProvider}
            onChange={(e) => setFilterProvider(e.target.value)}
            className="border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
          >
            <option value="">All Providers</option>
            {providersData?.providers?.map((provider) => (
              <option key={provider.id} value={provider.type}>{provider.name}</option>
            ))}
          </select>
        </div>
        <button 
          onClick={() => {
            setEditingTemplate(null);
            setShowModal(true);
          }} 
          className="btn btn-primary"
        >
          Add Template
        </button>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
          </div>
        ) : data?.templates?.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            No VM templates configured. Add one to get started.
          </div>
        ) : (
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Template</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">OS</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Min Specs</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {data?.templates?.map((template) => (
                <tr key={template.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">{template.name}</div>
                    <div className="text-sm text-gray-500">{template.description}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <span className="text-sm font-mono text-gray-900">{template.template_name}</span>
                      {template.cloud_init && (
                        <span className="ml-2 inline-flex px-2 py-0.5 text-xs font-medium rounded bg-blue-100 text-blue-800">Cloud-Init</span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {getProviderLabel(template.provider)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <div>{template.os_type}</div>
                    <div className="text-xs text-gray-400">{template.os_family} {template.os_version}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {template.min_cpu} vCPU / {template.min_memory_mb}MB / {template.min_disk_gb}GB
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getStatusBadge(template.status)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button onClick={() => handleEdit(template)} className="text-primary-600 hover:text-primary-900 mr-4">
                      Edit
                    </button>
                    <button onClick={() => handleDelete(template)} className="text-red-600 hover:text-red-900">
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <form onSubmit={handleSubmit}>
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">
                  {editingTemplate ? 'Edit VM Template' : 'Add VM Template'}
                </h3>
              </div>
              <div className="px-6 py-4 space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Display Name *</label>
                    <input
                      type="text"
                      name="name"
                      required
                      defaultValue={editingTemplate?.name}
                      placeholder="e.g., Ubuntu 22.04 LTS"
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Template Name *</label>
                    <input
                      type="text"
                      name="template_name"
                      required
                      defaultValue={editingTemplate?.template_name}
                      placeholder="e.g., ubuntu-22.04-template"
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                    <p className="mt-1 text-xs text-gray-500">The actual template name in the provider</p>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Provider *</label>
                    <select
                      name="provider"
                      required
                      defaultValue={editingTemplate?.provider}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="">Select Provider</option>
                      {providersData?.providers?.map((provider) => (
                        <option key={provider.id} value={provider.type}>{provider.name}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Zone</label>
                    <select
                      name="zone_id"
                      defaultValue={editingTemplate?.zone_id}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="">All Zones</option>
                      {zonesData?.zones?.map((zone: Zone) => (
                        <option key={zone.id} value={zone.id}>{zone.name}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">OS Type *</label>
                    <select
                      name="os_type"
                      required
                      defaultValue={editingTemplate?.os_type}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="">Select OS Type</option>
                      {OS_TYPES.map((os) => (
                        <option key={os.value} value={os.value}>{os.label}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">OS Family</label>
                    <select
                      name="os_family"
                      defaultValue={editingTemplate?.os_family}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="">Select OS Family</option>
                      {OS_FAMILIES.map((os) => (
                        <option key={os.value} value={os.value}>{os.label}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">OS Version</label>
                    <input
                      type="text"
                      name="os_version"
                      defaultValue={editingTemplate?.os_version}
                      placeholder="e.g., 22.04"
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Min CPU</label>
                    <input
                      type="number"
                      name="min_cpu"
                      min="1"
                      defaultValue={editingTemplate?.min_cpu || 1}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Min Memory (MB)</label>
                    <input
                      type="number"
                      name="min_memory_mb"
                      min="256"
                      step="256"
                      defaultValue={editingTemplate?.min_memory_mb || 512}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Min Disk (GB)</label>
                    <input
                      type="number"
                      name="min_disk_gb"
                      min="1"
                      defaultValue={editingTemplate?.min_disk_gb || 10}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Default User</label>
                    <input
                      type="text"
                      name="default_user"
                      defaultValue={editingTemplate?.default_user}
                      placeholder="e.g., ubuntu, centos, root"
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div className="flex items-center pt-6">
                    <input
                      type="checkbox"
                      name="cloud_init"
                      value="true"
                      defaultChecked={editingTemplate?.cloud_init ?? true}
                      className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                    />
                    <label className="ml-2 block text-sm text-gray-900">Cloud-Init Enabled</label>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <input
                    type="text"
                    name="description"
                    defaultValue={editingTemplate?.description}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>
              <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    setEditingTemplate(null);
                  }}
                  className="btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending || updateMutation.isPending}
                  className="btn btn-primary"
                >
                  {createMutation.isPending || updateMutation.isPending ? 'Saving...' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
