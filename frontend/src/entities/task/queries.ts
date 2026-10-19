import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tasksApi, getAuthToken, type CreateInboxTaskRequest } from '@/shared/api';
import type { Task, Tag, CreateTagRequest, UpdateTaskRequest } from './types';
import { taskDetailApi } from './api';

export const TASKS_QUERY_KEY = ['tasks', 'inbox'] as const;

export const ONBOARDING_TASKS: Task[] = [
  {
    id: 'onboarding-1',
    workspaceId: 'demo-workspace',
    title: 'Нажми J и K для клавиатурного перемещения по строкам',
    descriptionMarkdown: 'Клавиши J (вниз) и K (вверх) позволяют мгновенно перемещаться между задачами со snappy-анимацией.',
    isCompleted: false,
    priority: 'high',
    position: 0,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'onboarding-2',
    workspaceId: 'demo-workspace',
    title: 'Нажми Пробел или кликни круглый чекбокс, чтобы завершить задачу',
    descriptionMarkdown: 'Клавиша Space или X переключает статус задачи.',
    isCompleted: false,
    priority: 'medium',
    position: 1,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'onboarding-3',
    workspaceId: 'demo-workspace',
    title: 'Нажми C в любом месте экрана для мгновенного сброса мысли (Quick Capture)',
    descriptionMarkdown: 'Горячая клавиша C открывает всплывающее окно быстрого захвата.',
    isCompleted: false,
    priority: 'pending',
    position: 2,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const useInboxTasks = () => {
  return useQuery<Task[], Error>({
    queryKey: TASKS_QUERY_KEY,
    queryFn: async () => {
      if (!getAuthToken()) return [];
      return await tasksApi.getInbox();
    },
    enabled: !!getAuthToken(),
  });
};

export const useTasksQuery = useInboxTasks;

export const useCreateTaskMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (req: CreateInboxTaskRequest) => {
      if (!getAuthToken()) {
        const demoTask: Task = {
          id: `task-${Date.now()}`,
          workspaceId: 'demo-workspace',
          title: req.title,
          descriptionMarkdown: req.descriptionMarkdown ?? '',
          isCompleted: false,
          priority: req.priority ?? 'none',
          position: 0,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        };
        return demoTask;
      }
      return tasksApi.createInbox(req);
    },
    onMutate: async (newTask) => {
      await queryClient.cancelQueries({ queryKey: TASKS_QUERY_KEY });
      const previousTasks = queryClient.getQueryData<Task[]>(TASKS_QUERY_KEY);
      const optimisticTask: Task = {
        id: `temp-${Date.now()}`,
        workspaceId: '',
        title: newTask.title,
        descriptionMarkdown: newTask.descriptionMarkdown ?? '',
        isCompleted: false,
        priority: newTask.priority ?? 'none',
        position: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, (old = []) => [optimisticTask, ...old]);
      return { previousTasks };
    },
    onError: (_err, _newTask, context) => {
      if (getAuthToken() && context?.previousTasks) queryClient.setQueryData(TASKS_QUERY_KEY, context.previousTasks);
    },
    onSettled: () => {
      if (getAuthToken()) queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
    },
  });
};

export const useToggleTaskMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (taskId: string) => {
      if (!getAuthToken()) return {} as Task;
      return tasksApi.toggle(taskId);
    },
    onMutate: async (taskId) => {
      await queryClient.cancelQueries({ queryKey: TASKS_QUERY_KEY });
      const previousTasks = queryClient.getQueryData<Task[]>(TASKS_QUERY_KEY);
      const toggleInList = (old: Task[] | undefined) =>
        old ? old.map((t) => t.id === taskId ? { ...t, isCompleted: !t.isCompleted, completedAt: !t.isCompleted ? new Date().toISOString() : null } : t) : [];
      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, toggleInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['project-tasks'] }, toggleInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['tasks', 'today'] }, toggleInList);
      return { previousTasks };
    },
    onError: (_err, _taskId, context) => {
      if (getAuthToken() && context?.previousTasks) queryClient.setQueryData(TASKS_QUERY_KEY, context.previousTasks);
    },
    onSettled: () => {
      if (getAuthToken()) {
        queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
        queryClient.invalidateQueries({ queryKey: ['project-tasks'] });
        queryClient.invalidateQueries({ queryKey: ['tasks', 'today'] });
      }
    },
  });
};

