export type Priority = 'none' | 'low' | 'medium' | 'high' | 'pending' | 'urgent';

export interface User {
  id: string;
  email: string;
  displayName: string;
  createdAt: string;
}

export interface Workspace {
  id: string;
  name: string;
  type?: 'personal' | 'shared';
  role: 'owner' | 'member';
  ownerId?: string;
  owner_id?: string;
  createdAt?: string;
  created_at?: string;
}

export interface Tag {
  id: string;
  workspaceId: string;
  name: string;
  colorHex: string;
  createdAt: string;
}

export interface TagWithMetrics extends Tag {
  taskCount: number;
}

export interface CreateTagRequest {
  name: string;
  colorHex?: string;
}

export interface UpdateTagRequest {
  name?: string;
  colorHex?: string;
}

export interface Task {
  id: string;
  workspaceId: string;
  parentId?: string | null;
  subtasksCount?: number;
  subtasksCompletedCount?: number;
  subtasks?: Task[];
  projectId?: string | null;
  columnId?: string | null;
  title: string;
  descriptionMarkdown: string;
  isCompleted: boolean;
  completedAt?: string | null;
  dueDate?: string | null;
  priority: Priority;
  focusOrder?: number | null;
  position: number;
  tags?: Tag[];
  isArchived?: boolean;
  archivedAt?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface UnarchiveTaskRequest {
  columnId: string;
  boardId?: string | null;
}

export interface ArchiveCompletedColumnTasksResponse {
  archivedCount: number;
}

export interface TaskActivity {
  id: string;
  taskId: string;
  userId?: string | null;
  userName?: string | null;
  activityType: string;
  payload: Record<string, unknown>;
  createdAt: string;
}

export interface UpdateTaskRequest {
  title?: string;
  descriptionMarkdown?: string;
  dueDate?: string | null;
  priority?: Priority;
  focusOrder?: number | null;
  projectId?: string | null;
  columnId?: string | null;
  parentId?: string | null;
}

export interface SetFocusRequest {
  focusOrder?: number | null;
}

export interface AuthResponse {
  token: string;
  user: User;
  workspace: Workspace;
}

export interface MeResponse {
  user: User;
  workspace: Workspace;
}

export interface RegisterRequest {
  email: string;
  password: string;
  displayName: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface CreateInboxTaskRequest {
  title: string;
  descriptionMarkdown?: string;
  priority?: Priority;
  dueDate?: string | null;
  parentId?: string | null;
}

export interface Attachment {
  id: string;
  workspaceId: string;
  taskId: string;
  fileName: string;
  fileSize: number;
  contentType: string;
  contentUrl: string;
  createdAt: string;
}

export interface Project {
  id: string;
  workspaceId: string;
  name: string;
  description: string;
  color: string;
  createdAt: string;
  updatedAt: string;
}

export interface SearchResponse {
  tasks: Task[];
  projects: Project[];
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown> | null;
}
