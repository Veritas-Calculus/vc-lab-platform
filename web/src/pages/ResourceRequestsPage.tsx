import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { resourceRequestApi } from '@/api/resources';
import { regionApi, zoneApi, tfProviderApi, moduleApi } from '@/api/infra';
import { providerApi, credentialApi } from '@/api/settings';
import { sshKeyApi } from '@/api/sshkeys';
import { ipPoolApi } from '@/api/ipam';
import { vmTemplateApi } from '@/api/templates';
import { useAuthStore, useIsAdmin } from '@/stores/authStore';
import type { ResourceRequest, CreateResourceRequestReq, Environment, Zone, ProviderConfig, ProviderType, SSHKey, IPPool, VMTemplate, TerraformProvider, TerraformModule, Credential } from '@/types';

/**
 * Resource requests management page.
 */
export default function ResourceRequestsPage() {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const isAdmin = useIsAdmin();
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [selectedRegionId, setSelectedRegionId] = useState('');
  const [zones, setZones] = useState<Zone[]>([]);
  const [confirmModal, setConfirmModal] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
  }>({ isOpen: false, title: '', message: '', onConfirm: () => {} });
  const [inputModal, setInputModal] = useState<{
    isOpen: boolean;
    title: string;
    message: string;
    onConfirm: (value: string) => void;
  }>({ isOpen: false, title: '', message: '', onConfirm: () => {} });
  const [errorDetailModal, setErrorDetailModal] = useState<{
    isOpen: boolean;
    request: ResourceRequest | null;
  }>({ isOpen: false, request: null });

  // Fetch regions
  const { data: regionsData } = useQuery({
    queryKey: ['regions-all'],
    queryFn: () => regionApi.listAll(),
  });

  // Fetch providers from settings
  const { data: providersData } = useQuery({
    queryKey: ['settings-providers'],
    queryFn: () => providerApi.list({ pageSize: 100 }),
  });
  const providers: ProviderConfig[] = providersData?.providers || [];

  // Fetch SSH keys
  const { data: sshKeysData } = useQuery({
    queryKey: ['ssh-keys'],
    queryFn: () => sshKeyApi.list({ pageSize: 100 }),
  });
  const sshKeys: SSHKey[] = sshKeysData?.ssh_keys || [];

  // Fetch IP pools
  const { data: ipPoolsData } = useQuery({
    queryKey: ['ip-pools'],
    queryFn: () => ipPoolApi.list({ pageSize: 100 }),
  });
  const ipPools: IPPool[] = ipPoolsData?.ip_pools || [];

  // Fetch VM templates
  const { data: vmTemplatesData } = useQuery({
    queryKey: ['vm-templates'],
    queryFn: () => vmTemplateApi.list({ pageSize: 100 }),
  });
  const vmTemplates: VMTemplate[] = vmTemplatesData?.templates || [];

  // Fetch Terraform providers
  const { data: tfProvidersData } = useQuery({
    queryKey: ['tf-providers'],
    queryFn: () => tfProviderApi.list(1, 100),
  });
  const tfProviders: TerraformProvider[] = tfProvidersData?.providers || [];

  // Fetch Terraform modules
  const { data: tfModulesData } = useQuery({
    queryKey: ['tf-modules'],
    queryFn: () => moduleApi.list(1, 100),
  });
  const tfModules: TerraformModule[] = tfModulesData?.modules || [];

  // Fetch credentials
  const { data: credentialsData } = useQuery({
    queryKey: ['credentials'],
    queryFn: () => credentialApi.list({ pageSize: 100 }),
  });
  const credentials: Credential[] = credentialsData?.credentials || [];

  // Fetch requests
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ['resource-requests', page, statusFilter],
    queryFn: () =>
      resourceRequestApi.list({
        page,
        pageSize: 20,
        status: statusFilter || undefined,
      }),
  });

  // Create request mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateResourceRequestReq) => resourceRequestApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resource-requests'] });
      setShowModal(false);
    },
  });

  // Approve mutation
  const approveMutation = useMutation({
    mutationFn: (id: string) => resourceRequestApi.approve(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resource-requests'] });
    },
  });

  // Reject mutation
  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      resourceRequestApi.reject(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resource-requests'] });
    },
  });

  // Retry mutation
  const retryMutation = useMutation({
    mutationFn: (id: string) => resourceRequestApi.retry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resource-requests'] });
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => resourceRequestApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['resource-requests'] });
    },
  });

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    const spec = {
      cpu: parseInt(formData.get('cpu') as string) || 2,
      memory: parseInt(formData.get('memory') as string) || 4,
      disk: parseInt(formData.get('disk') as string) || 50,
      disk_type: 'ssd',
      os_type: formData.get('os_type') as string || 'linux',
      os_image: formData.get('os_image') as string || 'ubuntu-22.04',
      network: 'default',
      ssh_key_id: formData.get('ssh_key_id') as string || undefined,
      ip_pool_id: formData.get('ip_pool_id') as string || undefined,
      vm_template_id: formData.get('vm_template_id') as string || undefined,
    };

    const regionId = formData.get('region_id') as string;
    const zoneId = formData.get('zone_id') as string;
    const tfProviderId = formData.get('tf_provider_id') as string;
    const tfModuleId = formData.get('tf_module_id') as string;
    const credentialId = formData.get('credential_id') as string;

    createMutation.mutate({
      title: formData.get('title') as string,
      description: formData.get('description') as string,
      type: formData.get('type') as 'vm' | 'container' | 'bare_metal',
      environment: formData.get('environment') as Environment,
      provider: formData.get('provider') as ProviderType,
      region_id: regionId || undefined,
      zone_id: zoneId || undefined,
      tf_provider_id: tfProviderId || undefined,
      tf_module_id: tfModuleId || undefined,
      credential_id: credentialId || undefined,
      spec: JSON.stringify(spec),
      quantity: parseInt(formData.get('quantity') as string) || 1,
    });
  };

  const handleRegionChange = async (regionId: string) => {
    setSelectedRegionId(regionId);
    if (regionId) {
      const zoneList = await zoneApi.listByRegion(regionId);
      setZones(zoneList);
    } else {
      setZones([]);
    }
  };

  const handleApprove = (request: ResourceRequest) => {
    setConfirmModal({
      isOpen: true,
      title: 'Approve Request',
      message: `Are you sure you want to approve request "${request.title}"?`,
      onConfirm: () => {
        approveMutation.mutate(request.id);
        setConfirmModal({ ...confirmModal, isOpen: false });
      },
    });
  };

  const handleReject = (request: ResourceRequest) => {
    setInputModal({
      isOpen: true,
      title: 'Reject Request',
      message: 'Please provide a reason for rejection:',
      onConfirm: (reason: string) => {
        rejectMutation.mutate({ id: request.id, reason });
        setInputModal({ ...inputModal, isOpen: false });
      },
    });
  };

  const handleRetry = (request: ResourceRequest) => {
    setConfirmModal({
      isOpen: true,
      title: 'Retry Request',
      message: `Are you sure you want to retry provisioning for "${request.title}"?`,
      onConfirm: () => {
        retryMutation.mutate(request.id);
        setConfirmModal({ ...confirmModal, isOpen: false });
      },
    });
  };

  const handleDelete = (request: ResourceRequest) => {
    setConfirmModal({
      isOpen: true,
      title: 'Delete Request',
      message: `Are you sure you want to delete the request "${request.title}"? This action cannot be undone.`,
      onConfirm: () => {
        deleteMutation.mutate(request.id);
        setConfirmModal({ ...confirmModal, isOpen: false });
      },
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'approved':
        return 'bg-green-100 text-green-800';
      case 'rejected':
        return 'bg-red-100 text-red-800';
      case 'provisioning':
        return 'bg-blue-100 text-blue-800';
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Resource Requests</h1>
          <p className="text-gray-600">Request and manage compute resources</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="btn btn-secondary flex items-center gap-2"
          >
            <svg
              className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            {isFetching ? 'Refreshing...' : 'Refresh'}
          </button>
          <button onClick={() => setShowModal(true)} className="btn btn-primary">
            New Request
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="card mb-6">
        <div className="flex flex-wrap gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="input py-2"
            >
              <option value="">All Status</option>
              <option value="pending">Pending</option>
              <option value="approved">Approved</option>
              <option value="rejected">Rejected</option>
              <option value="provisioning">Provisioning</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
            </select>
          </div>
        </div>
      </div>

      {/* Error message */}
      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          {error instanceof Error ? error.message : 'Failed to load requests'}
        </div>
      )}

      {/* Requests table */}
      <div className="card overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Request
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Environment
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Provider
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Quantity
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Requester
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Created
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {data?.requests?.map((request) => (
                    <tr key={request.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4">
                        <div>
                          <p className="text-sm font-medium text-gray-900">{request.title}</p>
                          <p className="text-sm text-gray-500 truncate max-w-xs">
                            {request.description}
                          </p>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex px-2 py-1 text-xs font-medium rounded ${
                            request.environment === 'prod'
                              ? 'bg-red-100 text-red-800'
                              : request.environment === 'staging'
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-green-100 text-green-800'
                          }`}
                        >
                          {request.environment}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 uppercase">
                        {request.provider}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {request.quantity}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex px-2 py-1 text-xs font-medium rounded ${getStatusColor(request.status)}`}>
                          {request.status}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {request.requester?.display_name || request.requester?.username || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(request.created_at).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                        {isAdmin && request.status === 'pending' && (
                          <>
                            <button
                              onClick={() => handleApprove(request)}
                              className="text-green-600 hover:text-green-900"
                              disabled={approveMutation.isPending}
                            >
                              Approve
                            </button>
                            <button
                              onClick={() => handleReject(request)}
                              className="text-red-600 hover:text-red-900"
                              disabled={rejectMutation.isPending}
                            >
                              Reject
                            </button>
                          </>
                        )}
                        {request.status === 'failed' && (
                          <>
                            <button
                              onClick={() => setErrorDetailModal({ isOpen: true, request })}
                              className="text-orange-600 hover:text-orange-900"
                            >
                              View Error
                            </button>
                            {isAdmin && (
                              <button
                                onClick={() => handleRetry(request)}
                                className="text-blue-600 hover:text-blue-900"
                                disabled={retryMutation.isPending}
                              >
                                Retry
                              </button>
                            )}
                            <button
                              onClick={() => handleDelete(request)}
                              className="text-red-600 hover:text-red-900"
                              disabled={deleteMutation.isPending}
                            >
                              Delete
                            </button>
                          </>
                        )}
                        {request.status === 'rejected' && (
                          <button
                            onClick={() => handleDelete(request)}
                            className="text-red-600 hover:text-red-900"
                            disabled={deleteMutation.isPending}
                          >
                            Delete
                          </button>
                        )}
                        {request.requester_id === user?.id && request.status === 'pending' && (
                          <>
                            <span className="text-gray-400 mr-2">Awaiting approval</span>
                            <button
                              onClick={() => handleDelete(request)}
                              className="text-red-600 hover:text-red-900"
                              disabled={deleteMutation.isPending}
                            >
                              Cancel
                            </button>
                          </>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Empty state */}
            {!data?.requests?.length && (
              <div className="p-8 text-center text-gray-500">
                <p>No requests found</p>
              </div>
            )}

            {/* Pagination */}
            {data && data.total_pages > 1 && (
              <div className="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
                <p className="text-sm text-gray-500">
                  Showing page {page} of {data.total_pages} ({data.total} total)
                </p>
                <div className="flex gap-2">
                  <button
                    onClick={() => setPage(page - 1)}
                    disabled={page === 1}
                    className="btn btn-secondary disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setPage(page + 1)}
                    disabled={page === data.total_pages}
                    className="btn btn-secondary disabled:opacity-50"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Create Request Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">New Resource Request</h2>
              <form onSubmit={handleCreate}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Title *
                    </label>
                    <input name="title" type="text" className="input" required />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Description
                    </label>
                    <textarea name="description" className="input" rows={3} />
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Type *
                      </label>
                      <select name="type" className="input" required>
                        <option value="vm">Virtual Machine</option>
                        <option value="container">Container</option>
                        <option value="bare_metal">Bare Metal</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Environment *
                      </label>
                      <select name="environment" className="input" required>
                        <option value="dev">Development</option>
                        <option value="test">Test</option>
                        <option value="staging">Staging</option>
                        <option value="prod">Production</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Provider *
                      </label>
                      <select name="provider" className="input" required>
                        <option value="">Select Provider</option>
                        {providers.map((p) => (
                          <option key={p.id} value={p.type}>
                            {p.name} ({p.type})
                          </option>
                        ))}
                      </select>
                      {providers.length === 0 && (
                        <p className="text-xs text-orange-600 mt-1">
                          No providers configured. Please add providers in Settings first.
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Region
                      </label>
                      <select
                        name="region_id"
                        className="input"
                        value={selectedRegionId}
                        onChange={(e) => handleRegionChange(e.target.value)}
                      >
                        <option value="">Select Region</option>
                        {(regionsData || []).map((region) => (
                          <option key={region.id} value={region.id}>
                            {region.display_name}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Zone
                      </label>
                      <select name="zone_id" className="input" disabled={!selectedRegionId}>
                        <option value="">Select Zone</option>
                        {zones.map((zone) => (
                          <option key={zone.id} value={zone.id}>
                            {zone.display_name}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        CPU (cores)
                      </label>
                      <input name="cpu" type="number" className="input" min="1" max="64" defaultValue="2" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Memory (GB)
                      </label>
                      <input name="memory" type="number" className="input" min="1" max="256" defaultValue="4" />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Disk (GB)
                      </label>
                      <input name="disk" type="number" className="input" min="10" max="2000" defaultValue="50" />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        OS Type
                      </label>
                      <select name="os_type" className="input">
                        <option value="linux">Linux</option>
                        <option value="windows">Windows</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        OS Image
                      </label>
                      <select name="os_image" className="input">
                        <option value="ubuntu-22.04">Ubuntu 22.04</option>
                        <option value="ubuntu-20.04">Ubuntu 20.04</option>
                        <option value="centos-8">CentOS 8</option>
                        <option value="debian-12">Debian 12</option>
                        <option value="windows-2022">Windows Server 2022</option>
                      </select>
                    </div>
                  </div>
                  {/* New fields for Terragrunt config generation */}
                  <div className="border-t border-gray-200 pt-4 mt-4">
                    <h4 className="text-sm font-medium text-gray-900 mb-3">Advanced Configuration</h4>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          SSH Key
                        </label>
                        <select name="ssh_key_id" className="input">
                          <option value="">Use Default</option>
                          {sshKeys.map((key) => (
                            <option key={key.id} value={key.id}>
                              {key.name} {key.is_default ? '(default)' : ''}
                            </option>
                          ))}
                        </select>
                        {sshKeys.length === 0 && (
                          <p className="text-xs text-orange-600 mt-1">
                            No SSH keys configured. Add one in SSH Keys page.
                          </p>
                        )}
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          IP Pool
                        </label>
                        <select name="ip_pool_id" className="input">
                          <option value="">Auto-assign</option>
                          {ipPools.map((pool) => (
                            <option key={pool.id} value={pool.id}>
                              {pool.name} ({pool.cidr})
                            </option>
                          ))}
                        </select>
                        {ipPools.length === 0 && (
                          <p className="text-xs text-orange-600 mt-1">
                            No IP pools configured. Add one in IPAM page.
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="mt-4">
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        VM Template
                      </label>
                      <select name="vm_template_id" className="input">
                        <option value="">Select Template (optional)</option>
                        {vmTemplates.map((template) => (
                          <option key={template.id} value={template.id}>
                            {template.name} - {template.template_name} ({template.os_type})
                          </option>
                        ))}
                      </select>
                      <p className="text-xs text-gray-500 mt-1">
                        If selected, this template will be used for provisioning instead of OS image selection above.
                      </p>
                    </div>
                  </div>
                  {/* Terraform Configuration Section */}
                  <div className="border-t border-gray-200 pt-4 mt-4">
                    <h4 className="text-sm font-medium text-gray-900 mb-3">Terraform Configuration</h4>
                    <div className="grid grid-cols-3 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Terraform Provider *
                        </label>
                        <select name="tf_provider_id" className="input" required>
                          <option value="">Select Provider</option>
                          {tfProviders.map((p) => (
                            <option key={p.id} value={p.id}>
                              {p.name} ({p.namespace}/{p.name})
                            </option>
                          ))}
                        </select>
                        {tfProviders.length === 0 && (
                          <p className="text-xs text-orange-600 mt-1">
                            No Terraform providers configured. Add one in Infrastructure &gt; Providers.
                          </p>
                        )}
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Terraform Module *
                        </label>
                        <select name="tf_module_id" className="input" required>
                          <option value="">Select Module</option>
                          {tfModules.map((m) => (
                            <option key={m.id} value={m.id}>
                              {m.name} ({m.source})
                            </option>
                          ))}
                        </select>
                        {tfModules.length === 0 && (
                          <p className="text-xs text-orange-600 mt-1">
                            No Terraform modules configured. Add one in Infrastructure &gt; Modules.
                          </p>
                        )}
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          Credential *
                        </label>
                        <select name="credential_id" className="input" required>
                          <option value="">Select Credential</option>
                          {credentials.map((c) => (
                            <option key={c.id} value={c.id}>
                              {c.name} ({c.type})
                            </option>
                          ))}
                        </select>
                        {credentials.length === 0 && (
                          <p className="text-xs text-orange-600 mt-1">
                            No credentials configured. Add one in Settings &gt; Credentials.
                          </p>
                        )}
                      </div>
                    </div>
                    <p className="text-xs text-gray-500 mt-2">
                      These settings are required for Terraform-based provisioning. Configure providers, modules, and credentials in the respective management pages.
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Quantity
                    </label>
                    <input name="quantity" type="number" className="input" min="1" max="10" defaultValue="1" />
                  </div>
                </div>
                <div className="mt-6 flex justify-end gap-3">
                  <button
                    type="button"
                    onClick={() => setShowModal(false)}
                    className="btn btn-secondary"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="btn btn-primary"
                    disabled={createMutation.isPending}
                  >
                    {createMutation.isPending ? 'Submitting...' : 'Submit Request'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Confirm Modal */}
      {confirmModal.isOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">{confirmModal.title}</h3>
              <p className="text-gray-600">{confirmModal.message}</p>
            </div>
            <div className="flex justify-end gap-3 px-6 py-4 bg-gray-50 rounded-b-lg">
              <button
                onClick={() => setConfirmModal({ ...confirmModal, isOpen: false })}
                className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={confirmModal.onConfirm}
                className="px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Input Modal */}
      {inputModal.isOpen && (
        <InputModalComponent
          title={inputModal.title}
          message={inputModal.message}
          onConfirm={inputModal.onConfirm}
          onCancel={() => setInputModal({ ...inputModal, isOpen: false })}
        />
      )}

      {/* Error Detail Modal */}
      {errorDetailModal.isOpen && errorDetailModal.request && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[80vh] overflow-hidden">
            <div className="p-6 border-b border-gray-200">
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">Request Failed</h3>
                  <p className="text-sm text-gray-500">{errorDetailModal.request.title}</p>
                </div>
                <button
                  onClick={() => setErrorDetailModal({ isOpen: false, request: null })}
                  className="text-gray-400 hover:text-gray-500"
                >
                  <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <div className="p-6 overflow-y-auto max-h-[60vh]">
              <div className="mb-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">Error Message</h4>
                <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                  <pre className="text-sm text-red-800 whitespace-pre-wrap break-words font-mono">
                    {errorDetailModal.request.error_message || 'No error message available'}
                  </pre>
                </div>
              </div>
              {errorDetailModal.request.provision_log && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Provision Log</h4>
                  <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                    <pre className="text-sm text-gray-700 whitespace-pre-wrap break-words font-mono max-h-64 overflow-y-auto">
                      {errorDetailModal.request.provision_log}
                    </pre>
                  </div>
                </div>
              )}
            </div>
            <div className="flex justify-end px-6 py-4 bg-gray-50 rounded-b-lg border-t border-gray-200">
              <button
                onClick={() => setErrorDetailModal({ isOpen: false, request: null })}
                className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function InputModalComponent({
  title,
  message,
  onConfirm,
  onCancel,
}: {
  title: string;
  message: string;
  onConfirm: (value: string) => void;
  onCancel: () => void;
}) {
  const [value, setValue] = useState('');

  const handleConfirm = () => {
    if (value.trim()) {
      onConfirm(value);
      setValue('');
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
          <p className="text-gray-600 mb-4">{message}</p>
          <input
            type="text"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder="Enter reason..."
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            autoFocus
          />
        </div>
        <div className="flex justify-end gap-3 px-6 py-4 bg-gray-50 rounded-b-lg">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={!value.trim()}
            className="px-4 py-2 text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            Reject
          </button>
        </div>
      </div>
    </div>
  );
}
