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
  DEMO_PROJECT_TASKS,
  updateDemoTaskPosition,
} from './demo';

export * from './demo';

export const useProjectsQuery = (workspaceId?: string) => {
  return useQuery<Project[], Error>({
    queryKey: ['projects', workspaceId],
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) return [];
      try {
        const data = await projectApi.listProjects(workspaceId);
        return data ?? [];
      } catch {
        return [];
      }
    },
    enabled: !!workspaceId,
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
    onSuccess: (created) => {
      if (workspaceId && created) {
        queryClient.setQueryData<Project[]>(['projects', workspaceId], (old = []) => [
          ...old.filter((p) => !p.id.startsWith('temp-') && p.id !== created.id),
          created,
        ]);
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
      if (!workspaceId || !projectId || !getAuthToken()) return [];
      try {
        const boards = await projectApi.listProjectBoards(workspaceId, projectId);
        return boards ?? [];
      } catch {
        return [];
      }
    },
    enabled: !!workspaceId && !!projectId,
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

export const useDeleteBoardMutation = (workspaceId?: string, projectId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (boardId: string) => {
      if (!workspaceId || !projectId || !getAuthToken()) {
        const queryKey = ['boards', workspaceId, projectId];
        const prev = queryClient.getQueryData<Board[]>(queryKey) ?? [];
        queryClient.setQueryData<Board[]>(queryKey, prev.filter((b) => b.id !== boardId));
        return;
      }
      return projectApi.deleteProjectBoard(workspaceId, projectId, boardId);
    },
    onMutate: async (boardId: string) => {
      const queryKey = ['boards', workspaceId, projectId];
      await queryClient.cancelQueries({ queryKey });
      const previousBoards = queryClient.getQueryData<Board[]>(queryKey);
      queryClient.setQueryData<Board[]>(queryKey, (old = []) =>
        old.filter((b) => b.id !== boardId)
      );
      return { previousBoards };
    },
    onError: (_err, _boardId, context) => {
      if (context?.previousBoards) {
        queryClient.setQueryData(['boards', workspaceId, projectId], context.previousBoards);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['boards', workspaceId, projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['tasks', 'inbox'] });
    },
  });
};

export interface UpdateBoardVariables {
  boardId: string;
  name: string;
}

export const useUpdateBoardMutation = (workspaceId?: string, projectId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ boardId, name }: UpdateBoardVariables) => {
      if (!workspaceId || !projectId || !getAuthToken()) {
        const queryKey = ['boards', workspaceId, projectId];
        const existing = queryClient.getQueryData<Board[]>(queryKey) ?? [];
        const found = existing.find((b) => b.id === boardId);
        return found ? { ...found, name } : ({ id: boardId, name } as Board);
      }
      return projectApi.updateProjectBoard(workspaceId, projectId, boardId, name);
    },
    onMutate: async ({ boardId, name }) => {
      const queryKey = ['boards', workspaceId, projectId];
      await queryClient.cancelQueries({ queryKey });
      const previousBoards = queryClient.getQueryData<Board[]>(queryKey);
      queryClient.setQueryData<Board[]>(queryKey, (old = []) =>
        old.map((b) => (b.id === boardId ? { ...b, name } : b))
      );
      return { previousBoards };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousBoards) {
        queryClient.setQueryData(['boards', workspaceId, projectId], context.previousBoards);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['boards', workspaceId, projectId] });
    },
  });
};

export * from './columnQueries';

export const useProjectTasksQuery = (workspaceId?: string, projectId?: string) => {
  return useQuery<Task[], Error>({
    queryKey: ['project-tasks', workspaceId, projectId],
    queryFn: async () => {
      if (!workspaceId || !projectId || !getAuthToken()) return [];
      try {
        const tasks = await projectApi.listProjectTasks(workspaceId, projectId);
        return tasks ?? [];
      } catch {
        return [];
      }
    },
    enabled: !!workspaceId && !!projectId,
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
