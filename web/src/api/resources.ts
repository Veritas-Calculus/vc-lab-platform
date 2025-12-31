import apiClient from './client';
import type {
  Resource,
  ResourceListResponse,
  CreateResourceRequest,
  ResourceRequest,
  RequestListResponse,
  CreateResourceRequestReq,
} from '@/types';

interface ResourceListParams {
  page?: number;
  pageSize?: number;
  type?: string;
  provider?: string;
  status?: string;
  environment?: string;
  ownerId?: string;
}

interface RequestListParams {
  page?: number;
  pageSize?: number;
  status?: string;
  environment?: string;
  requesterId?: string;
}

/**
 * Resource API functions.
 */
export const resourceApi = {
  /**
   * List resources with pagination and filters.
   */
  async list(params: ResourceListParams = {}): Promise<ResourceListResponse> {
    const response = await apiClient.get<ResourceListResponse>('/resources', {
      params: {
        page: params.page ?? 1,
        page_size: params.pageSize ?? 20,
        type: params.type,
        provider: params.provider,
        status: params.status,
        environment: params.environment,
        owner_id: params.ownerId,
      },
    });
    return response.data;
  },

  /**
   * Get resource by ID.
   */
  async getById(id: string): Promise<Resource> {
    const response = await apiClient.get<Resource>(`/resources/${id}`);
    return response.data;
  },

  /**
   * Create a new resource.
   */
  async create(data: CreateResourceRequest): Promise<Resource> {
    const response = await apiClient.post<Resource>('/resources', data);
    return response.data;
  },

  /**
   * Update a resource.
   */
  async update(id: string, data: Partial<Resource>): Promise<Resource> {
    const response = await apiClient.put<Resource>(`/resources/${id}`, data);
    return response.data;
  },

  /**
   * Delete a resource.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/resources/${id}`);
  },
};

/**
 * Resource request API functions.
 */
export const resourceRequestApi = {
  /**
   * List resource requests with pagination and filters.
   */
  async list(params: RequestListParams = {}): Promise<RequestListResponse> {
    const response = await apiClient.get<RequestListResponse>('/resource-requests', {
      params: {
        page: params.page ?? 1,
        page_size: params.pageSize ?? 20,
        status: params.status,
        environment: params.environment,
        requester_id: params.requesterId,
      },
    });
    return response.data;
  },

  /**
   * Get resource request by ID.
   */
  async getById(id: string): Promise<ResourceRequest> {
    const response = await apiClient.get<ResourceRequest>(`/resource-requests/${id}`);
    return response.data;
  },

  /**
   * Create a new resource request.
   */
  async create(data: CreateResourceRequestReq): Promise<ResourceRequest> {
    const response = await apiClient.post<ResourceRequest>('/resource-requests', data);
    return response.data;
  },

  /**
   * Approve a resource request.
   */
  async approve(id: string, reason?: string): Promise<ResourceRequest> {
    const response = await apiClient.post<ResourceRequest>(`/resource-requests/${id}/approve`, {
      reason,
    });
    return response.data;
  },

  /**
   * Reject a resource request.
   */
  async reject(id: string, reason: string): Promise<ResourceRequest> {
    const response = await apiClient.post<ResourceRequest>(`/resource-requests/${id}/reject`, {
      reason,
    });
    return response.data;
  },
};
