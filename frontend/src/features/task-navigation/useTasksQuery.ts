import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tasksApi, getAuthToken, type Task, type CreateInboxTaskRequest } from '@/shared/api';

export const TASKS_QUERY_KEY = ['tasks', 'inbox'] as const;

export const ONBOARDING_TASKS: Task[] = [
  {
    id: 'onboarding-1',
    workspaceId: 'demo-workspace',
    title: 'Нажми J и K для клавиатурного перемещения по строкам',
    descriptionMarkdown:
      'Клавиши J (вниз) и K (вверх) позволяют мгновенно перемещаться между задачами со snappy-анимацией.',
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
      if (!getAuthToken()) {
        return [];
      }
      return await tasksApi.getInbox();
    },
    enabled: !!getAuthToken(),
  });
};

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

      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, (old = []) => [
        optimisticTask,
        ...old,
      ]);

      return { previousTasks };
    },
    onError: (_err, _newTask, context) => {
      if (getAuthToken() && context?.previousTasks) {
        queryClient.setQueryData(TASKS_QUERY_KEY, context.previousTasks);
      }
    },
    onSettled: () => {
      if (getAuthToken()) {
        queryClient.invalidateQueries({ queryKey: TASKS_QUERY_KEY });
      }
    },
  });
};

export const useToggleTaskMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (taskId: string) => {
      if (!getAuthToken()) {
        return {} as Task;
      }
      return tasksApi.toggle(taskId);
    },
    onMutate: async (taskId) => {
      await queryClient.cancelQueries({ queryKey: TASKS_QUERY_KEY });
      const previousTasks = queryClient.getQueryData<Task[]>(TASKS_QUERY_KEY);

      const toggleInList = (old: Task[] | undefined) =>
        old
          ? old.map((task) =>
              task.id === taskId
                ? {
                    ...task,
                    isCompleted: !task.isCompleted,
                    completedAt: !task.isCompleted ? new Date().toISOString() : null,
                  }
                : task
            )
          : [];

      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, toggleInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['project-tasks'] }, toggleInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['tasks', 'today'] }, toggleInList);

      return { previousTasks };
    },
    onError: (_err, _taskId, context) => {
      if (getAuthToken() && context?.previousTasks) {
        queryClient.setQueryData(TASKS_QUERY_KEY, context.previousTasks);
      }
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
      if (!getAuthToken()) {
        return;
      }
      return tasksApi.delete(taskId);
    },
    onMutate: async (taskId) => {
      await queryClient.cancelQueries({ queryKey: TASKS_QUERY_KEY });
      const previousTasks = queryClient.getQueryData<Task[]>(TASKS_QUERY_KEY);

      const filterInList = (old: Task[] | undefined) =>
        old ? old.filter((task) => task.id !== taskId) : [];

      queryClient.setQueryData<Task[]>(TASKS_QUERY_KEY, filterInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['project-tasks'] }, filterInList);
      queryClient.setQueriesData<Task[]>({ queryKey: ['tasks', 'today'] }, filterInList);

      return { previousTasks };
    },
    onError: (_err, _taskId, context) => {
      if (getAuthToken() && context?.previousTasks) {
        queryClient.setQueryData(TASKS_QUERY_KEY, context.previousTasks);
      }
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
