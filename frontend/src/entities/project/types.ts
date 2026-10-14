export type ColumnStatusCategory = 'backlog' | 'todo' | 'in_progress' | 'done';

export interface Project {
  id: string;
  workspaceId: string;
  name: string;
  description: string;
  color: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string | null;
  color?: string | null;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string | null;
  color?: string | null;
}

export interface Board {
  id: string;
  projectId: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateBoardRequest {
  name: string;
}

export interface Column {
  id: string;
  boardId: string;
  name: string;
  position: number;
  statusCategory: ColumnStatusCategory;
  createdAt: string;
  updatedAt: string;
}

export interface CreateColumnRequest {
  name: string;
  position?: number | null;
  statusCategory?: ColumnStatusCategory;
}

export interface UpdateColumnRequest {
  name?: string;
  position?: number;
  statusCategory?: ColumnStatusCategory;
}

export interface MoveTaskRequest {
  columnId: string;
  position: number;
  projectId?: string | null;
}
