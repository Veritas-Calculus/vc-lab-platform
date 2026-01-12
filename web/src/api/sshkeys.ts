import apiClient from './client';
import type {
  SSHKey,
  SSHKeyListResponse,
  CreateSSHKeyReq,
  UpdateSSHKeyReq,
} from '@/types';

/**
 * SSH Key API client.
 */
export const sshKeyApi = {
  /**
   * List SSH keys with pagination.
   */
  async list(params?: {
    page?: number;
    pageSize?: number;
  }): Promise<SSHKeyListResponse> {
    const response = await apiClient.get<SSHKeyListResponse>('/settings/ssh-keys', {
      params: {
        page: params?.page || 1,
        page_size: params?.pageSize || 20,
      },
    });
    return response.data;
  },

  /**
   * Get SSH key by ID.
   */
  async getById(id: string): Promise<SSHKey> {
    const response = await apiClient.get<SSHKey>(`/settings/ssh-keys/${id}`);
    return response.data;
  },

  /**
   * Get default SSH key.
   */
  async getDefault(): Promise<SSHKey> {
    const response = await apiClient.get<SSHKey>('/settings/ssh-keys/default');
    return response.data;
  },

  /**
   * Create a new SSH key.
   */
  async create(data: CreateSSHKeyReq): Promise<SSHKey> {
    const response = await apiClient.post<SSHKey>('/settings/ssh-keys', data);
    return response.data;
  },

  /**
   * Update an SSH key.
   */
  async update(id: string, data: UpdateSSHKeyReq): Promise<SSHKey> {
    const response = await apiClient.put<SSHKey>(`/settings/ssh-keys/${id}`, data);
    return response.data;
  },

  /**
   * Delete an SSH key.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/settings/ssh-keys/${id}`);
  },

  /**
   * Set an SSH key as default.
   */
  async setDefault(id: string): Promise<void> {
    await apiClient.post(`/settings/ssh-keys/${id}/set-default`);
  },
};
