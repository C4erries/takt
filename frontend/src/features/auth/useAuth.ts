import { useState, useEffect, useCallback } from 'react';
import { authApi } from '@/shared/api';
import type { User, Workspace, LoginRequest, RegisterRequest } from '@/shared/api';

export interface UseAuthReturn {
  user: User | null;
  setUser: React.Dispatch<React.SetStateAction<User | null>>;
  workspace: Workspace | null;
  setWorkspace: React.Dispatch<React.SetStateAction<Workspace | null>>;
  isLoading: boolean;
  error: string | null;
  login: (data: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
  refresh: () => Promise<void>;
}

export const useAuth = (): UseAuthReturn => {
  const [user, setUser] = useState<User | null>(null);
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await authApi.getMe();
      setUser(data.user);
      setWorkspace(data.workspace);
      setError(null);
    } catch {
      setUser(null);
      setWorkspace(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const login = async (data: LoginRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await authApi.login(data);
      setUser(res.user);
      setWorkspace(res.workspace);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Ошибка авторизации';
      setError(msg);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (data: RegisterRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await authApi.register(data);
      setUser(res.user);
      setWorkspace(res.workspace);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Ошибка регистрации';
      setError(msg);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await authApi.logout();
    } finally {
      setUser(null);
      setWorkspace(null);
      setIsLoading(false);
    }
  };

  const clearError = () => setError(null);

  return {
    user,
    setUser,
    workspace,
    setWorkspace,
    isLoading,
    error,
    login,
    register,
    logout,
    clearError,
    refresh,
  };
};
