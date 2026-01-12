import apiClient from './client';
import type {
  VMTemplate,
  VMTemplateListResponse,
  CreateVMTemplateReq,
  UpdateVMTemplateReq,
} from '@/types';

/**
 * VM Template API client.
 */
export const vmTemplateApi = {
  /**
   * List VM templates with optional filtering.
   */
  async list(params?: {
    page?: number;
    pageSize?: number;
    provider?: string;
    osType?: string;
    zoneId?: string;
  }): Promise<VMTemplateListResponse> {
    const response = await apiClient.get<VMTemplateListResponse>('/infra/vm-templates', {
      params: {
        page: params?.page || 1,
        page_size: params?.pageSize || 20,
        provider: params?.provider,
        os_type: params?.osType,
        zone_id: params?.zoneId,
      },
    });
    return response.data;
  },

  /**
   * List all templates for a provider (for dropdowns).
   */
  async listByProvider(provider: string): Promise<{ templates: VMTemplate[] }> {
    const response = await apiClient.get<{ templates: VMTemplate[] }>('/infra/vm-templates', {
      params: {
        all: 'true',
        provider,
      },
    });
    return response.data;
  },

  /**
   * Get VM template by ID.
   */
  async getById(id: string): Promise<VMTemplate> {
    const response = await apiClient.get<VMTemplate>(`/infra/vm-templates/${id}`);
    return response.data;
  },

  /**
   * Create a new VM template.
   */
  async create(data: CreateVMTemplateReq): Promise<VMTemplate> {
    const response = await apiClient.post<VMTemplate>('/infra/vm-templates', data);
    return response.data;
  },

  /**
   * Update a VM template.
   */
  async update(id: string, data: UpdateVMTemplateReq): Promise<VMTemplate> {
    const response = await apiClient.put<VMTemplate>(`/infra/vm-templates/${id}`, data);
    return response.data;
  },

  /**
   * Delete a VM template.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/infra/vm-templates/${id}`);
  },
};
