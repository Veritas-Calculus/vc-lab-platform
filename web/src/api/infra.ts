/**
 * Infrastructure API client for regions, zones, and terraform resources management.
 */

import apiClient from './client';
import type {
  Region,
  Zone,
  CreateRegionReq,
  UpdateRegionReq,
  CreateZoneReq,
  UpdateZoneReq,
  TerraformRegistry,
  TerraformProvider,
  TerraformModule,
  CreateTfRegistryReq,
  UpdateTfRegistryReq,
  CreateTfProviderReq,
  UpdateTfProviderReq,
  CreateTfModuleReq,
  UpdateTfModuleReq,
} from '../types';

interface RegionListResponse {
  regions: Region[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

interface ZoneListResponse {
  zones: Zone[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

interface RegistryListResponse {
  registries: TerraformRegistry[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

interface TfProviderListResponse {
  providers: TerraformProvider[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

interface ModuleListResponse {
  modules: TerraformModule[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export const regionApi = {
  list: async (page = 1, pageSize = 20): Promise<RegionListResponse> => {
    const response = await apiClient.get<RegionListResponse>(
      `/infra/regions?page=${page}&page_size=${pageSize}`
    );
    return response.data;
  },

  listAll: async (): Promise<Region[]> => {
    const response = await apiClient.get<{ regions: Region[] }>('/infra/regions?all=true');
    return response.data.regions || [];
  },

  getById: async (id: string): Promise<Region> => {
    const response = await apiClient.get<Region>(`/infra/regions/${id}`);
    return response.data;
  },

  create: async (data: CreateRegionReq): Promise<Region> => {
    const response = await apiClient.post<Region>('/infra/regions', data);
    return response.data;
  },

  update: async (id: string, data: UpdateRegionReq): Promise<Region> => {
    const response = await apiClient.put<Region>(`/infra/regions/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/infra/regions/${id}`);
  },
};

export const zoneApi = {
  list: async (page = 1, pageSize = 20): Promise<ZoneListResponse> => {
    const response = await apiClient.get<ZoneListResponse>(
      `/infra/zones?page=${page}&page_size=${pageSize}`
    );
    return response.data;
  },

  listAll: async (): Promise<ZoneListResponse> => {
    const response = await apiClient.get<ZoneListResponse>(
      `/infra/zones?all=true`
    );
    return response.data;
  },

  listByRegion: async (regionId: string): Promise<Zone[]> => {
    const response = await apiClient.get<{ zones: Zone[] }>(`/infra/zones?region_id=${regionId}`);
    return response.data.zones || [];
  },

  getById: async (id: string): Promise<Zone> => {
    const response = await apiClient.get<Zone>(`/infra/zones/${id}`);
    return response.data;
  },

  create: async (data: CreateZoneReq): Promise<Zone> => {
    const response = await apiClient.post<Zone>('/infra/zones', data);
    return response.data;
  },

  update: async (id: string, data: UpdateZoneReq): Promise<Zone> => {
    const response = await apiClient.put<Zone>(`/infra/zones/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/infra/zones/${id}`);
  },
};

export const registryApi = {
  list: async (page = 1, pageSize = 20): Promise<RegistryListResponse> => {
    const response = await apiClient.get<RegistryListResponse>(
      `/infra/registries?page=${page}&page_size=${pageSize}`
    );
    return response.data;
  },

  listAll: async (): Promise<TerraformRegistry[]> => {
    const response = await apiClient.get<{ registries: TerraformRegistry[] }>('/infra/registries?all=true');
    return response.data.registries || [];
  },

  getById: async (id: string): Promise<TerraformRegistry> => {
    const response = await apiClient.get<TerraformRegistry>(`/infra/registries/${id}`);
    return response.data;
  },

  create: async (data: CreateTfRegistryReq): Promise<TerraformRegistry> => {
    const response = await apiClient.post<TerraformRegistry>('/infra/registries', data);
    return response.data;
  },

  update: async (id: string, data: UpdateTfRegistryReq): Promise<TerraformRegistry> => {
    const response = await apiClient.put<TerraformRegistry>(`/infra/registries/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/infra/registries/${id}`);
  },
};

export const tfProviderApi = {
  list: async (page = 1, pageSize = 20): Promise<TfProviderListResponse> => {
    const response = await apiClient.get<TfProviderListResponse>(
      `/infra/providers?page=${page}&page_size=${pageSize}`
    );
    return response.data;
  },

  listByRegistry: async (registryId: string): Promise<TerraformProvider[]> => {
    const response = await apiClient.get<{ providers: TerraformProvider[] }>(`/infra/providers?registry_id=${registryId}`);
    return response.data.providers || [];
  },

  getById: async (id: string): Promise<TerraformProvider> => {
    const response = await apiClient.get<TerraformProvider>(`/infra/providers/${id}`);
    return response.data;
  },

  create: async (data: CreateTfProviderReq): Promise<TerraformProvider> => {
    const response = await apiClient.post<TerraformProvider>('/infra/providers', data);
    return response.data;
  },

  update: async (id: string, data: UpdateTfProviderReq): Promise<TerraformProvider> => {
    const response = await apiClient.put<TerraformProvider>(`/infra/providers/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/infra/providers/${id}`);
  },
};

export const moduleApi = {
  list: async (page = 1, pageSize = 20): Promise<ModuleListResponse> => {
    const response = await apiClient.get<ModuleListResponse>(
      `/infra/modules?page=${page}&page_size=${pageSize}`
    );
    return response.data;
  },

  listAll: async (): Promise<TerraformModule[]> => {
    const response = await apiClient.get<{ modules: TerraformModule[] }>('/infra/modules?all=true');
    return response.data.modules || [];
  },

  getById: async (id: string): Promise<TerraformModule> => {
    const response = await apiClient.get<TerraformModule>(`/infra/modules/${id}`);
    return response.data;
  },

  create: async (data: CreateTfModuleReq): Promise<TerraformModule> => {
    const response = await apiClient.post<TerraformModule>('/infra/modules', data);
    return response.data;
  },

  update: async (id: string, data: UpdateTfModuleReq): Promise<TerraformModule> => {
    const response = await apiClient.put<TerraformModule>(`/infra/modules/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/infra/modules/${id}`);
  },
};