export const useDeleteTaskMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (taskId: string) => {
      if (!getAuthToken()) return;
      return tasksApi.delete(taskId);
    },
    onMutate: async (taskId) => {
      await queryClient.cancelQueries({ queryKey: TASKS_QUERY_KEY });
      const previousTasks = queryClient.getQueryData<Task[]>(TASKS_QUERY_KEY);
      const filterInList = (old: Task[] | undefined) => old ? old.filter((t) => t.id !== taskId) : [];
      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, filterInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['project-tasks'] }, filterInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['tasks', 'today'] }, filterInList);
      return { previousTasks };
    },
    onError: (_err, _taskId, context) => {
      if (getAuthToken() && context?.previousTasks) queryClient.setQueryData(TASKS_QUERY_KEY, context.previousTasks);
    },
    onSettled: () => {
      if (getAuthToken()) {
        queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
        queryClient.invalidateQueries({ queryKey: ['project-tasks'] });
        queryClient.invalidateQueries({ queryKey: ['tasks', 'today'] });
      }
    },
  });
};

export const DEMO_WORKSPACE_TAGS: Tag[] = [
  { id: 'tag-1', workspaceId: 'demo-workspace', name: 'фокус', colorHex: 'matcha', createdAt: new Date().toISOString() },
  { id: 'tag-2', workspaceId: 'demo-workspace', name: 'работа', colorHex: 'denim', createdAt: new Date().toISOString() },
  { id: 'tag-3', workspaceId: 'demo-workspace', name: '20 мин', colorHex: 'straw', createdAt: new Date().toISOString() },
  { id: 'tag-4', workspaceId: 'demo-workspace', name: 'дом', colorHex: 'persimmon', createdAt: new Date().toISOString() },
];

export const DEMO_TODAY_TASKS: Task[] = [
  {
    id: 'today-1',
    workspaceId: 'demo-workspace',
    title: 'Подготовить релиз Bento-инспектора и матричного календаря',
    descriptionMarkdown: 'Собрать воедино Editorial Split-Toggle и сетку месяца в стиле Hobonichi Techo.',
    isCompleted: false,
    priority: 'pending',
    focusOrder: 1,
    position: 1000.0,
    dueDate: new Date().toISOString().split('T')[0],
    tags: [DEMO_WORKSPACE_TAGS[0], DEMO_WORKSPACE_TAGS[1]],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'today-2',
    workspaceId: 'demo-workspace',
    title: 'Проверить отсутствие сырых Tailwind цветов через make check-arch',
    descriptionMarkdown: 'Строгое соблюдение палитры Washi, Tatami, Sumi, Matcha.',
    isCompleted: false,
    priority: 'high',
    focusOrder: 2,
    position: 2000.0,
    dueDate: new Date().toISOString().split('T')[0],
    tags: [DEMO_WORKSPACE_TAGS[2]],
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const useTodayTasksQuery = (workspaceId?: string) => {
  return useQuery<Task[], Error>({
    queryKey: ['tasks', 'today', workspaceId],
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) return [];
      try {
        const data = await taskDetailApi.getTodayTasks(workspaceId);
        return data ?? [];
      } catch {
        return [];
      }
    },
    enabled: !!workspaceId,
  });
};

export interface UpdateTaskVariables {
  taskId: string;
  data: UpdateTaskRequest;
}

