import apiClient from './client';
import type { TokenPair, User } from '@/types';

/**
 * Authentication API functions.
 */
export const authApi = {
  /**
   * Login with username and password.
   */
  async login(username: string, password: string): Promise<TokenPair> {
    const response = await apiClient.post<TokenPair>('/auth/login', {
      username,
      password,
    });
    return response.data;
  },

  /**
   * Refresh access token.
   */
  async refreshToken(refreshToken: string): Promise<TokenPair> {
    const response = await apiClient.post<TokenPair>('/auth/refresh', {
      refresh_token: refreshToken,
    });
    return response.data;
  },

  /**
   * Logout current user.
   */
  async logout(): Promise<void> {
    await apiClient.post('/auth/logout');
  },

  /**
   * Get current user profile.
   */
  async getCurrentUser(): Promise<User> {
    const response = await apiClient.get<User>('/users/me');
    return response.data;
  },

  /**
   * Update current user profile.
   */
  async updateCurrentUser(data: Partial<User>): Promise<User> {
    const response = await apiClient.put<User>('/users/me', data);
    return response.data;
  },

  /**
   * Change current user password.
   */
  async changePassword(oldPassword: string, newPassword: string): Promise<void> {
    await apiClient.put('/users/me/password', {
      old_password: oldPassword,
      new_password: newPassword,
    });
  },
};
