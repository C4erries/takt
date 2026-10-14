import { request } from '@/shared/api';
import type { Task } from '@/shared/api';
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  Board,
  Column,
  CreateColumnRequest,
  UpdateColumnRequest,
  MoveTaskRequest,
} from './types';

export const projectApi = {
  async listProjects(workspaceId: string): Promise<Project[]> {
    return request<Project[]>(`/workspaces/${workspaceId}/projects`, {
      method: 'GET',
    });
  },

  async getProject(workspaceId: string, projectId: string): Promise<Project> {
    return request<Project>(`/workspaces/${workspaceId}/projects/${projectId}`, {
      method: 'GET',
    });
  },

  async createProject(workspaceId: string, data: CreateProjectRequest): Promise<Project> {
    return request<Project>(`/workspaces/${workspaceId}/projects`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async updateProject(
    workspaceId: string,
    projectId: string,
    data: UpdateProjectRequest
  ): Promise<Project> {
    return request<Project>(`/workspaces/${workspaceId}/projects/${projectId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  async deleteProject(workspaceId: string, projectId: string): Promise<void> {
    await request<void>(`/workspaces/${workspaceId}/projects/${projectId}`, {
      method: 'DELETE',
    });
  },

  async listProjectBoards(workspaceId: string, projectId: string): Promise<Board[]> {
    return request<Board[]>(`/workspaces/${workspaceId}/projects/${projectId}/boards`, {
      method: 'GET',
    });
  },

  async createProjectBoard(
    workspaceId: string,
    projectId: string,
    name: string
  ): Promise<Board> {
    return request<Board>(`/workspaces/${workspaceId}/projects/${projectId}/boards`, {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  },

  async listBoardColumns(workspaceId: string, boardId: string): Promise<Column[]> {
    return request<Column[]>(`/workspaces/${workspaceId}/boards/${boardId}/columns`, {
      method: 'GET',
    });
  },

  async createBoardColumn(
    workspaceId: string,
    boardId: string,
    data: CreateColumnRequest
  ): Promise<Column> {
    return request<Column>(`/workspaces/${workspaceId}/boards/${boardId}/columns`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async updateBoardColumn(
    workspaceId: string,
    boardId: string,
    columnId: string,
    data: UpdateColumnRequest
  ): Promise<Column> {
    return request<Column>(
      `/workspaces/${workspaceId}/boards/${boardId}/columns/${columnId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      }
    );
  },

  async deleteBoardColumn(
    workspaceId: string,
    boardId: string,
    columnId: string
  ): Promise<void> {
    await request<void>(
      `/workspaces/${workspaceId}/boards/${boardId}/columns/${columnId}`,
      {
        method: 'DELETE',
      }
    );
  },

  async listProjectTasks(workspaceId: string, projectId: string): Promise<Task[]> {
    return request<Task[]>(
      `/workspaces/${workspaceId}/projects/${projectId}/tasks`,
      {
        method: 'GET',
      }
    );
  },

  async moveTask(
    workspaceId: string,
    taskId: string,
    data: MoveTaskRequest
  ): Promise<Task> {
    return request<Task>(`/workspaces/${workspaceId}/tasks/${taskId}/move`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
};
