import { useMutation, useQueryClient } from '@tanstack/react-query';
import { userApi } from './api';
import type { User, UpdateUserProfileRequest } from '@/shared/api';

export const useUpdateProfileMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: UpdateUserProfileRequest): Promise<User> => {
      return userApi.updateProfile(data);
    },
    onSuccess: (updatedUser) => {
      queryClient.setQueryData(['auth', 'me'], (old: { user: User; workspace: unknown } | undefined) => {
        if (!old) return old;
        return { ...old, user: updatedUser };
      });
      queryClient.invalidateQueries({ queryKey: ['auth'] });
      queryClient.invalidateQueries({ queryKey: ['workspaces'] });
    },
  });
};

export const useUploadAvatarMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (file: File): Promise<User> => {
      return userApi.uploadAvatar(file);
    },
    onSuccess: (updatedUser) => {
      queryClient.setQueryData(['auth', 'me'], (old: { user: User; workspace: unknown } | undefined) => {
        if (!old) return old;
        return { ...old, user: updatedUser };
      });
      queryClient.invalidateQueries({ queryKey: ['auth'] });
      queryClient.invalidateQueries({ queryKey: ['workspaces'] });
    },
  });
};
