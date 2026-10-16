import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import type { Task } from '@/shared/api';
import { projectApi } from './api';
import type {
  Project,
  CreateProjectRequest,
  Board,
  Column,
  MoveTaskRequest,
} from './types';
import {
  DEMO_PROJECT_ID,
  DEMO_BOARDS,
  DEMO_PROJECTS,
  DEMO_COLUMNS,
  DEMO_PROJECT_TASKS,
  updateDemoTaskPosition,
  updateDemoColumnPosition,
} from './demo';

export * from './demo';

export const useProjectsQuery = (workspaceId?: string) => {
  return useQuery<Project[], Error>({
    queryKey: ['projects', workspaceId],
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) return DEMO_PROJECTS;
      try {
        const data = await projectApi.listProjects(workspaceId);
        return data.length > 0 ? data : DEMO_PROJECTS;
      } catch {
        return DEMO_PROJECTS;
      }
    },
    enabled: true,
  });
};

export const useCreateProjectMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (req: CreateProjectRequest) => {
      if (!workspaceId || !getAuthToken()) {
        return {
          id: `project-${Date.now()}`,
          workspaceId: workspaceId || 'demo-workspace',
          name: req.name,
          description: req.description || '',
          color: req.color || '#2D4F3E',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        } as Project;
      }
      return projectApi.createProject(workspaceId, req);
    },
    onMutate: async (newProject) => {
      await queryClient.cancelQueries({ queryKey: ['projects', workspaceId] });
      const previousProjects = queryClient.getQueryData<Project[]>(['projects', workspaceId]);
      const optimistic: Project = {
        id: `temp-${Date.now()}`,
        workspaceId: workspaceId || 'demo-workspace',
        name: newProject.name,
        description: newProject.description || '',
        color: newProject.color || '#2D4F3E',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      queryClient.setQueryData<Project[]>(['projects', workspaceId], (old = []) => [...old, optimistic]);
      return { previousProjects };
    },
    onError: (_err, _newProject, context) => {
      if (context?.previousProjects) {
        queryClient.setQueryData(['projects', workspaceId], context.previousProjects);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', workspaceId] });
    },
  });
};

export const useDeleteProjectMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (projectId: string) => {
      if (!workspaceId || !getAuthToken()) return;
      return projectApi.deleteProject(workspaceId, projectId);
    },
    onMutate: async (projectId) => {
      await queryClient.cancelQueries({ queryKey: ['projects', workspaceId] });
      const previousProjects = queryClient.getQueryData<Project[]>(['projects', workspaceId]);
      queryClient.setQueryData<Project[]>(['projects', workspaceId], (old = []) =>
        old.filter((p) => p.id !== projectId)
      );
      return { previousProjects };
    },
    onError: (_err, _projectId, context) => {
      if (context?.previousProjects) {
        queryClient.setQueryData(['projects', workspaceId], context.previousProjects);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['tasks', 'inbox'] });
    },
  });
};

export const useProjectBoardsQuery = (workspaceId?: string, projectId?: string) => {
  return useQuery<Board[], Error>({
    queryKey: ['boards', workspaceId, projectId],
    queryFn: async () => {
      if (!workspaceId || !projectId || !getAuthToken()) return DEMO_BOARDS;
      try {
        const boards = await projectApi.listProjectBoards(workspaceId, projectId);
        return boards.length > 0 ? boards : DEMO_BOARDS;
      } catch {
        return DEMO_BOARDS;
      }
    },
    enabled: !!projectId,
  });
};

