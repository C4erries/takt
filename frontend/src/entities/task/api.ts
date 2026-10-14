import { request } from '@/shared/api';
import type {
  Task,
  Tag,
  CreateTagRequest,
  UpdateTaskRequest,
  SetFocusRequest,
} from './types';

export const taskDetailApi = {
  async getTodayTasks(workspaceId: string): Promise<Task[]> {
    return request<Task[]>(`/workspaces/${workspaceId}/tasks/today`, {
      method: 'GET',
    });
  },

  async updateTask(
    workspaceId: string,
    taskId: string,
    data: UpdateTaskRequest
  ): Promise<Task> {
    return request<Task>(`/workspaces/${workspaceId}/tasks/${taskId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  async setTaskFocus(
    workspaceId: string,
    taskId: string,
    data: SetFocusRequest
  ): Promise<Task> {
    return request<Task>(`/workspaces/${workspaceId}/tasks/${taskId}/focus`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async listWorkspaceTags(workspaceId: string): Promise<Tag[]> {
    return request<Tag[]>(`/workspaces/${workspaceId}/tags`, {
      method: 'GET',
    });
  },

  async createWorkspaceTag(
    workspaceId: string,
    data: CreateTagRequest
  ): Promise<Tag> {
    return request<Tag>(`/workspaces/${workspaceId}/tags`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async deleteWorkspaceTag(
    workspaceId: string,
    tagId: string
  ): Promise<void> {
    await request<void>(`/workspaces/${workspaceId}/tags/${tagId}`, {
      method: 'DELETE',
    });
  },

  async addTagToTask(
    workspaceId: string,
    taskId: string,
    tagId: string
  ): Promise<void> {
    await request<void>(
      `/workspaces/${workspaceId}/tasks/${taskId}/tags/${tagId}`,
      {
        method: 'POST',
      }
    );
  },

  async removeTagFromTask(
    workspaceId: string,
    taskId: string,
    tagId: string
  ): Promise<void> {
    await request<void>(
      `/workspaces/${workspaceId}/tasks/${taskId}/tags/${tagId}`,
      {
        method: 'DELETE',
      }
    );
  },
};
