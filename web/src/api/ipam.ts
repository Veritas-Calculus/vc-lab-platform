import apiClient from './client';
import type {
  IPPool,
  IPPoolListResponse,
  CreateIPPoolReq,
  UpdateIPPoolReq,
  IPAllocation,
  IPAllocationListResponse,
  AllocateIPReq,
} from '@/types';

/**
 * IP Pool API client.
 */
export const ipPoolApi = {
  /**
   * List IP pools with optional zone filtering.
   */
  async list(params?: {
    page?: number;
    pageSize?: number;
    zoneId?: string;
  }): Promise<IPPoolListResponse> {
    const response = await apiClient.get<IPPoolListResponse>('/ipam/pools', {
      params: {
        page: params?.page || 1,
        page_size: params?.pageSize || 20,
        zone_id: params?.zoneId,
      },
    });
    return response.data;
  },

  /**
   * Get IP pool by ID with available count.
   */
  async getById(id: string): Promise<{ pool: IPPool; available_count: number }> {
    const response = await apiClient.get<{ pool: IPPool; available_count: number }>(`/ipam/pools/${id}`);
    return response.data;
  },

  /**
   * Create a new IP pool.
   */
  async create(data: CreateIPPoolReq): Promise<IPPool> {
    const response = await apiClient.post<IPPool>('/ipam/pools', data);
    return response.data;
  },

  /**
   * Update an IP pool.
   */
  async update(id: string, data: UpdateIPPoolReq): Promise<IPPool> {
    const response = await apiClient.put<IPPool>(`/ipam/pools/${id}`, data);
    return response.data;
  },

  /**
   * Delete an IP pool.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/ipam/pools/${id}`);
  },

  /**
   * List IP allocations for a pool.
   */
  async listAllocations(poolId: string, params?: {
    page?: number;
    pageSize?: number;
  }): Promise<IPAllocationListResponse> {
    const response = await apiClient.get<IPAllocationListResponse>(`/ipam/pools/${poolId}/allocations`, {
      params: {
        page: params?.page || 1,
        page_size: params?.pageSize || 50,
      },
    });
    return response.data;
  },
};

/**
 * IP Allocation API client.
 */
export const ipAllocationApi = {
  /**
   * Allocate an IP address from a pool.
   */
  async allocate(data: AllocateIPReq): Promise<IPAllocation> {
    const response = await apiClient.post<IPAllocation>('/ipam/allocations', data);
    return response.data;
  },

  /**
   * Release an allocated IP address.
   */
  async release(id: string): Promise<void> {
    await apiClient.delete(`/ipam/allocations/${id}`);
  },

  /**
   * Get IP allocations for a resource.
   */
  async getByResource(resourceId: string): Promise<{ allocations: IPAllocation[] }> {
    const response = await apiClient.get<{ allocations: IPAllocation[] }>(`/ipam/allocations/resource/${resourceId}`);
    return response.data;
  },
};
