import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import { projectApi } from './api';
import type { Column, UpdateColumnRequest } from './types';
import { DEMO_COLUMNS, updateDemoColumnPosition } from './demo';

export interface MoveColumnVariables {
  columnId: string;
  targetBoardId: string;
}

export const useMoveColumnMutation = (workspaceId?: string, boardId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ columnId, targetBoardId }: MoveColumnVariables) => {
      if (!workspaceId || !boardId || !getAuthToken()) {
        const sourceCols = queryClient.getQueryData<Column[]>(['columns', workspaceId, boardId]) ?? DEMO_COLUMNS;
        const col = sourceCols.find((c) => c.id === columnId);
        if (col) {
          const updated = { ...col, boardId: targetBoardId };
          queryClient.setQueryData<Column[]>(['columns', workspaceId, targetBoardId], (old = []) => [...old, updated]);
          return updated;
        }
        return {} as Column;
      }
      return projectApi.updateBoardColumn(workspaceId, boardId, columnId, { boardId: targetBoardId });
    },
    onMutate: async ({ columnId, targetBoardId }) => {
      const sourceKey = ['columns', workspaceId, boardId];
      const targetKey = ['columns', workspaceId, targetBoardId];
      await queryClient.cancelQueries({ queryKey: sourceKey });
      await queryClient.cancelQueries({ queryKey: targetKey });

      const previousSource = queryClient.getQueryData<Column[]>(sourceKey);
      const previousTarget = queryClient.getQueryData<Column[]>(targetKey);
      const movingCol = previousSource?.find((c) => c.id === columnId);

      if (movingCol) {
        queryClient.setQueryData<Column[]>(sourceKey, (old = []) => old.filter((c) => c.id !== columnId));
        queryClient.setQueryData<Column[]>(targetKey, (old = []) => [...old, { ...movingCol, boardId: targetBoardId }]);
      }
      return { previousSource, previousTarget };
    },
    onError: (_err, vars, context) => {
      if (context?.previousSource) queryClient.setQueryData(['columns', workspaceId, boardId], context.previousSource);
      if (context?.previousTarget) queryClient.setQueryData(['columns', workspaceId, vars.targetBoardId], context.previousTarget);
    },
    onSettled: (_data, _err, vars) => {
      queryClient.invalidateQueries({ queryKey: ['columns', workspaceId, boardId] });
      if (vars?.targetBoardId) {
        queryClient.invalidateQueries({ queryKey: ['columns', workspaceId, vars.targetBoardId] });
      }
    },
  });
};

export interface UpdateColumnVariables {
  columnId: string;
  data: UpdateColumnRequest;
}

export const useUpdateColumnMutation = (workspaceId?: string, boardId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ columnId, data }: UpdateColumnVariables) => {
      if (!workspaceId || !boardId || !getAuthToken()) {
        const queryKey = ['columns', workspaceId, boardId];
        const cols = queryClient.getQueryData<Column[]>(queryKey) ?? DEMO_COLUMNS;
        const updatedCols: Column[] = cols.map((c) => {
          if (c.id === columnId) {
            return {
              ...c,
              ...(data.name !== undefined ? { name: data.name } : {}),
              ...(data.position !== undefined ? { position: data.position } : {}),
              ...(data.statusCategory !== undefined ? { statusCategory: data.statusCategory } : {}),
              ...(data.boardId ? { boardId: data.boardId } : {}),
              isInbox: data.isInbox !== undefined ? data.isInbox : c.isInbox,
            };
          }
          if (data.isInbox) {
            return { ...c, isInbox: false };
          }
          return c;
        });
        queryClient.setQueryData<Column[]>(queryKey, updatedCols);
        return updatedCols.find((c) => c.id === columnId) as Column;
      }
      return projectApi.updateBoardColumn(workspaceId, boardId, columnId, data);
    },
    onMutate: async ({ columnId, data }) => {
      const queryKey = ['columns', workspaceId, boardId];
      await queryClient.cancelQueries({ queryKey });
      const previousColumns = queryClient.getQueryData<Column[]>(queryKey);

      queryClient.setQueryData<Column[]>(queryKey, (old = []) =>
        old.map((c) => {
          if (c.id === columnId) {
            return {
              ...c,
              ...(data.name !== undefined ? { name: data.name } : {}),
              ...(data.position !== undefined ? { position: data.position } : {}),
              ...(data.statusCategory !== undefined ? { statusCategory: data.statusCategory } : {}),
              ...(data.boardId ? { boardId: data.boardId } : {}),
              isInbox: data.isInbox !== undefined ? data.isInbox : c.isInbox,
            };
          }
          if (data.isInbox) {
            return { ...c, isInbox: false };
          }
          return c;
        })
      );

      return { previousColumns };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousColumns) {
        queryClient.setQueryData(['columns', workspaceId, boardId], context.previousColumns);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['columns', workspaceId, boardId] });
    },
  });
};

