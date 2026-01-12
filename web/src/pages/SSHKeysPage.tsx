import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { sshKeyApi } from '@/api/sshkeys';
import { useIsAdmin } from '@/stores/authStore';
import type { SSHKey, CreateSSHKeyReq } from '@/types';

/**
 * SSH Keys management page.
 */
export default function SSHKeysPage() {
  const queryClient = useQueryClient();
  const isAdmin = useIsAdmin();
  const [showModal, setShowModal] = useState(false);
  const [editingKey, setEditingKey] = useState<SSHKey | null>(null);

  // Fetch SSH keys
  const { data, isLoading } = useQuery({
    queryKey: ['ssh-keys'],
    queryFn: () => sshKeyApi.list(),
  });

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateSSHKeyReq) => sshKeyApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ssh-keys'] });
      setShowModal(false);
      setEditingKey(null);
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateSSHKeyReq> }) => 
      sshKeyApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ssh-keys'] });
      setShowModal(false);
      setEditingKey(null);
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => sshKeyApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ssh-keys'] });
    },
  });

  // Set default mutation
  const setDefaultMutation = useMutation({
    mutationFn: (id: string) => sshKeyApi.setDefault(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ssh-keys'] });
    },
  });

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data = {
      name: formData.get('name') as string,
      public_key: formData.get('public_key') as string,
      description: formData.get('description') as string,
      is_default: formData.get('is_default') === 'true',
    };

    if (editingKey) {
      updateMutation.mutate({ id: editingKey.id, data });
    } else {
      createMutation.mutate(data);
    }
  };

  const handleDelete = (key: SSHKey) => {
    if (confirm(`Delete SSH key "${key.name}"?`)) {
      deleteMutation.mutate(key.id);
    }
  };

  const handleEdit = (key: SSHKey) => {
    setEditingKey(key);
    setShowModal(true);
  };

  const handleSetDefault = (key: SSHKey) => {
    if (!key.is_default) {
      setDefaultMutation.mutate(key.id);
    }
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
          <p className="text-gray-500">You do not have permission to manage SSH keys.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">SSH Keys</h1>
        <p className="text-gray-600">Manage SSH public keys for VM provisioning</p>
      </div>

      {/* Actions */}
      <div className="flex justify-end mb-4">
        <button 
          onClick={() => {
            setEditingKey(null);
            setShowModal(true);
          }} 
          className="btn btn-primary"
        >
          Add SSH Key
        </button>
      </div>

      {/* Table */}
      <div className="card overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
          </div>
        ) : data?.ssh_keys?.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            No SSH keys configured. Add one to get started.
          </div>
        ) : (
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Fingerprint</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {data?.ssh_keys?.map((key) => (
                <tr key={key.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <span className="text-sm font-medium text-gray-900">{key.name}</span>
                      {key.is_default && (
                        <span className="ml-2 inline-flex px-2 py-0.5 text-xs font-medium rounded bg-blue-100 text-blue-800">Default</span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">
                    {key.fingerprint || 'N/A'}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {key.description || '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getStatusBadge(key.status)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button 
                      onClick={() => handleSetDefault(key)}
                      className={`text-blue-600 hover:text-blue-900 mr-4 ${key.is_default ? 'opacity-50 cursor-not-allowed' : ''}`}
                      disabled={key.is_default}
                    >
                      Set Default
                    </button>
                    <button onClick={() => handleEdit(key)} className="text-primary-600 hover:text-primary-900 mr-4">
                      Edit
                    </button>
                    <button onClick={() => handleDelete(key)} className="text-red-600 hover:text-red-900">
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
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
            <form onSubmit={handleSubmit}>
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">
                  {editingKey ? 'Edit SSH Key' : 'Add SSH Key'}
                </h3>
              </div>
              <div className="px-6 py-4 space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Name *</label>
                  <input
                    type="text"
                    name="name"
                    required
                    defaultValue={editingKey?.name}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Public Key *</label>
                  <textarea
                    name="public_key"
                    required
                    rows={4}
                    defaultValue={editingKey?.public_key}
                    placeholder="ssh-rsa AAAA... or ssh-ed25519 AAAA..."
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500 font-mono text-sm"
                  />
                  <p className="mt-1 text-sm text-gray-500">Paste your SSH public key (usually from ~/.ssh/id_rsa.pub or ~/.ssh/id_ed25519.pub)</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <input
                    type="text"
                    name="description"
                    defaultValue={editingKey?.description}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    name="is_default"
                    value="true"
                    defaultChecked={editingKey?.is_default}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                  />
                  <label className="ml-2 block text-sm text-gray-900">Set as default SSH key</label>
                </div>
              </div>
              <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    setEditingKey(null);
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
