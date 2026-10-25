import { request } from '@/shared/api';
import type { User, UpdateUserProfileRequest } from '@/shared/api';

export const userApi = {
  async updateProfile(data: UpdateUserProfileRequest): Promise<User> {
    return request<User>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  async uploadAvatar(file: File): Promise<User> {
    const formData = new FormData();
    formData.append('avatar', file);
    return request<User>('/users/me/avatar', {
      method: 'POST',
      body: formData,
    });
  },
};
