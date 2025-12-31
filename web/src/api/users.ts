import apiClient from './client';
import type {
  User,
  UserListResponse,
  CreateUserRequest,
  UpdateUserRequest,
  Role,
  RoleListResponse,
  CreateRoleRequest,
  UpdateRoleRequest,
} from '@/types';

/**
 * User API functions.
 */
export const userApi = {
  /**
   * List users with pagination.
   */
  async list(page = 1, pageSize = 20): Promise<UserListResponse> {
    const response = await apiClient.get<UserListResponse>('/users', {
      params: { page, page_size: pageSize },
    });
    return response.data;
  },

  /**
   * Get user by ID.
   */
  async getById(id: string): Promise<User> {
    const response = await apiClient.get<User>(`/users/${id}`);
    return response.data;
  },

  /**
   * Create a new user.
   */
  async create(data: CreateUserRequest): Promise<User> {
    const response = await apiClient.post<User>('/users', data);
    return response.data;
  },

  /**
   * Update a user.
   */
  async update(id: string, data: UpdateUserRequest): Promise<User> {
    const response = await apiClient.put<User>(`/users/${id}`, data);
    return response.data;
  },

  /**
   * Delete a user.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/users/${id}`);
  },
};

/**
 * Role API functions.
 */
export const roleApi = {
  /**
   * List roles with pagination.
   */
  async list(page = 1, pageSize = 20): Promise<RoleListResponse> {
    const response = await apiClient.get<RoleListResponse>('/roles', {
      params: { page, page_size: pageSize },
    });
    return response.data;
  },

  /**
   * Get role by ID.
   */
  async getById(id: string): Promise<Role> {
    const response = await apiClient.get<Role>(`/roles/${id}`);
    return response.data;
  },

  /**
   * Create a new role.
   */
  async create(data: CreateRoleRequest): Promise<Role> {
    const response = await apiClient.post<Role>('/roles', data);
    return response.data;
  },

  /**
   * Update a role.
   */
  async update(id: string, data: UpdateRoleRequest): Promise<Role> {
    const response = await apiClient.put<Role>(`/roles/${id}`, data);
    return response.data;
  },

  /**
   * Delete a role.
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/roles/${id}`);
  },
};
