import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { resourceRequestApi } from '@/api/resources';
import { useAuthStore, useIsAdmin } from '@/stores/authStore';
import type { ResourceRequest, CreateResourceRequestReq, Environment, ProviderType } from '@/types';

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

  // Fetch requests
  const { data, isLoading, error } = useQuery({
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

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    createMutation.mutate({
      title: formData.get('title') as string,
      description: formData.get('description') as string,
      environment: formData.get('environment') as Environment,
      provider: formData.get('provider') as ProviderType,
      spec: {
        cpu: parseInt(formData.get('cpu') as string) || 2,
        memory: parseInt(formData.get('memory') as string) || 4,
        disk: parseInt(formData.get('disk') as string) || 50,
        disk_type: 'ssd',
        os_type: formData.get('os_type') as string || 'linux',
        os_image: formData.get('os_image') as string || 'ubuntu-22.04',
        network: 'default',
      },
      quantity: parseInt(formData.get('quantity') as string) || 1,
    });
  };

  const handleApprove = (request: ResourceRequest) => {
    if (confirm(`Approve request "${request.title}"?`)) {
      approveMutation.mutate(request.id);
    }
  };

  const handleReject = (request: ResourceRequest) => {
    const reason = prompt('Please provide a reason for rejection:');
    if (reason) {
      rejectMutation.mutate({ id: request.id, reason });
    }
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
        return 'bg-gray-100 text-gray-800';
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
        <button onClick={() => setShowModal(true)} className="btn btn-primary">
          New Request
        </button>
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
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        {isAdmin && request.status === 'pending' && (
                          <>
                            <button
                              onClick={() => handleApprove(request)}
                              className="text-green-600 hover:text-green-900 mr-3"
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
                        {request.requester_id === user?.id && request.status === 'pending' && (
                          <span className="text-gray-400">Awaiting approval</span>
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
                  <div className="grid grid-cols-2 gap-4">
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
                        <option value="pve">Proxmox VE</option>
                        <option value="vmware">VMware</option>
                        <option value="openstack">OpenStack</option>
                        <option value="aws">AWS</option>
                        <option value="aliyun">Aliyun</option>
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
    </div>
  );
}