export const useCreateBoardMutation = (workspaceId?: string, projectId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (name: string) => {
      if (!workspaceId || !projectId || !getAuthToken()) {
        const newBoardId = `board-${Date.now()}`;
        const demoBoard: Board = {
          id: newBoardId,
          projectId: projectId || DEMO_PROJECT_ID,
          name,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };
        const defaultCols: Column[] = [
          { id: `c-1-${Date.now()}`, boardId: newBoardId, name: 'К выполнению', position: 1000.0, statusCategory: 'todo', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
          { id: `c-2-${Date.now()}`, boardId: newBoardId, name: 'В работе', position: 2000.0, statusCategory: 'in_progress', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
          { id: `c-3-${Date.now()}`, boardId: newBoardId, name: 'Готово', position: 3000.0, statusCategory: 'done', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
        ];
        queryClient.setQueryData(['columns', workspaceId, newBoardId], defaultCols);
        return demoBoard;
      }
      return projectApi.createProjectBoard(workspaceId, projectId, name);
    },
    onMutate: async (name) => {
      const queryKey = ['boards', workspaceId, projectId];
      await queryClient.cancelQueries({ queryKey });
      const previousBoards = queryClient.getQueryData<Board[]>(queryKey);
      const optimistic: Board = {
        id: `temp-board-${Date.now()}`,
        projectId: projectId || DEMO_PROJECT_ID,
        name,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      queryClient.setQueryData<Board[]>(queryKey, (old = []) => [...old, optimistic]);
      return { previousBoards };
    },
    onError: (_err, _name, context) => {
      if (context?.previousBoards) {
        queryClient.setQueryData(['boards', workspaceId, projectId], context.previousBoards);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['boards', workspaceId, projectId] });
      queryClient.invalidateQueries({ queryKey: ['columns', workspaceId] });
    },
  });
};

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

export const useBoardColumnsQuery = (workspaceId?: string, boardId?: string) => {
  return useQuery<Column[], Error>({
    queryKey: ['columns', workspaceId, boardId],
    queryFn: async () => {
      if (!workspaceId || !boardId || !getAuthToken()) {
        const matching = DEMO_COLUMNS.filter((c) => c.boardId === boardId);
        const result = matching.length > 0 ? matching : DEMO_COLUMNS;
        return [...result].sort((a, b) => a.position - b.position);
      }
      try {
        const cols = await projectApi.listBoardColumns(workspaceId, boardId);
        const result = cols.length > 0 ? cols : DEMO_COLUMNS;
        return [...result].sort((a, b) => a.position - b.position);
      } catch {
        return [...DEMO_COLUMNS].sort((a, b) => a.position - b.position);
      }
    },
    enabled: !!boardId,
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

export const useProjectTasksQuery = (workspaceId?: string, projectId?: string) => {
  return useQuery<Task[], Error>({
    queryKey: ['project-tasks', workspaceId, projectId],
    queryFn: async () => {
      if (!workspaceId || !projectId || !getAuthToken()) return [...DEMO_PROJECT_TASKS];
      try {
        const tasks = await projectApi.listProjectTasks(workspaceId, projectId);
        return tasks.length > 0 ? tasks : [...DEMO_PROJECT_TASKS];
      } catch {
        return [...DEMO_PROJECT_TASKS];
      }
    },
    enabled: !!projectId,
  });
};

export interface MoveTaskVariables {
  taskId: string;
  data: MoveTaskRequest;
}

export const useMoveTaskMutation = (workspaceId?: string, projectId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ taskId, data }: MoveTaskVariables) => {
      if (!workspaceId || !getAuthToken()) {
        updateDemoTaskPosition(taskId, data.columnId, data.position);
        const task = DEMO_PROJECT_TASKS.find((t) => t.id === taskId);
        return task ? { ...task } : ({} as Task);
      }
      return projectApi.moveTask(workspaceId, taskId, data);
    },
    onMutate: async ({ taskId, data }) => {
      if (!workspaceId || !getAuthToken()) {
        updateDemoTaskPosition(taskId, data.columnId, data.position);
      }
      const queryKey = ['project-tasks', workspaceId, projectId];
      await queryClient.cancelQueries({ queryKey });
      const previousTasks = queryClient.getQueryData<Task[]>(queryKey);
      queryClient.setQueryData<Task[]>(queryKey, (old = []) =>
        old.map((t) => (t.id === taskId ? { ...t, columnId: data.columnId, position: data.position } : t))
      );
      return { previousTasks };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousTasks) {
        queryClient.setQueryData(['project-tasks', workspaceId, projectId], context.previousTasks);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['project-tasks', workspaceId, projectId] });
    },
  });
};
