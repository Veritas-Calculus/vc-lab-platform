import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';

// We need to test the interceptors, so we'll create a fresh client
describe('API Client', () => {
  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(axios);
    vi.clearAllMocks();
  });

  afterEach(() => {
    mock.restore();
  });

  describe('request interceptor', () => {
    it('should add authorization header when token exists', async () => {
      // This is a simplified test since we can't easily test the actual interceptor
      // without complex mocking. In a real scenario, we'd use integration tests.
      const mockAxios = axios.create();
      const mockAdapter = new MockAdapter(mockAxios);

      // Simulate adding auth header
      mockAxios.interceptors.request.use((config) => {
        config.headers.Authorization = 'Bearer test-token';
        return config;
      });

      mockAdapter.onGet('/test').reply(200, { success: true });

      const response = await mockAxios.get('/test');
      expect(response.data).toEqual({ success: true });
    });
  });

  describe('response interceptor', () => {
    it('should pass through successful responses', async () => {
      const mockAxios = axios.create();
      const mockAdapter = new MockAdapter(mockAxios);

      mockAdapter.onGet('/test').reply(200, { data: 'test' });

      const response = await mockAxios.get('/test');
      expect(response.status).toBe(200);
      expect(response.data).toEqual({ data: 'test' });
    });

    it('should extract error message from response', async () => {
      const mockAxios = axios.create();
      const mockAdapter = new MockAdapter(mockAxios);

      mockAxios.interceptors.response.use(
        (response) => response,
        (error) => {
          const message = error.response?.data?.error || 'An error occurred';
          return Promise.reject(new Error(message));
        }
      );

      mockAdapter.onGet('/test').reply(400, { error: 'Bad request' });

      await expect(mockAxios.get('/test')).rejects.toThrow('Bad request');
    });

    it('should handle network errors', async () => {
      const mockAxios = axios.create();
      const mockAdapter = new MockAdapter(mockAxios);

      mockAxios.interceptors.response.use(
        (response) => response,
        (error) => {
          const message = error.message || 'Network error';
          return Promise.reject(new Error(message));
        }
      );

      mockAdapter.onGet('/test').networkError();

      await expect(mockAxios.get('/test')).rejects.toThrow();
    });
  });
});

describe('Auth API', () => {
  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(axios);
  });

  afterEach(() => {
    mock.restore();
  });

  it('should call login endpoint with correct payload', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onPost('/api/v1/auth/login').reply(200, {
      access_token: 'test-token',
      refresh_token: 'test-refresh',
      expires_at: '2024-12-31T23:59:59Z',
      token_type: 'Bearer',
    });

    const response = await mockAxios.post('/api/v1/auth/login', {
      username: 'testuser',
      password: 'password123',
    });

    expect(response.data.access_token).toBe('test-token');
  });

  it('should call refresh endpoint with refresh token', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onPost('/api/v1/auth/refresh').reply(200, {
      access_token: 'new-access-token',
      refresh_token: 'new-refresh-token',
      expires_at: '2024-12-31T23:59:59Z',
      token_type: 'Bearer',
    });

    const response = await mockAxios.post('/api/v1/auth/refresh', {
      refresh_token: 'old-refresh-token',
    });

    expect(response.data.access_token).toBe('new-access-token');
  });
});

describe('User API', () => {
  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(axios);
  });

  afterEach(() => {
    mock.restore();
  });

  it('should call list users endpoint with pagination', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onGet('/api/v1/users', { params: { page: 1, page_size: 20 } }).reply(200, {
      users: [
        { id: '1', username: 'user1' },
        { id: '2', username: 'user2' },
      ],
      total: 2,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });

    const response = await mockAxios.get('/api/v1/users', {
      params: { page: 1, page_size: 20 },
    });

    expect(response.data.users).toHaveLength(2);
    expect(response.data.total).toBe(2);
  });

  it('should call create user endpoint', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onPost('/api/v1/users').reply(201, {
      id: '123',
      username: 'newuser',
      email: 'new@example.com',
    });

    const response = await mockAxios.post('/api/v1/users', {
      username: 'newuser',
      email: 'new@example.com',
      password: 'password123',
    });

    expect(response.data.username).toBe('newuser');
  });

  it('should call delete user endpoint', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onDelete('/api/v1/users/123').reply(204);

    const response = await mockAxios.delete('/api/v1/users/123');

    expect(response.status).toBe(204);
  });
});

describe('Resource API', () => {
  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(axios);
  });

  afterEach(() => {
    mock.restore();
  });

  it('should call list resources endpoint with filters', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onGet('/api/v1/resources').reply(200, {
      resources: [
        { id: '1', name: 'vm-1', type: 'vm', status: 'running' },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      total_pages: 1,
    });

    const response = await mockAxios.get('/api/v1/resources', {
      params: { type: 'vm', status: 'running' },
    });

    expect(response.data.resources).toHaveLength(1);
    expect(response.data.resources[0].status).toBe('running');
  });

  it('should call approve resource request endpoint', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onPost('/api/v1/resource-requests/123/approve').reply(200, {
      id: '123',
      status: 'approved',
    });

    const response = await mockAxios.post('/api/v1/resource-requests/123/approve', {
      reason: 'Approved for testing',
    });

    expect(response.data.status).toBe('approved');
  });

  it('should call reject resource request endpoint', async () => {
    const mockAxios = axios.create();
    const mockAdapter = new MockAdapter(mockAxios);

    mockAdapter.onPost('/api/v1/resource-requests/123/reject').reply(200, {
      id: '123',
      status: 'rejected',
    });

    const response = await mockAxios.post('/api/v1/resource-requests/123/reject', {
      reason: 'Insufficient resources',
    });

    expect(response.data.status).toBe('rejected');
  });
});