export const useBoardColumnsQuery = (workspaceId?: string, boardId?: string) => {
  return useQuery<Column[], Error>({
    queryKey: ['columns', workspaceId, boardId],
    queryFn: async () => {
      if (!workspaceId || !boardId || !getAuthToken()) {
        return [];
      }
      try {
        const cols = await projectApi.listBoardColumns(workspaceId, boardId);
        return [...(cols ?? [])].sort((a, b) => a.position - b.position);
      } catch {
        return [];
      }
    },
    enabled: !!workspaceId && !!boardId,
  });
};

export interface ReorderColumnVariables {
  columnId: string;
  position: number;
}

export const useReorderColumnMutation = (workspaceId?: string, boardId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ columnId, position }: ReorderColumnVariables) => {
      if (!workspaceId || !boardId || !getAuthToken()) {
        updateDemoColumnPosition(columnId, position);
        const col = DEMO_COLUMNS.find((c) => c.id === columnId);
        return col ? { ...col, position } : ({ id: columnId, position } as Column);
      }
      return projectApi.updateBoardColumn(workspaceId, boardId, columnId, { position });
    },
    onMutate: async ({ columnId, position }) => {
      if (!workspaceId || !boardId || !getAuthToken()) {
        updateDemoColumnPosition(columnId, position);
      }
      const queryKey = ['columns', workspaceId, boardId];
      await queryClient.cancelQueries({ queryKey });
      const previousColumns = queryClient.getQueryData<Column[]>(queryKey);

      queryClient.setQueryData<Column[]>(queryKey, (old = []) =>
        old
          .map((c) => (c.id === columnId ? { ...c, position } : c))
          .sort((a, b) => a.position - b.position)
      );

      return { previousColumns };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousColumns) {
        queryClient.setQueryData(['columns', workspaceId, boardId], context.previousColumns);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['columns', workspaceId, boardId] });
    },
  });
};

export interface CreateColumnVariables {
  name: string;
  statusCategory?: Column['statusCategory'];
  position?: number;
}

export const useCreateColumnMutation = (workspaceId?: string, boardId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ name, statusCategory = 'todo', position }: CreateColumnVariables) => {
      const queryKey = ['columns', workspaceId, boardId];
      const existing = queryClient.getQueryData<Column[]>(queryKey) ?? DEMO_COLUMNS;
      const maxPos = existing.reduce((max, c) => Math.max(max, c.position), 0);
      const computedPos = position !== undefined ? position : maxPos + 1000;

      if (!workspaceId || !boardId || !getAuthToken()) {
        const newCol: Column = {
          id: `demo-col-${Date.now()}`,
          boardId: boardId || 'demo-board',
          name,
          position: computedPos,
          statusCategory,
          isInbox: false,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };
        DEMO_COLUMNS.push(newCol);
        queryClient.setQueryData<Column[]>(queryKey, [...existing, newCol]);
        return newCol;
      }
      return projectApi.createBoardColumn(workspaceId, boardId, {
        name,
        statusCategory,
        position: computedPos,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['columns', workspaceId, boardId] });
    },
  });
};

