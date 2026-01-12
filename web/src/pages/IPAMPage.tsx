import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ipPoolApi, ipAllocationApi } from '@/api/ipam';
import { zoneApi } from '@/api/infra';
import { useIsAdmin } from '@/stores/authStore';
import type { IPPool, IPAllocation, CreateIPPoolReq, Zone } from '@/types';

type ViewMode = 'pools' | 'allocations';

/**
 * IPAM (IP Address Management) page.
 */
export default function IPAMPage() {
  const queryClient = useQueryClient();
  const isAdmin = useIsAdmin();
  const [viewMode, setViewMode] = useState<ViewMode>('pools');
  const [showPoolModal, setShowPoolModal] = useState(false);
  const [showAllocateModal, setShowAllocateModal] = useState(false);
  const [editingPool, setEditingPool] = useState<IPPool | null>(null);
  const [selectedPoolId, setSelectedPoolId] = useState<string | null>(null);

  // Fetch IP pools
  const { data: poolsData, isLoading: poolsLoading } = useQuery({
    queryKey: ['ip-pools'],
    queryFn: () => ipPoolApi.list(),
  });

  // Fetch zones for dropdown
  const { data: zonesData } = useQuery({
    queryKey: ['zones', 'all'],
    queryFn: () => zoneApi.listAll(),
  });

  // Fetch allocations for selected pool
  const { data: allocationsData, isLoading: allocationsLoading } = useQuery({
    queryKey: ['ip-allocations', selectedPoolId],
    queryFn: () => selectedPoolId ? ipPoolApi.listAllocations(selectedPoolId) : null,
    enabled: !!selectedPoolId && viewMode === 'allocations',
  });

  // Create pool mutation
  const createPoolMutation = useMutation({
    mutationFn: (data: CreateIPPoolReq) => ipPoolApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip-pools'] });
      setShowPoolModal(false);
      setEditingPool(null);
    },
  });

  // Update pool mutation
  const updatePoolMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateIPPoolReq> }) => 
      ipPoolApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip-pools'] });
      setShowPoolModal(false);
      setEditingPool(null);
    },
  });

  // Delete pool mutation
  const deletePoolMutation = useMutation({
    mutationFn: (id: string) => ipPoolApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip-pools'] });
    },
  });

  // Allocate IP mutation
  const allocateMutation = useMutation({
    mutationFn: (data: { pool_id: string; hostname?: string; ip_address?: string }) => 
      ipAllocationApi.allocate(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip-allocations'] });
      queryClient.invalidateQueries({ queryKey: ['ip-pools'] });
      setShowAllocateModal(false);
    },
  });

  // Release IP mutation
  const releaseMutation = useMutation({
    mutationFn: (id: string) => ipAllocationApi.release(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ip-allocations'] });
      queryClient.invalidateQueries({ queryKey: ['ip-pools'] });
    },
  });

  const handleSubmitPool = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data: CreateIPPoolReq = {
      name: formData.get('name') as string,
      cidr: formData.get('cidr') as string,
      gateway: formData.get('gateway') as string,
      dns: formData.get('dns') as string,
      vlan_tag: parseInt(formData.get('vlan_tag') as string) || 0,
      start_ip: formData.get('start_ip') as string,
      end_ip: formData.get('end_ip') as string,
      zone_id: formData.get('zone_id') as string,
      network_type: formData.get('network_type') as string,
      description: formData.get('description') as string,
    };

    if (editingPool) {
      updatePoolMutation.mutate({ id: editingPool.id, data });
    } else {
      createPoolMutation.mutate(data);
    }
  };

  const handleAllocateIP = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    allocateMutation.mutate({
      pool_id: formData.get('pool_id') as string,
      hostname: formData.get('hostname') as string,
      ip_address: formData.get('ip_address') as string || undefined,
    });
  };

  const handleDeletePool = (pool: IPPool) => {
    if (confirm(`Delete IP pool "${pool.name}"? This cannot be undone if there are active allocations.`)) {
      deletePoolMutation.mutate(pool.id);
    }
  };

  const handleReleaseIP = (allocation: IPAllocation) => {
    if (confirm(`Release IP ${allocation.ip_address}?`)) {
      releaseMutation.mutate(allocation.id);
    }
  };

  const handleViewAllocations = (pool: IPPool) => {
    setSelectedPoolId(pool.id);
    setViewMode('allocations');
  };

  const getPoolStatusBadge = (status: number) => {
    const statusMap: Record<number, { label: string; color: string }> = {
      0: { label: 'Disabled', color: 'bg-gray-100 text-gray-800' },
      1: { label: 'Active', color: 'bg-green-100 text-green-800' },
    };
    const { label, color } = statusMap[status] || { label: 'Unknown', color: 'bg-gray-100 text-gray-800' };
    return (
      <span className={`inline-flex px-2 py-1 text-xs font-medium rounded ${color}`}>
        {label}
      </span>
    );
  };

  const getAllocationStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      available: 'bg-blue-100 text-blue-800',
      reserved: 'bg-yellow-100 text-yellow-800',
      allocated: 'bg-purple-100 text-purple-800',
    };
    return (
      <span className={`inline-flex px-2 py-1 text-xs font-medium rounded ${colors[status] || 'bg-gray-100 text-gray-800'}`}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
  };

  if (!isAdmin) {
    return (
      <div className="p-6">
        <div className="card p-8 text-center">
          <p className="text-gray-500">You do not have permission to manage IP addresses.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">IP Address Management</h1>
        <p className="text-gray-600">Manage IP pools and allocations for VM provisioning</p>
      </div>

      {/* View Toggle */}
      {viewMode === 'allocations' && selectedPoolId && (
        <div className="mb-4">
          <button
            onClick={() => setViewMode('pools')}
            className="text-primary-600 hover:text-primary-900 flex items-center"
          >
            ← Back to IP Pools
          </button>
        </div>
      )}

      {/* IP Pools View */}
      {viewMode === 'pools' && (
        <>
          <div className="flex justify-end mb-4">
            <button 
              onClick={() => {
                setEditingPool(null);
                setShowPoolModal(true);
              }} 
              className="btn btn-primary"
            >
              Add IP Pool
            </button>
          </div>

          <div className="card overflow-hidden">
            {poolsLoading ? (
              <div className="p-8 text-center">
                <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
              </div>
            ) : poolsData?.ip_pools?.length === 0 ? (
              <div className="p-8 text-center text-gray-500">
                No IP pools configured. Add one to get started.
              </div>
            ) : (
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CIDR</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Range</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Gateway</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Zone</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {poolsData?.ip_pools?.map((pool) => (
                    <tr key={pool.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">{pool.name}</div>
                        <div className="text-sm text-gray-500">{pool.description}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-mono">
                        {pool.cidr}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">
                        {pool.start_ip} - {pool.end_ip}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">
                        {pool.gateway}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {pool.zone?.name || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getPoolStatusBadge(pool.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button 
                          onClick={() => handleViewAllocations(pool)} 
                          className="text-blue-600 hover:text-blue-900 mr-4"
                        >
                          View IPs
                        </button>
                        <button 
                          onClick={() => {
                            setEditingPool(pool);
                            setShowPoolModal(true);
                          }} 
                          className="text-primary-600 hover:text-primary-900 mr-4"
                        >
                          Edit
                        </button>
                        <button 
                          onClick={() => handleDeletePool(pool)} 
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
        </>
      )}

      {/* Allocations View */}
      {viewMode === 'allocations' && selectedPoolId && (
        <>
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-lg font-medium text-gray-900">
              IP Allocations - {poolsData?.ip_pools?.find(p => p.id === selectedPoolId)?.name}
            </h2>
            <button 
              onClick={() => setShowAllocateModal(true)} 
              className="btn btn-primary"
            >
              Allocate IP
            </button>
          </div>

          <div className="card overflow-hidden">
            {allocationsLoading ? (
              <div className="p-8 text-center">
                <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
              </div>
            ) : allocationsData?.allocations?.length === 0 ? (
              <div className="p-8 text-center text-gray-500">
                No IP allocations in this pool.
              </div>
            ) : (
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">IP Address</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Hostname</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Allocated At</th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {allocationsData?.allocations?.map((allocation) => (
                    <tr key={allocation.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                        {allocation.ip_address}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {allocation.hostname || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getAllocationStatusBadge(allocation.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {allocation.allocated_at 
                          ? new Date(allocation.allocated_at).toLocaleString() 
                          : '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        {allocation.status === 'allocated' && (
                          <button 
                            onClick={() => handleReleaseIP(allocation)} 
                            className="text-red-600 hover:text-red-900"
                          >
                            Release
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </>
      )}

      {/* Pool Modal */}
      {showPoolModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <form onSubmit={handleSubmitPool}>
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">
                  {editingPool ? 'Edit IP Pool' : 'Add IP Pool'}
                </h3>
              </div>
              <div className="px-6 py-4 space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Name *</label>
                    <input
                      type="text"
                      name="name"
                      required
                      defaultValue={editingPool?.name}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Zone *</label>
                    <select
                      name="zone_id"
                      required
                      defaultValue={editingPool?.zone_id}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="">Select Zone</option>
                      {zonesData?.zones?.map((zone: Zone) => (
                        <option key={zone.id} value={zone.id}>{zone.name}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">CIDR *</label>
                  <input
                    type="text"
                    name="cidr"
                    required
                    placeholder="192.168.1.0/24"
                    defaultValue={editingPool?.cidr}
                    disabled={!!editingPool}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Start IP *</label>
                    <input
                      type="text"
                      name="start_ip"
                      required
                      placeholder="192.168.1.10"
                      defaultValue={editingPool?.start_ip}
                      disabled={!!editingPool}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">End IP *</label>
                    <input
                      type="text"
                      name="end_ip"
                      required
                      placeholder="192.168.1.254"
                      defaultValue={editingPool?.end_ip}
                      disabled={!!editingPool}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Gateway *</label>
                    <input
                      type="text"
                      name="gateway"
                      required
                      placeholder="192.168.1.1"
                      defaultValue={editingPool?.gateway}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">DNS</label>
                    <input
                      type="text"
                      name="dns"
                      placeholder="8.8.8.8,8.8.4.4"
                      defaultValue={editingPool?.dns}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">VLAN Tag</label>
                    <input
                      type="number"
                      name="vlan_tag"
                      min="0"
                      max="4094"
                      defaultValue={editingPool?.vlan_tag || 0}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Network Type</label>
                    <select
                      name="network_type"
                      defaultValue={editingPool?.network_type || 'bridge'}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                    >
                      <option value="bridge">Bridge</option>
                      <option value="nat">NAT</option>
                      <option value="vxlan">VXLAN</option>
                    </select>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <input
                    type="text"
                    name="description"
                    defaultValue={editingPool?.description}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>
              <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowPoolModal(false);
                    setEditingPool(null);
                  }}
                  className="btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={createPoolMutation.isPending || updatePoolMutation.isPending}
                  className="btn btn-primary"
                >
                  {createPoolMutation.isPending || updatePoolMutation.isPending ? 'Saving...' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Allocate Modal */}
      {showAllocateModal && selectedPoolId && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <form onSubmit={handleAllocateIP}>
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">Allocate IP Address</h3>
              </div>
              <div className="px-6 py-4 space-y-4">
                <input type="hidden" name="pool_id" value={selectedPoolId} />
                <div>
                  <label className="block text-sm font-medium text-gray-700">Hostname</label>
                  <input
                    type="text"
                    name="hostname"
                    placeholder="e.g., web-server-01"
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Specific IP (optional)</label>
                  <input
                    type="text"
                    name="ip_address"
                    placeholder="Leave empty for next available"
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                  />
                  <p className="mt-1 text-sm text-gray-500">Leave empty to allocate the next available IP</p>
                </div>
              </div>
              <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowAllocateModal(false)}
                  className="btn btn-secondary"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={allocateMutation.isPending}
                  className="btn btn-primary"
                >
                  {allocateMutation.isPending ? 'Allocating...' : 'Allocate'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
