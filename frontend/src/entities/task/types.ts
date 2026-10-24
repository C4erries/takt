import type {
  Priority,
  Task,
  Tag,
  CreateTagRequest,
  UpdateTaskRequest,
  SetFocusRequest,
  CreateInboxTaskRequest,
  UnarchiveTaskRequest,
  ArchiveCompletedColumnTasksResponse,
} from '@/shared/api';

export type {
  Priority,
  Task,
  Tag,
  CreateTagRequest,
  UpdateTaskRequest,
  SetFocusRequest,
  CreateInboxTaskRequest,
  UnarchiveTaskRequest,
  ArchiveCompletedColumnTasksResponse,
};

export interface UpdateTagRequest {
  colorHex?: string;
  name?: string;
}

export interface CreateTaskData {
  title: string;
  descriptionMarkdown?: string;
  priority?: Priority;
  dueDate?: string | null;
  parentId?: string | null;
  projectId?: string | null;
  columnId?: string | null;
}

export interface UpdateTaskData {
  title?: string;
  descriptionMarkdown?: string;
  dueDate?: string | null;
  priority?: Priority;
  focusOrder?: number | null;
  projectId?: string | null;
  columnId?: string | null;
  parentId?: string | null;
}
