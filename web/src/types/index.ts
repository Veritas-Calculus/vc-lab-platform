/**
 * Common types used across the application.
 */

// User source types
export type UserSource = 'local' | 'ldap' | 'oidc' | 'saml' | 'oauth2';

// User types
export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  phone: string;
  avatar: string;
  source: UserSource;
  external_id?: string;
  is_system: boolean;
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
  is_system: boolean;
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
export type ProviderType = 'pve' | 'vmware' | 'openstack' | 'aws' | 'aliyun' | 'gcp' | 'azure';
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
  type: 'vm' | 'container' | 'bare_metal';
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
  rejected_at: string | null;
  provision_started_at: string | null;
  provision_completed_at: string | null;
  provision_log: string;
  terraform_state: string;
  resource_id: string | null;
  resource?: Resource;
  error_message: string;
  reason: string;
  expires_at: string | null;
  created_at: string;
  updated_at: string;
}

export type RequestStatus = 'pending' | 'approved' | 'rejected' | 'provisioning' | 'completed' | 'failed';

export interface CreateResourceRequestReq {
  title: string;
  description?: string;
  type: 'vm' | 'container' | 'bare_metal';
  environment: Environment;
  provider: ProviderType;
  region_id?: string;
  zone_id?: string;
  tf_provider_id?: string;  // Selected Terraform provider
  tf_module_id?: string;    // Selected Terraform module
  credential_id?: string;   // Selected credential for access
  spec: string;
  quantity?: number;
  expires_at?: string;
}

