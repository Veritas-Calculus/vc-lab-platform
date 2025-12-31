import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { roleApi } from '@/api/users';
import type { Role, CreateRoleRequest } from '@/types';

/**
 * Roles management page.
 */
export default function RolesPage() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [showModal, setShowModal] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);

  // Fetch roles
  const { data, isLoading, error } = useQuery({
    queryKey: ['roles', page],
    queryFn: () => roleApi.list(page, 20),
  });

  // Create role mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateRoleRequest) => roleApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setShowModal(false);
    },
  });

  // Delete role mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => roleApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
    },
  });

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);

    createMutation.mutate({
      name: formData.get('name') as string,
      code: formData.get('code') as string,
      description: formData.get('description') as string,
    });
  };

  const handleDelete = (role: Role) => {
    if (confirm(`Are you sure you want to delete role "${role.name}"?`)) {
      deleteMutation.mutate(role.id);
    }
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Roles</h1>
          <p className="text-gray-600">Manage roles and permissions</p>
        </div>
        <button
          onClick={() => {
            setEditingRole(null);
            setShowModal(true);
          }}
          className="btn btn-primary"
        >
          Add Role
        </button>
      </div>

      {/* Error message */}
      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg">
          {error instanceof Error ? error.message : 'Failed to load roles'}
        </div>
      )}

      {/* Roles grid */}
      {isLoading ? (
        <div className="p-8 text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {data?.roles?.map((role) => (
              <div key={role.id} className="card">
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">{role.name}</h3>
                    <p className="text-sm text-gray-500">Code: {role.code}</p>
                  </div>
                  <span
                    className={`inline-flex px-2 py-1 text-xs font-medium rounded ${
                      role.status === 1
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                    }`}
                  >
                    {role.status === 1 ? 'Active' : 'Disabled'}
                  </span>
                </div>
                <p className="text-sm text-gray-600 mb-4">
                  {role.description || 'No description'}
                </p>

                {/* Permissions */}
                <div className="mb-4">
                  <p className="text-xs font-medium text-gray-500 uppercase mb-2">
                    Permissions ({role.permissions?.length || 0})
                  </p>
                  <div className="flex flex-wrap gap-1">
                    {role.permissions?.slice(0, 5).map((perm) => (
                      <span
                        key={perm.id}
                        className="inline-flex px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded"
                      >
                        {perm.code}
                      </span>
                    ))}
                    {(role.permissions?.length ?? 0) > 5 && (
                      <span className="inline-flex px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded">
                        +{(role.permissions?.length ?? 0) - 5} more
                      </span>
                    )}
                  </div>
                </div>

                {/* Actions */}
                <div className="flex justify-end gap-2 pt-3 border-t">
                  <button
                    onClick={() => {
                      setEditingRole(role);
                      setShowModal(true);
                    }}
                    className="text-sm text-primary-600 hover:text-primary-900"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => handleDelete(role)}
                    className="text-sm text-red-600 hover:text-red-900"
                    disabled={deleteMutation.isPending}
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>

          {/* Pagination */}
          {data && data.total_pages > 1 && (
            <div className="mt-6 flex items-center justify-between">
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

      {/* Create/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                {editingRole ? 'Edit Role' : 'Create Role'}
              </h2>
              <form onSubmit={handleCreate}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Name
                    </label>
                    <input
                      name="name"
                      type="text"
                      defaultValue={editingRole?.name}
                      className="input"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Code
                    </label>
                    <input
                      name="code"
                      type="text"
                      defaultValue={editingRole?.code}
                      className="input"
                      required
                      pattern="^[a-z_]+$"
                      title="Only lowercase letters and underscores"
                      disabled={!!editingRole}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Description
                    </label>
                    <textarea
                      name="description"
                      defaultValue={editingRole?.description}
                      className="input"
                      rows={3}
                    />
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
                    {createMutation.isPending ? 'Saving...' : 'Save'}
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
