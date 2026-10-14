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
  type: 'personal' | 'shared';
  role: 'owner' | 'member';
  createdAt: string;
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
  createdAt: string;
  updatedAt: string;
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
  dueDate?: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown> | null;
}