export const useUpdateTaskMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ taskId, data }: UpdateTaskVariables) => {
      if (!workspaceId || !getAuthToken()) return {} as Task;
      return taskDetailApi.updateTask(workspaceId, taskId, data);
    },
    onMutate: async ({ taskId, data }) => {
      const updateInList = (old: Task[] | undefined) =>
        old ? old.map((t) => (t.id === taskId ? { ...t, ...data } : t)) : [];
      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, updateInList);
      queryClient.setQueryData<Task[]>(['tasks', 'today', workspaceId], updateInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['project-tasks'] }, updateInList);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ['tasks', 'today', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['project-tasks'] });
    },
  });
};

export interface SetFocusVariables {
  taskId: string;
  focusOrder: number | null;
}

export const useSetTaskFocusMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ taskId, focusOrder }: SetFocusVariables) => {
      if (!workspaceId || !getAuthToken()) return {} as Task;
      return taskDetailApi.setTaskFocus(workspaceId, taskId, { focusOrder });
    },
    onMutate: async ({ taskId, focusOrder }) => {
      const updateFocus = (old: Task[] | undefined) =>
        old ? old.map((t) => (t.id === taskId ? { ...t, focusOrder } : t)) : [];
      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, updateFocus);
      queryClient.setQueryData<Task[]>(['tasks', 'today', workspaceId], updateFocus);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks', 'today', workspaceId] });
      queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
    },
  });
};

export const useWorkspaceTagsQuery = (workspaceId?: string) => {
  return useQuery<Tag[], Error>({
    queryKey: ['tags', workspaceId],
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) return [];
      try {
        const data = await taskDetailApi.listWorkspaceTags(workspaceId);
        return data ?? [];
      } catch {
        return [];
      }
    },
    enabled: !!workspaceId,
  });
};

export const useCreateTagMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (req: CreateTagRequest) => {
      if (!workspaceId || !getAuthToken()) {
        const demo: Tag = {
          id: `tag-${Date.now()}`,
          workspaceId: workspaceId || 'demo-workspace',
          name: req.name,
          colorHex: req.colorHex || 'matcha',
          createdAt: new Date().toISOString(),
        };
        return demo;
      }
      return taskDetailApi.createWorkspaceTag(workspaceId, req);
    },
    onSuccess: (newTag) => {
      queryClient.setQueryData<Tag[]>(['tags', workspaceId], (old = []) => [...old, newTag]);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tags', workspaceId] });
    },
  });
};

export interface TaskTagVariables {
  taskId: string;
  tag: Tag;
}

export const useAddTagToTaskMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ taskId, tag }: TaskTagVariables) => {
      if (!workspaceId || !getAuthToken()) return;
      return taskDetailApi.addTagToTask(workspaceId, taskId, tag.id);
    },
    onMutate: async ({ taskId, tag }) => {
      const addTag = (old: Task[] | undefined) =>
        old ? old.map((t) => (t.id === taskId ? { ...t, tags: [...(t.tags?.filter((x) => x.id !== tag.id) ?? []), tag] } : t)) : [];
      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, addTag);
      queryClient.setQueryData<Task[]>(['tasks', 'today', workspaceId], addTag);
      queryClient.setQueriesData<Task[]>({ queryKey: ['project-tasks'] }, addTag);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ['tasks', 'today', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['project-tasks'] });
    },
  });
};

export const useRemoveTagFromTaskMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ taskId, tag }: TaskTagVariables) => {
      if (!workspaceId || !getAuthToken()) return;
      return taskDetailApi.removeTagFromTask(workspaceId, taskId, tag.id);
    },
    onMutate: async ({ taskId, tag }) => {
      const removeTag = (old: Task[] | undefined) =>
        old ? old.map((t) => (t.id === taskId ? { ...t, tags: (t.tags ?? []).filter((x) => x.id !== tag.id) } : t)) : [];
      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, removeTag);
      queryClient.setQueryData<Task[]>(['tasks', 'today', workspaceId], removeTag);
      queryClient.setQueriesData<Task[]>({ queryKey: ['project-tasks'] }, removeTag);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ['tasks', 'today', workspaceId] });
      queryClient.invalidateQueries({ queryKey: ['project-tasks'] });
    },
  });
};
