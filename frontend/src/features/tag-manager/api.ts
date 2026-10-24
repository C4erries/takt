import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { request, getAuthToken, type Tag, type TagWithMetrics, type UpdateTagRequest } from '@/shared/api';

export const tagManagerApi = {
  async listWorkspaceTagsMetrics(workspaceId: string): Promise<TagWithMetrics[]> {
    return request<TagWithMetrics[]>(`/workspaces/${workspaceId}/tags/metrics`, {
      method: 'GET',
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

  async deleteWorkspaceTag(workspaceId: string, tagId: string): Promise<void> {
    await request<void>(`/workspaces/${workspaceId}/tags/${tagId}`, {
      method: 'DELETE',
    });
  },

  async deleteUnusedWorkspaceTags(workspaceId: string): Promise<void> {
    await request<void>(`/workspaces/${workspaceId}/tags/unused`, {
      method: 'DELETE',
    });
  },
};

export const useWorkspaceTagsMetricsQuery = (workspaceId?: string) => {
  return useQuery({
    queryKey: ['tags-metrics', workspaceId],
    queryFn: async (): Promise<TagWithMetrics[]> => {
      if (!workspaceId || !getAuthToken()) {
        return [];
      }
      return tagManagerApi.listWorkspaceTagsMetrics(workspaceId);
    },
    enabled: Boolean(workspaceId),
  });
};

export interface UpdateTagVariables {
  tagId: string;
  data: UpdateTagRequest;
}

export const useUpdateWorkspaceTagMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ tagId, data }: UpdateTagVariables) => {
      if (!workspaceId || !getAuthToken()) {
        return {
          id: tagId,
          workspaceId: workspaceId || 'demo-workspace',
          name: data.name || '',
          colorHex: data.colorHex || '#10B981',
          createdAt: new Date().toISOString(),
        } as Tag;
      }
      return tagManagerApi.updateWorkspaceTag(workspaceId, tagId, data);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tags-metrics', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['tags', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      queryClient.invalidateQueries({ queryKey: ['project-tasks'] });
    },
  });
};

export const useDeleteWorkspaceTagMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (tagId: string) => {
      if (!workspaceId || !getAuthToken()) return;
      return tagManagerApi.deleteWorkspaceTag(workspaceId, tagId);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tags-metrics', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['tags', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      queryClient.invalidateQueries({ queryKey: ['project-tasks'] });
    },
  });
};

export const useDeleteUnusedTagsMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      if (!workspaceId || !getAuthToken()) return;
      return tagManagerApi.deleteUnusedWorkspaceTags(workspaceId);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tags-metrics', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['tags', workspaceId] });
    },
  });
};
