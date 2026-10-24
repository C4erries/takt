import { useQuery } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import type { TaskActivity } from './types';
import { taskDetailApi } from './api';

export const useTaskActivitiesQuery = (
  workspaceId?: string,
  taskId?: string,
  enabled: boolean = true
) => {
  return useQuery<TaskActivity[], Error>({
    queryKey: ['task-activities', workspaceId, taskId],
    queryFn: async () => {
      if (!workspaceId || !taskId || !getAuthToken()) return [];
      try {
        const data = await taskDetailApi.listActivities(workspaceId, taskId);
        return data ?? [];
      } catch {
        return [];
      }
    },
    enabled: Boolean(workspaceId && taskId && enabled),
  });
};
