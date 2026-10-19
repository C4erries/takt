// Re-export task query hooks and keys from entities/task to prevent FSD circular dependencies
export {
  TASKS_QUERY_KEY,
  ONBOARDING_TASKS,
  useInboxTasks,
  useTasksQuery,
  useCreateTaskMutation,
  useToggleTaskMutation,
  useDeleteTaskMutation,
} from '@/entities/task';
