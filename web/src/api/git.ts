/**
 * Git repository and node config API client
 */

import apiClient from './client';
import type {
  GitRepository,
  GitRepoListResponse,
  CreateGitRepoReq,
  UpdateGitRepoReq,
  TestConnectionReq,
  NodeConfig,
  NodeConfigListResponse,
  GitModuleListResponse,
} from '../types';

// Git Repository API
export const gitRepoApi = {
  list: async (page = 1, pageSize = 20): Promise<GitRepoListResponse> => {
    const response = await apiClient.get<GitRepoListResponse>('/git/repositories', {
      params: { page, page_size: pageSize },
    });
    return response.data;
  },

  get: async (id: string): Promise<GitRepository> => {
    const response = await apiClient.get<GitRepository>(`/git/repositories/${id}`);
    return response.data;
  },

  create: async (data: CreateGitRepoReq): Promise<GitRepository> => {
    const response = await apiClient.post<GitRepository>('/git/repositories', data);
    return response.data;
  },

  update: async (id: string, data: UpdateGitRepoReq): Promise<GitRepository> => {
    const response = await apiClient.put<GitRepository>(`/git/repositories/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/git/repositories/${id}`);
  },

  testConnection: async (id: string): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>(`/git/repositories/${id}/test`);
    return response.data;
  },

  testConnectionDirect: async (data: TestConnectionReq): Promise<{ message: string }> => {
    const response = await apiClient.post<{ message: string }>('/git/repositories/test-connection', data);
    return response.data;
  },
};

// Node Config API
export const nodeConfigApi = {
  list: async (repoId: string, page = 1, pageSize = 20): Promise<NodeConfigListResponse> => {
    const response = await apiClient.get<NodeConfigListResponse>('/git/node-configs', {
      params: { repo_id: repoId, page, page_size: pageSize },
    });
    return response.data;
  },

  get: async (id: string): Promise<NodeConfig> => {
    const response = await apiClient.get<NodeConfig>(`/git/node-configs/${id}`);
    return response.data;
  },

  getByRequest: async (requestId: string): Promise<NodeConfig> => {
    const response = await apiClient.get<NodeConfig>(`/git/node-configs/by-request/${requestId}`);
    return response.data;
  },

  commit: async (id: string, message: string): Promise<{ message: string; commit_sha: string }> => {
    const response = await apiClient.post<{ message: string; commit_sha: string }>(
      `/git/node-configs/${id}/commit`,
      { message }
    );
    return response.data;
  },
};

// Git Modules API - scan Terraform modules from git repository
export const gitModulesApi = {
  list: async (): Promise<GitModuleListResponse> => {
    const response = await apiClient.get<GitModuleListResponse>('/git/modules');
    return response.data;
  },

  sync: async (): Promise<GitModuleListResponse> => {
    const response = await apiClient.post<GitModuleListResponse>('/git/modules/sync');
    return response.data;
  },
};
