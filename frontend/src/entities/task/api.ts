import { request } from '@/shared/api';
import type {
  Task,
  Tag,
  CreateTagRequest,
  UpdateTagRequest,
  UpdateTaskRequest,
  SetFocusRequest,
  CreateInboxTaskRequest,
  CreateTaskData,
} from './types';

export const taskDetailApi = {
  async getTodayTasks(workspaceId: string): Promise<Task[]> {
    return request<Task[]>(`/workspaces/${workspaceId}/tasks/today`, {
      method: 'GET',
    });
  },

  async listSubtasks(workspaceId: string, taskId: string): Promise<Task[]> {
    return request<Task[]>(`/workspaces/${workspaceId}/tasks/${taskId}/subtasks`, {
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

  async updateWorkspaceTag(
    workspaceId: string,
    tagId: string,
    data: UpdateTagRequest
  ): Promise<Tag> {
    return request<Tag>(`/workspaces/${workspaceId}/tags/${tagId}`, {
      method: 'PATCH',
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

  async duplicateTask(workspaceId: string, taskId: string): Promise<Task> {
    return request<Task>(`/workspaces/${workspaceId}/tasks/${taskId}/duplicate`, {
      method: 'POST',
    });
  },
};

export async function duplicateTask(workspaceId: string, taskId: string): Promise<Task> {
  return taskDetailApi.duplicateTask(workspaceId, taskId);
}

export async function listSubtasks(workspaceId: string, taskId: string): Promise<Task[]> {
  return taskDetailApi.listSubtasks(workspaceId, taskId);
}

export async function createInboxTask(data: CreateInboxTaskRequest): Promise<Task> {
  const body: Record<string, unknown> = {
    title: data.title,
  };
  if (data.descriptionMarkdown !== undefined) body.descriptionMarkdown = data.descriptionMarkdown;
  if (data.priority !== undefined) body.priority = data.priority;
  if (data.dueDate !== undefined) body.dueDate = data.dueDate;
  if (data.parentId !== undefined && data.parentId !== null) body.parentId = data.parentId;

  return request<Task>('/tasks/inbox', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export async function createTask(_workspaceId: string, data: CreateTaskData): Promise<Task> {
  const body: Record<string, unknown> = {
    title: data.title,
  };
  if (data.descriptionMarkdown !== undefined) body.descriptionMarkdown = data.descriptionMarkdown;
  if (data.priority !== undefined) body.priority = data.priority;
  if (data.dueDate !== undefined) body.dueDate = data.dueDate;
  if (data.parentId !== undefined && data.parentId !== null) body.parentId = data.parentId;

  return request<Task>('/tasks/inbox', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
