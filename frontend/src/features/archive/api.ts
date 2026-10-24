import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { request, getAuthToken } from '@/shared/api';
import type { Task, UnarchiveTaskRequest, ArchiveCompletedColumnTasksResponse } from '@/entities/task';

export const archiveApi = {
  async getArchivedTasks(workspaceId: string, projectId?: string): Promise<Task[]> {
    const query = projectId ? `?project_id=${encodeURIComponent(projectId)}` : '';
    return request<Task[]>(`/workspaces/${workspaceId}/tasks/archived${query}`, {
      method: 'GET',
    });
  },

  async archiveCompletedColumnTasks(
    workspaceId: string,
    columnId: string
  ): Promise<ArchiveCompletedColumnTasksResponse> {
    return request<ArchiveCompletedColumnTasksResponse>(
      `/workspaces/${workspaceId}/columns/${columnId}/archive-completed`,
      {
        method: 'POST',
      }
    );
  },

  async unarchiveTask(
    workspaceId: string,
    taskId: string,
    data: UnarchiveTaskRequest
  ): Promise<Task> {
    return request<Task>(`/workspaces/${workspaceId}/tasks/${taskId}/unarchive`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },
};

export const useArchivedTasksQuery = (workspaceId?: string, projectId?: string) => {
  return useQuery<Task[], Error>({
    queryKey: ['tasks', 'archived', workspaceId, projectId ?? 'all'],
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) return [];
      try {
        const data = await archiveApi.getArchivedTasks(workspaceId, projectId);
        return data ?? [];
      } catch {
        return [];
      }
    },
    enabled: !!workspaceId,
  });
};

export const useArchiveCompletedColumnMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (columnId: string) => {
      if (!workspaceId || !getAuthToken()) {
        return { archivedCount: 0 };
      }
      return archiveApi.archiveCompletedColumnTasks(workspaceId, columnId);
    },
    onSuccess: () => {
      [
        ['tasks'],
        ['project-tasks'],
        ['columns'],
        ['tasks', 'archived'],
        ['subtasks'],
      ].forEach((key) => queryClient.invalidateQueries({ queryKey: key }));
    },
  });
};

export const useUnarchiveTaskMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ taskId, data }: { taskId: string; data: UnarchiveTaskRequest }) => {
      if (!workspaceId || !getAuthToken()) {
        return { id: taskId } as Task;
      }
      return archiveApi.unarchiveTask(workspaceId, taskId, data);
    },
    onSuccess: () => {
      [
        ['tasks', 'archived'],
        ['tasks'],
        ['project-tasks'],
        ['columns'],
      ].forEach((key) => queryClient.invalidateQueries({ queryKey: key }));
    },
  });
};
