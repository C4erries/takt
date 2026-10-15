import { request } from '@/shared/api';
import type {
  Workspace,
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
  WorkspaceMember,
  WorkspaceInvite,
  InvitePreview,
  WorkspaceSwitchResponse,
} from './types';

export const workspaceApi = {
  async fetchWorkspaces(): Promise<Workspace[]> {
    return request<Workspace[]>('/workspaces', { method: 'GET' });
  },

  async createWorkspace(data: CreateWorkspaceRequest): Promise<Workspace> {
    return request<Workspace>('/workspaces', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  async updateWorkspace(workspaceId: string, data: UpdateWorkspaceRequest): Promise<Workspace> {
    return request<Workspace>(`/workspaces/${workspaceId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  },

  async switchWorkspace(workspaceId: string): Promise<WorkspaceSwitchResponse> {
    return request<WorkspaceSwitchResponse>(`/workspaces/${workspaceId}/switch`, {
      method: 'POST',
    });
  },

  async fetchWorkspaceMembers(workspaceId: string): Promise<WorkspaceMember[]> {
    return request<WorkspaceMember[]>(`/workspaces/${workspaceId}/members`, {
      method: 'GET',
    });
  },

  async removeWorkspaceMember(workspaceId: string, userId: string): Promise<void> {
    await request<void>(`/workspaces/${workspaceId}/members/${userId}`, {
      method: 'DELETE',
    });
  },

  async createWorkspaceInvite(workspaceId: string): Promise<WorkspaceInvite> {
    return request<WorkspaceInvite>(`/workspaces/${workspaceId}/invites`, {
      method: 'POST',
    });
  },

  async fetchInvitePreview(token: string): Promise<InvitePreview> {
    return request<InvitePreview>(`/invites/${token}`, {
      method: 'GET',
    });
  },

  async acceptInvite(token: string): Promise<WorkspaceSwitchResponse> {
    return request<WorkspaceSwitchResponse>(`/invites/${token}/accept`, {
      method: 'POST',
    });
  },
};
