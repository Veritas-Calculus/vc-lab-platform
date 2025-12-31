import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User, TokenPair } from '@/types';
import { authApi } from '@/api/auth';

/**
 * Authentication store state interface.
 */
interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

/**
 * Authentication store actions interface.
 */
interface AuthActions {
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshTokens: () => Promise<void>;
  setUser: (user: User) => void;
  clearError: () => void;
}

type AuthStore = AuthState & AuthActions;

/**
 * Authentication store using Zustand with persistence.
 */
export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      // State
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      login: async (username: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const tokenPair = await authApi.login(username, password);
          set({
            accessToken: tokenPair.access_token,
            refreshToken: tokenPair.refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });
          
          // Fetch user profile after login
          const user = await authApi.getCurrentUser();
          set({ user });
        } catch (error) {
          const message = error instanceof Error ? error.message : 'Login failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      logout: async () => {
        const { accessToken } = get();
        try {
          if (accessToken) {
            await authApi.logout();
          }
        } catch {
          // Ignore logout errors
        } finally {
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
          });
        }
      },

      refreshTokens: async () => {
        const { refreshToken } = get();
        if (!refreshToken) {
          throw new Error('No refresh token available');
        }

        try {
          const tokenPair = await authApi.refreshToken(refreshToken);
          set({
            accessToken: tokenPair.access_token,
            refreshToken: tokenPair.refresh_token,
          });
        } catch (error) {
          // If refresh fails, logout
          await get().logout();
          throw error;
        }
      },

      setUser: (user: User) => {
        set({ user });
      },

      clearError: () => {
        set({ error: null });
      },
    }),
    {
      name: 'vc-lab-auth',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
        user: state.user,
      }),
    }
  )
);

/**
 * Hook to check if user has a specific role.
 */
export function useHasRole(roleCode: string): boolean {
  const user = useAuthStore((state) => state.user);
  return user?.roles?.some((role) => role.code === roleCode) ?? false;
}

/**
 * Hook to check if user is an admin.
 */
export function useIsAdmin(): boolean {
  return useHasRole('admin');
}
