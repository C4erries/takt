export type Priority = 'none' | 'low' | 'medium' | 'high' | 'urgent';

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

export interface CreateTagRequest {
  name: string;
  colorHex?: string;
}

export interface Task {
  id: string;
  workspaceId: string;
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
  createdAt: string;
  updatedAt: string;
}

export interface UpdateTaskRequest {
  title?: string;
  descriptionMarkdown?: string;
  dueDate?: string | null;
  priority?: Priority;
  focusOrder?: number | null;
  projectId?: string | null;
  columnId?: string | null;
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
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown> | null;
}
