/**
 * Common types used across the application.
 */

// User types
export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  phone: string;
  avatar: string;
  status: number;
  last_login_at: string | null;
  last_login_ip: string;
  roles: Role[];
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  display_name?: string;
  phone?: string;
  role_ids?: string[];
}

export interface UpdateUserRequest {
  email?: string;
  display_name?: string;
  phone?: string;
  status?: number;
}

// Role types
export interface Role {
  id: string;
  name: string;
  code: string;
  description: string;
  status: number;
  permissions: Permission[];
  created_at: string;
  updated_at: string;
}

export interface CreateRoleRequest {
  name: string;
  code: string;
  description?: string;
  permission_ids?: string[];
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  status?: number;
}

// Permission types
export interface Permission {
  id: string;
  name: string;
  code: string;
  description: string;
  resource: string;
  action: string;
}

// Auth types
export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  token_type: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

// Resource types
export interface Resource {
  id: string;
  name: string;
  type: ResourceType;
  provider: ProviderType;
  status: ResourceStatus;
  spec: string;
  ip_address: string;
  hostname: string;
  owner_id: string;
  owner?: User;
  environment: Environment;
  external_id: string;
  expires_at: string | null;
  tags: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export type ResourceType = 'vm' | 'container' | 'bare_metal';
export type ProviderType = 'pve' | 'vmware' | 'openstack' | 'aws' | 'aliyun';
export type ResourceStatus = 'pending' | 'provisioning' | 'running' | 'stopped' | 'error';
export type Environment = 'dev' | 'test' | 'staging' | 'prod';

export interface ResourceSpec {
  cpu: number;
  memory: number;
  disk: number;
  disk_type: string;
  os_type: string;
  os_image: string;
  network: string;
}

export interface CreateResourceRequest {
  name: string;
  type: ResourceType;
  provider: ProviderType;
  environment: Environment;
  spec: ResourceSpec;
  owner_id: string;
  description?: string;
  tags?: string[];
}

// Resource request types
export interface ResourceRequest {
  id: string;
  title: string;
  description: string;
  spec: string;
  environment: Environment;
  provider: ProviderType;
  quantity: number;
  status: RequestStatus;
  requester_id: string;
  requester?: User;
  approver_id: string | null;
  approver?: User;
  approved_at: string | null;
  reason: string;
  expires_at: string | null;
  created_at: string;
  updated_at: string;
}

export type RequestStatus = 'pending' | 'approved' | 'rejected' | 'provisioning' | 'completed';

export interface CreateResourceRequestReq {
  title: string;
  description?: string;
  environment: Environment;
  provider: ProviderType;
  spec: ResourceSpec;
  quantity?: number;
  expires_at?: string;
}

// Pagination types
export interface PaginatedResponse<T> {
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface UserListResponse extends PaginatedResponse<User> {
  users: User[];
}

export interface RoleListResponse extends PaginatedResponse<Role> {
  roles: Role[];
}

export interface ResourceListResponse extends PaginatedResponse<Resource> {
  resources: Resource[];
}

export interface RequestListResponse extends PaginatedResponse<ResourceRequest> {
  requests: ResourceRequest[];
}

// API error type
export interface ApiError {
  error: string;
  details?: Record<string, string>;
}
