import {
  AuthResponse,
  CreateInboxTaskRequest,
  LoginRequest,
  MeResponse,
  RegisterRequest,
  Task,
  Attachment,
  SearchResponse,
  ApiError,
} from './types';

const TOKEN_STORAGE_KEY = 'takt_jwt_token';

export const getAuthToken = (): string | null => {
  return localStorage.getItem(TOKEN_STORAGE_KEY);
};

export const setAuthToken = (token: string | null): void => {
  if (token) {
    localStorage.setItem(TOKEN_STORAGE_KEY, token);
  } else {
    localStorage.removeItem(TOKEN_STORAGE_KEY);
  }
};

const BASE_URL = '/api/v1';

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {});

  if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
    headers.set('Content-Type', 'application/json');
  }

  const token = getAuthToken();
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (!response.ok) {
    let errorData: ApiError | null = null;
    try {
      errorData = await response.json();
    } catch {
      // not a json response
    }

    const message = errorData?.message || `HTTP error ${response.status}: ${response.statusText}`;
    const err = new Error(message) as Error & { code?: string; details?: unknown };
    err.code = errorData?.code || 'UNKNOWN_ERROR';
    err.details = errorData?.details;
    throw err;
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

export const authApi = {
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const res = await request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    if (res.token) {
      setAuthToken(res.token);
    }
    return res;
  },

  async login(data: LoginRequest): Promise<AuthResponse> {
    const res = await request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    if (res.token) {
      setAuthToken(res.token);
    }
    return res;
  },

  async getMe(): Promise<MeResponse> {
    return request<MeResponse>('/auth/me', {
      method: 'GET',
    });
  },

  async logout(): Promise<{ message: string }> {
    try {
      const res = await request<{ message: string }>('/auth/logout', {
        method: 'POST',
      });
      setAuthToken(null);
      return res;
    } catch (err) {
      setAuthToken(null);
      throw err;
    }
  },
};

export const tasksApi = {
  async getInbox(): Promise<Task[]> {
    return request<Task[]>('/tasks/inbox', {
      method: 'GET',
    });
  },

  async createInbox(data: CreateInboxTaskRequest): Promise<Task> {
    return request<Task>('/tasks/inbox', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async toggle(id: string): Promise<Task> {
    return request<Task>(`/tasks/${id}/toggle`, {
      method: 'PATCH',
    });
  },

  async delete(id: string): Promise<void> {
    await request<void>(`/tasks/${id}`, {
      method: 'DELETE',
    });
  },

  async listSubtasks(workspaceId: string, taskId: string): Promise<Task[]> {
    return request<Task[]>(`/workspaces/${workspaceId}/tasks/${taskId}/subtasks`, {
      method: 'GET',
    });
  },
};

export const attachmentsApi = {
  async uploadTaskAttachment(workspaceId: string, taskId: string, file: File): Promise<Attachment> {
    const formData = new FormData();
    formData.append('file', file);
    return request<Attachment>(`/workspaces/${workspaceId}/tasks/${taskId}/attachments`, {
      method: 'POST',
      body: formData,
    });
  },

  async listTaskAttachments(workspaceId: string, taskId: string): Promise<Attachment[]> {
    return request<Attachment[]>(`/workspaces/${workspaceId}/tasks/${taskId}/attachments`, {
      method: 'GET',
    });
  },

  async deleteAttachment(workspaceId: string, attachmentId: string): Promise<void> {
    await request<void>(`/workspaces/${workspaceId}/attachments/${attachmentId}`, {
      method: 'DELETE',
    });
  },
};

export const uploadTaskAttachment = attachmentsApi.uploadTaskAttachment;
export const listTaskAttachments = attachmentsApi.listTaskAttachments;
export const deleteAttachment = attachmentsApi.deleteAttachment;

export const searchApi = {
  async searchWorkspace(workspaceId: string, q: string): Promise<SearchResponse> {
    const encoded = encodeURIComponent(q);
    return request<SearchResponse>(`/workspaces/${workspaceId}/search?q=${encoded}`, {
      method: 'GET',
    });
  },
};

export const searchWorkspace = searchApi.searchWorkspace;


