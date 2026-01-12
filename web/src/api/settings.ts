import apiClient from './client';
import type {
  ProviderConfig,
  ProviderListResponse,
  CreateProviderReq,
  UpdateProviderReq,
  TestProviderConnectionReq,
  Credential,
  CredentialListResponse,
  CreateCredentialReq,
  UpdateCredentialReq,
  TestCredentialConnectionReq,
} from '@/types';

/**
 * Provider API client.
 */
export const providerApi = {
  /**
   * List providers with optional filtering.
   */
  async list(params?: {
    page?: number;
    pageSize?: number;
    type?: string;
  }): Promise<ProviderListResponse> {
    const response = await apiClient.get<ProviderListResponse>('/settings/providers', {
      params: {
        page: params?.page || 1,
        page_size: params?.pageSize || 20,
        type: params?.type,
      },
    });
    return response.data;
  },

  /**
   * Get provider by ID.
   */
  async getById(id: string): Promise<ProviderConfig> {
    const response = await apiClient.get<ProviderConfig>(`/settings/providers/${id}`);
    return response.data;
  },

  /**
   * Create a new provider.
   */
  async create(data: CreateProviderReq): Promise<ProviderConfig> {
    const response = await apiClient.post<ProviderConfig>('/settings/providers', data);
    return response.data;
  },

  /**
   * Update a provider.
   */
  async update(id: string, data: UpdateProviderReq): Promise<ProviderConfig> {
    const response = await apiClient.put<ProviderConfig>(`/settings/providers/${id}`, data);
    return response.data;
  },

  /**
   * Delete a provider.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/settings/providers/${id}`);
  },

  /**
   * Test provider connection.
   */
  async testConnection(data: TestProviderConnectionReq): Promise<{ message: string }> {
    const response = await apiClient.post<{ message: string }>('/settings/providers/test-connection', data);
    return response.data;
  },
};

/**
 * Credential API client.
 */
export const credentialApi = {
  /**
   * List credentials with optional filtering.
   */
  async list(params?: {
    page?: number;
    pageSize?: number;
    type?: string;
  }): Promise<CredentialListResponse> {
    const response = await apiClient.get<CredentialListResponse>('/settings/credentials', {
      params: {
        page: params?.page || 1,
        page_size: params?.pageSize || 20,
        type: params?.type,
      },
    });
    return response.data;
  },

  /**
   * Get credential by ID.
   */
  async getById(id: string): Promise<Credential> {
    const response = await apiClient.get<Credential>(`/settings/credentials/${id}`);
    return response.data;
  },

  /**
   * Create a new credential.
   */
  async create(data: CreateCredentialReq): Promise<Credential> {
    const response = await apiClient.post<Credential>('/settings/credentials', data);
    return response.data;
  },

  /**
   * Update a credential.
   */
  async update(id: string, data: UpdateCredentialReq): Promise<Credential> {
    const response = await apiClient.put<Credential>(`/settings/credentials/${id}`, data);
    return response.data;
  },

  /**
   * Delete a credential.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/settings/credentials/${id}`);
  },

  /**
   * Test credential connection.
   */
  async testConnection(data: TestCredentialConnectionReq): Promise<{ message: string }> {
    const response = await apiClient.post<{ message: string }>('/settings/credentials/test-connection', data);
    return response.data;
  },
};
