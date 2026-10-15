export type WorkspaceRole = 'owner' | 'member';

export interface Workspace {
  id: string;
  name: string;
  role: WorkspaceRole;
  ownerId?: string;
  owner_id?: string;
  type?: 'personal' | 'shared';
  createdAt?: string;
  created_at?: string;
}

export interface CreateWorkspaceRequest {
  name: string;
}

export interface UpdateWorkspaceRequest {
  name: string;
}

export interface WorkspaceMember {
  id: string;
  userId: string;
  user_id?: string;
  email: string;
  displayName: string;
  display_name?: string;
  role: WorkspaceRole;
  createdAt?: string;
  created_at?: string;
}

export interface WorkspaceInvite {
  id: string;
  workspaceId: string;
  workspace_id?: string;
  token: string;
  inviteUrl: string;
  invite_url?: string;
  expiresAt?: string | null;
  expires_at?: string | null;
  createdAt?: string;
  created_at?: string;
}

export interface InvitePreview {
  token: string;
  workspaceName: string;
  workspace_name?: string;
  invitedByName: string;
  invited_by_name?: string;
  expiresAt?: string | null;
  expires_at?: string | null;
}

export interface WorkspaceSwitchResponse {
  token: string;
  workspace: Workspace;
}
