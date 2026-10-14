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

export const DEMO_PROJECT_ID = 'demo-project-1';
export const DEMO_BOARD_ID = 'demo-board-1';

export const DEMO_PROJECTS: Project[] = [
  {
    id: DEMO_PROJECT_ID,
    workspaceId: 'demo-workspace',
    name: 'Запуск Takt 0.3.1',
    description: 'Внедрение проектов и Канбан-доски с тактильным Drag-and-Drop',
    color: '#2D4F3E',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const DEMO_COLUMNS: Column[] = [
  {
    id: 'demo-col-todo',
    boardId: DEMO_BOARD_ID,
    name: 'К выполнению',
    position: 1000.0,
    statusCategory: 'todo',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'demo-col-in-progress',
    boardId: DEMO_BOARD_ID,
    name: 'В работе',
    position: 2000.0,
    statusCategory: 'in_progress',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'demo-col-done',
    boardId: DEMO_BOARD_ID,
    name: 'Готово',
    position: 3000.0,
    statusCategory: 'done',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const DEMO_PROJECT_TASKS: Task[] = [
  {
    id: 'pt-1',
    workspaceId: 'demo-workspace',
    projectId: DEMO_PROJECT_ID,
    columnId: 'demo-col-todo',
    title: 'Настроить Fractional Indexing для колонок',
    descriptionMarkdown: 'Новая позиция вычисляется как (pos_A + pos_B) / 2',
    isCompleted: false,
    priority: 'high',
    position: 1000.0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'pt-2',
    workspaceId: 'demo-workspace',
    projectId: DEMO_PROJECT_ID,
    columnId: 'demo-col-in-progress',
    title: 'Интегрировать @dnd-kit/core с тактильным Washi-дизайном',
    descriptionMarkdown: 'Ощущение японской канцелярии Midori MD при перетаскивании карточек',
    isCompleted: false,
    priority: 'urgent',
    position: 1000.0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'pt-3',
    workspaceId: 'demo-workspace',
    projectId: DEMO_PROJECT_ID,
    columnId: 'demo-col-done',
    title: 'Разработать спецификацию OpenAPI для проектов и досок',
    descriptionMarkdown: 'Контракты согласованы и зафиксированы в api/openapi/paths/projects.yaml',
    isCompleted: true,
    completedAt: new Date().toISOString(),
    priority: 'medium',
    position: 1000.0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const useProjectsQuery = (workspaceId?: string) => {
  return useQuery<Project[], Error>({
    queryKey: ['projects', workspaceId],
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) {
        return DEMO_PROJECTS;
      }
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
        const demo: Project = {
          id: `project-${Date.now()}`,
          workspaceId: workspaceId || 'demo-workspace',
          name: req.name,
          description: req.description || '',
          color: req.color || '#2D4F3E',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };
        return demo;
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

      queryClient.setQueryData<Project[]>(['projects', workspaceId], (old = []) => [
        ...old,
        optimistic,
      ]);

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
      if (!workspaceId || !projectId || !getAuthToken()) {
        return [
          {
            id: DEMO_BOARD_ID,
            projectId: projectId || DEMO_PROJECT_ID,
            name: 'Основная',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        ];
      }
      try {
        const boards = await projectApi.listProjectBoards(workspaceId, projectId);
        if (boards.length === 0) {
          return [
            {
              id: DEMO_BOARD_ID,
              projectId,
              name: 'Основная',
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            },
          ];
        }
        return boards;
      } catch {
        return [
          {
            id: DEMO_BOARD_ID,
            projectId: projectId || DEMO_PROJECT_ID,
            name: 'Основная',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          },
        ];
      }
    },
    enabled: !!projectId,
  });
};

export const useBoardColumnsQuery = (workspaceId?: string, boardId?: string) => {
  return useQuery<Column[], Error>({
    queryKey: ['columns', workspaceId, boardId],
    queryFn: async () => {
      if (!workspaceId || !boardId || !getAuthToken()) {
        return DEMO_COLUMNS;
      }
      try {
        const cols = await projectApi.listBoardColumns(workspaceId, boardId);
        return cols.length > 0 ? cols : DEMO_COLUMNS;
      } catch {
        return DEMO_COLUMNS;
      }
    },
    enabled: !!boardId,
  });
};

export const useProjectTasksQuery = (workspaceId?: string, projectId?: string) => {
  return useQuery<Task[], Error>({
    queryKey: ['project-tasks', workspaceId, projectId],
    queryFn: async () => {
      if (!workspaceId || !projectId || !getAuthToken()) {
        return DEMO_PROJECT_TASKS;
      }
      try {
        const tasks = await projectApi.listProjectTasks(workspaceId, projectId);
        return tasks.length > 0 ? tasks : DEMO_PROJECT_TASKS;
      } catch {
        return DEMO_PROJECT_TASKS;
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
        return {} as Task;
      }
      return projectApi.moveTask(workspaceId, taskId, data);
    },
    onMutate: async ({ taskId, data }) => {
      const queryKey = ['project-tasks', workspaceId, projectId];
      await queryClient.cancelQueries({ queryKey });
      const previousTasks = queryClient.getQueryData<Task[]>(queryKey);

      queryClient.setQueryData<Task[]>(queryKey, (old = []) =>
        old.map((t) =>
          t.id === taskId
            ? { ...t, columnId: data.columnId, position: data.position }
            : t
        )
      );

      return { previousTasks };
    },
    onError: (_err, _vars, context) => {
      const queryKey = ['project-tasks', workspaceId, projectId];
      if (context?.previousTasks) {
        queryClient.setQueryData(queryKey, context.previousTasks);
      }
    },
    onSettled: () => {
      const queryKey = ['project-tasks', workspaceId, projectId];
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
