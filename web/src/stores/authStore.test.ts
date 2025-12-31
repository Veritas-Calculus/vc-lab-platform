import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAuthStore } from '@/stores/authStore';

// Mock the auth API
vi.mock('@/api/auth', () => ({
  authApi: {
    login: vi.fn(),
    logout: vi.fn(),
    refreshToken: vi.fn(),
    getCurrentUser: vi.fn(),
  },
}));

import { authApi } from '@/api/auth';

describe('authStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const { result } = renderHook(() => useAuthStore());

      expect(result.current.user).toBeNull();
      expect(result.current.accessToken).toBeNull();
      expect(result.current.refreshToken).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
    });
  });

  describe('login', () => {
    it('should login successfully', async () => {
      const mockTokenPair = {
        access_token: 'test-access-token',
        refresh_token: 'test-refresh-token',
        expires_at: '2024-12-31T23:59:59Z',
        token_type: 'Bearer',
      };

      const mockUser = {
        id: '123',
        username: 'testuser',
        email: 'test@example.com',
        display_name: 'Test User',
        phone: '',
        avatar: '',
        status: 1,
        last_login_at: null,
        last_login_ip: '',
        roles: [],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      vi.mocked(authApi.login).mockResolvedValue(mockTokenPair);
      vi.mocked(authApi.getCurrentUser).mockResolvedValue(mockUser);

      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.login('testuser', 'password123');
      });

      expect(result.current.isAuthenticated).toBe(true);
      expect(result.current.accessToken).toBe('test-access-token');
      expect(result.current.refreshToken).toBe('test-refresh-token');
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.error).toBeNull();
    });

    it('should handle login failure', async () => {
      vi.mocked(authApi.login).mockRejectedValue(new Error('Invalid credentials'));

      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        try {
          await result.current.login('testuser', 'wrongpassword');
        } catch {
          // Expected error
        }
      });

      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.accessToken).toBeNull();
      expect(result.current.error).toBe('Invalid credentials');
    });

    it('should set loading state during login', async () => {
      let resolveLogin: (value: unknown) => void;
      const loginPromise = new Promise((resolve) => {
        resolveLogin = resolve;
      });

      vi.mocked(authApi.login).mockReturnValue(loginPromise as never);

      const { result } = renderHook(() => useAuthStore());

      act(() => {
        result.current.login('testuser', 'password123');
      });

      expect(result.current.isLoading).toBe(true);

      await act(async () => {
        resolveLogin!({
          access_token: 'token',
          refresh_token: 'refresh',
          expires_at: '',
          token_type: 'Bearer',
        });
        vi.mocked(authApi.getCurrentUser).mockResolvedValue({} as never);
      });
    });
  });

  describe('logout', () => {
    it('should logout and clear state', async () => {
      // Set initial authenticated state
      useAuthStore.setState({
        user: { id: '123' } as never,
        accessToken: 'test-token',
        refreshToken: 'test-refresh',
        isAuthenticated: true,
      });

      vi.mocked(authApi.logout).mockResolvedValue(undefined);

      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.logout();
      });

      expect(result.current.user).toBeNull();
      expect(result.current.accessToken).toBeNull();
      expect(result.current.refreshToken).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
    });

    it('should handle logout API errors gracefully', async () => {
      useAuthStore.setState({
        user: { id: '123' } as never,
        accessToken: 'test-token',
        refreshToken: 'test-refresh',
        isAuthenticated: true,
      });

      vi.mocked(authApi.logout).mockRejectedValue(new Error('Network error'));

      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.logout();
      });

      // Should still clear state even if API call fails
      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.accessToken).toBeNull();
    });
  });

  describe('refreshTokens', () => {
    it('should refresh tokens successfully', async () => {
      useAuthStore.setState({
        refreshToken: 'old-refresh-token',
        isAuthenticated: true,
      });

      const mockNewTokens = {
        access_token: 'new-access-token',
        refresh_token: 'new-refresh-token',
        expires_at: '2024-12-31T23:59:59Z',
        token_type: 'Bearer',
      };

      vi.mocked(authApi.refreshToken).mockResolvedValue(mockNewTokens);

      const { result } = renderHook(() => useAuthStore());

      await act(async () => {
        await result.current.refreshTokens();
      });

      expect(result.current.accessToken).toBe('new-access-token');
      expect(result.current.refreshToken).toBe('new-refresh-token');
    });

    it('should throw error if no refresh token available', async () => {
      useAuthStore.setState({
        refreshToken: null,
      });

      const { result } = renderHook(() => useAuthStore());

      await expect(
        act(async () => {
          await result.current.refreshTokens();
        })
      ).rejects.toThrow('No refresh token available');
    });

    it('should logout if refresh fails', async () => {
      // Set initial state with authentication
      useAuthStore.setState({
        refreshToken: 'old-refresh-token',
        accessToken: 'old-access-token',
        isAuthenticated: true,
        user: { id: '123' } as never,
      });

      vi.mocked(authApi.refreshToken).mockRejectedValue(new Error('Token expired'));
      vi.mocked(authApi.logout).mockResolvedValue(undefined);

      const { result } = renderHook(() => useAuthStore());
      
      // Verify initial state
      expect(result.current.isAuthenticated).toBe(true);

      let error: Error | undefined;
      await act(async () => {
        try {
          await result.current.refreshTokens();
        } catch (e) {
          error = e as Error;
        }
      });

      expect(error?.message).toBe('Token expired');
      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.accessToken).toBeNull();
      expect(result.current.refreshToken).toBeNull();
    });
  });

  describe('clearError', () => {
    it('should clear error state', () => {
      useAuthStore.setState({
        error: 'Some error message',
      });

      const { result } = renderHook(() => useAuthStore());

      act(() => {
        result.current.clearError();
      });

      expect(result.current.error).toBeNull();
    });
  });

  describe('setUser', () => {
    it('should update user state', () => {
      const mockUser = {
        id: '456',
        username: 'updateduser',
        email: 'updated@example.com',
        display_name: 'Updated User',
        phone: '1234567890',
        avatar: '',
        status: 1,
        last_login_at: null,
        last_login_ip: '',
        roles: [],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      const { result } = renderHook(() => useAuthStore());

      act(() => {
        result.current.setUser(mockUser);
      });

      expect(result.current.user).toEqual(mockUser);
    });
  });
});