// Pagination types
export interface PaginatedResponse<T = unknown> {
  data?: T[];
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

// Provider configuration types
export interface ProviderConfig {
  id: string;
  name: string;
  type: ProviderType;
  endpoint: string;
  description: string;
  config: string;
  status: number;
  is_default: boolean;
  credential_id: string | null;
  credential?: Credential;
  created_at: string;
  updated_at: string;
}

export interface CreateProviderReq {
  name: string;
  type: ProviderType;
  endpoint: string;
  description?: string;
  config?: string;
  is_default?: boolean;
  credential_id?: string;
}

export interface TestProviderConnectionReq {
  type: ProviderType;
  endpoint: string;
  credential_id?: string;
  config?: string;
}

export interface UpdateProviderReq {
  name?: string;
  endpoint?: string;
  description?: string;
  config?: string;
  status?: number;
  is_default?: boolean;
}

export interface ProviderListResponse extends PaginatedResponse<ProviderConfig> {
  providers: ProviderConfig[];
}

// Credential types
export interface Credential {
  id: string;
  name: string;
  type: ProviderType;
  zone_id: string | null;
  zone?: Zone;
  endpoint: string;
  provider_id: string | null;
  provider?: ProviderConfig;
  description: string;
  status: number;
  last_used_at: string | null;
  created_by_id: string;
  created_by?: User;
  created_at: string;
  updated_at: string;
}

export interface CreateCredentialReq {
  name: string;
  type: ProviderType;
  zone_id?: string;
  endpoint?: string;
  provider_id?: string;
  access_key?: string;
  secret_key?: string;
  token?: string;
  description?: string;
}

export interface UpdateCredentialReq {
  name?: string;
  zone_id?: string;
  endpoint?: string;
  access_key?: string;
  secret_key?: string;
  token?: string;
  description?: string;
  status?: number;
}

export interface TestCredentialConnectionReq {
  type: ProviderType;
  endpoint: string;
  access_key?: string;
  secret_key?: string;
  token?: string;
}

export interface CredentialListResponse extends PaginatedResponse<Credential> {
  credentials: Credential[];
}

// Region types
export interface Region {
  id: string;
  name: string;
  code: string;
  display_name: string;
  description: string;
  status: number;
  zones?: Zone[];
  created_at: string;
  updated_at: string;
}

export interface CreateRegionReq {
  name: string;
  code: string;
  display_name?: string;
  description?: string;
}

export interface UpdateRegionReq {
  name?: string;
  display_name?: string;
  description?: string;
  status?: number;
}

// Zone types
export interface Zone {
  id: string;
  name: string;
  code: string;
  display_name: string;
  description: string;
  region_id: string;
  region?: Region;
  status: number;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateZoneReq {
  name: string;
  code: string;
  display_name?: string;
  description?: string;
  region_id: string;
  is_default?: boolean;
}

export interface UpdateZoneReq {
  name?: string;
  display_name?: string;
  description?: string;
  status?: number;
  is_default?: boolean;
}

// Terraform Registry types
export interface TerraformRegistry {
  id: string;
  name: string;
  endpoint: string;
  description: string;
  status: number;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateTfRegistryReq {
  name: string;
  endpoint: string;
  username?: string;
  token?: string;
  description?: string;
  is_default?: boolean;
}

export interface UpdateTfRegistryReq {
  name?: string;
  endpoint?: string;
  username?: string;
  token?: string;
  description?: string;
  status?: number;
  is_default?: boolean;
}

export interface TfRegistryListResponse extends PaginatedResponse<TerraformRegistry> {
  registries: TerraformRegistry[];
}

// Terraform Provider types
export interface TerraformProvider {
  id: string;
  name: string;
  namespace: string;
  source: string;
  version: string;
  registry_id: string;
  registry?: TerraformRegistry;
  description: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface CreateTfProviderReq {
  name: string;
  namespace?: string;
  source?: string;
  version?: string;
  registry_id: string;
  description?: string;
}

export interface UpdateTfProviderReq {
  name?: string;
  namespace?: string;
  source?: string;
  version?: string;
  description?: string;
  status?: number;
}

export interface TfProviderListResponse extends PaginatedResponse<TerraformProvider> {
  providers: TerraformProvider[];
}

// Terraform Module types
export interface TerraformModule {
  id: string;
  name: string;
  source: string;
  version: string;
  registry_id: string | null;
  registry?: TerraformRegistry;
  provider_id: string | null;
  provider?: TerraformProvider;
  description: string;
  variables: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface CreateTfModuleReq {
  name: string;
  source: string;
  version?: string;
  registry_id?: string;
  provider_id?: string;
  description?: string;
  variables?: string;
}

export interface UpdateTfModuleReq {
  name?: string;
  source?: string;
  version?: string;
  registry_id?: string;
  provider_id?: string;
  description?: string;
  variables?: string;
  status?: number;
}

export interface TfModuleListResponse extends PaginatedResponse<TerraformModule> {
  modules: TerraformModule[];
}

// Git Repository types
export type GitRepoType = 'modules' | 'storage';
export type GitAuthType = 'none' | 'token' | 'password' | 'ssh_key';

// GitModule represents a Terraform module discovered from a git repository
export interface GitModule {
  name: string;
  path: string;
  description?: string;
  source: string;
  variables?: string[];
  outputs?: string[];
}

export interface GitModuleListResponse {
  modules: GitModule[];
  total: number;
  message?: string;
}

export interface GitRepository {
  id: string;
  name: string;
  type: GitRepoType;
  url: string;
  branch: string;
  auth_type: GitAuthType;
  username?: string;
  base_path: string;
  description: string;
  status: number;
  is_default: boolean;
  last_sync_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateGitRepoReq {
  name: string;
  type: GitRepoType;
  url: string;
  branch?: string;
  auth_type?: GitAuthType;
  username?: string;
  token?: string;
  ssh_key?: string;
  base_path?: string;
  description?: string;
  is_default?: boolean;
}

export interface UpdateGitRepoReq {
  name?: string;
  url?: string;
  branch?: string;
  auth_type?: GitAuthType;
  username?: string;
  token?: string;
  ssh_key?: string;
  base_path?: string;
  description?: string;
  status?: number;
  is_default?: boolean;
}

export interface TestConnectionReq {
  url: string;
  branch?: string;
  auth_type?: GitAuthType;
  username?: string;
  token?: string;
  ssh_key?: string;
}

export interface GitRepoListResponse extends PaginatedResponse<GitRepository> {
  repositories: GitRepository[];
}

// Node Config types
export type NodeConfigStatus = 'pending' | 'approved' | 'provisioning' | 'active' | 'failed' | 'destroying' | 'destroyed';

export interface NodeConfig {
  id: string;
  name: string;
  path: string;
  resource_request_id: string;
  resource_request?: ResourceRequest;
  storage_repo_id: string;
  storage_repo?: GitRepository;
  module_repo_id: string | null;
  module_repo?: GitRepository;
  terragrunt_config: string;
  terraform_vars: string;
  status: NodeConfigStatus;
  commit_sha: string;
  pending_commit_sha: string;
  terraform_state: string;
  provision_log: string;
  error_message: string;
  provisioned_at: string | null;
  destroyed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface NodeConfigListResponse extends PaginatedResponse<NodeConfig> {
  node_configs: NodeConfig[];
}

// API error type
export interface ApiError {
  error: string;
  details?: Record<string, string>;
}

// SSH Key types
export interface SSHKey {
  id: string;
  name: string;
  public_key: string;
  fingerprint: string;
  description: string;
  created_by_id: string;
  created_by?: User;
  is_default: boolean;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface CreateSSHKeyReq {
  name: string;
  public_key: string;
  description?: string;
  is_default?: boolean;
}

export interface UpdateSSHKeyReq {
  name?: string;
  public_key?: string;
  description?: string;
  is_default?: boolean;
  status?: number;
}

export interface SSHKeyListResponse extends PaginatedResponse<SSHKey> {
  ssh_keys: SSHKey[];
}

// IP Pool types
export interface IPPool {
  id: string;
  name: string;
  cidr: string;
  gateway: string;
  dns: string;
  vlan_tag: number;
  start_ip: string;
  end_ip: string;
  zone_id: string;
  zone?: Zone;
  network_type: string;
  description: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface CreateIPPoolReq {
  name: string;
  cidr: string;
  gateway: string;
  dns?: string;
  vlan_tag?: number;
  start_ip: string;
  end_ip: string;
  zone_id: string;
  network_type?: string;
  description?: string;
}

export interface UpdateIPPoolReq {
  name?: string;
  gateway?: string;
  dns?: string;
  vlan_tag?: number;
  description?: string;
  status?: number;
}

export interface IPPoolListResponse extends PaginatedResponse<IPPool> {
  ip_pools: IPPool[];
}

// IP Allocation types
export type IPAllocationStatus = 'available' | 'reserved' | 'allocated';

export interface IPAllocation {
  id: string;
  ip_pool_id: string;
  ip_pool?: IPPool;
  ip_address: string;
  hostname: string;
  resource_id: string;
  status: IPAllocationStatus;
  allocated_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface AllocateIPReq {
  pool_id: string;
  hostname?: string;
  resource_id?: string;
  ip_address?: string; // Optional: specific IP to allocate
}

export interface IPAllocationListResponse extends PaginatedResponse<IPAllocation> {
  allocations: IPAllocation[];
}

// VM Template types
export interface VMTemplate {
  id: string;
  name: string;
  template_name: string;
  provider: string;
  os_type: string;
  os_family: string;
  os_version: string;
  zone_id: string;
  zone?: Zone;
  min_cpu: number;
  min_memory_mb: number;
  min_disk_gb: number;
  default_user: string;
  cloud_init: boolean;
  description: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface CreateVMTemplateReq {
  name: string;
  template_name: string;
  provider: string;
  os_type: string;
  os_family?: string;
  os_version?: string;
  zone_id?: string;
  min_cpu?: number;
  min_memory_mb?: number;
  min_disk_gb?: number;
  default_user?: string;
  cloud_init?: boolean;
  description?: string;
}

export interface UpdateVMTemplateReq {
  name?: string;
  template_name?: string;
  os_type?: string;
  os_family?: string;
  os_version?: string;
  min_cpu?: number;
  min_memory_mb?: number;
  min_disk_gb?: number;
  default_user?: string;
  cloud_init?: boolean;
  description?: string;
  status?: number;
}

export interface VMTemplateListResponse extends PaginatedResponse<VMTemplate> {
  templates: VMTemplate[];
}
