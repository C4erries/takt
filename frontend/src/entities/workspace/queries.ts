import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAuthToken, setAuthToken } from '@/shared/api';
import type {
  Workspace,
  CreateWorkspaceRequest,
  WorkspaceMember,
  WorkspaceInvite,
  InvitePreview,
  WorkspaceSwitchResponse,
} from './types';
import { workspaceApi } from './api';

export const WORKSPACES_QUERY_KEY = ['workspaces'];
export const ACTIVE_WORKSPACE_STORAGE_KEY = 'takt_active_workspace_id';

export const DEFAULT_WORKSPACES: Workspace[] = [
  {
    id: 'demo-workspace',
    name: 'Личное',
    role: 'owner',
    owner_id: 'demo-user',
    created_at: new Date().toISOString(),
  },
  {
    id: 'team-workspace',
    name: 'Команда Takt',
    role: 'member',
    owner_id: 'owner-user',
    created_at: new Date().toISOString(),
  },
];

export const DEFAULT_MEMBERS: WorkspaceMember[] = [
  {
    id: 'mem-1',
    userId: 'demo-user',
    user_id: 'demo-user',
    email: 'alex@takt.dev',
    displayName: 'Вы (Алекс)',
    display_name: 'Вы (Алекс)',
    role: 'owner',
    createdAt: new Date().toISOString(),
  },
  {
    id: 'mem-2',
    userId: 'user-2',
    user_id: 'user-2',
    email: 'elena@takt.dev',
    displayName: 'Елена',
    display_name: 'Елена',
    role: 'member',
    createdAt: new Date().toISOString(),
  },
];

export const useWorkspacesQuery = () => {
  return useQuery<Workspace[], Error>({
    queryKey: WORKSPACES_QUERY_KEY,
    queryFn: async () => {
      if (!getAuthToken()) {
        return DEFAULT_WORKSPACES;
      }
      try {
        const data = await workspaceApi.fetchWorkspaces();
        return data.length > 0 ? data : DEFAULT_WORKSPACES;
      } catch {
        return DEFAULT_WORKSPACES;
      }
    },
  });
};

export const useCreateWorkspaceMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateWorkspaceRequest): Promise<Workspace> => {
      if (!getAuthToken()) {
        const newWs: Workspace = {
          id: `ws-${Date.now()}`,
          name: data.name,
          role: 'owner',
          owner_id: 'demo-user',
          created_at: new Date().toISOString(),
        };
        return newWs;
      }
      try {
        return await workspaceApi.createWorkspace(data);
      } catch {
        const fallbackWs: Workspace = {
          id: `ws-${Date.now()}`,
          name: data.name,
          role: 'owner',
          owner_id: 'demo-user',
          created_at: new Date().toISOString(),
        };
        return fallbackWs;
      }
    },
    onSuccess: (newWorkspace) => {
      queryClient.setQueryData<Workspace[]>(WORKSPACES_QUERY_KEY, (old = []) => [
        ...old,
        newWorkspace,
      ]);
      localStorage.setItem(ACTIVE_WORKSPACE_STORAGE_KEY, newWorkspace.id);
      queryClient.invalidateQueries({ queryKey: WORKSPACES_QUERY_KEY });
    },
  });
};

export const useSwitchWorkspaceMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workspaceId: string): Promise<WorkspaceSwitchResponse> => {
      if (!getAuthToken()) {
        const workspaces = queryClient.getQueryData<Workspace[]>(WORKSPACES_QUERY_KEY) || DEFAULT_WORKSPACES;
        const target = workspaces.find((w) => w.id === workspaceId) || workspaces[0];
        return {
          token: 'demo-token',
          workspace: target,
        };
      }
      try {
        const res = await workspaceApi.switchWorkspace(workspaceId);
        if (res.token) {
          setAuthToken(res.token);
        }
        return res;
      } catch {
        const workspaces = queryClient.getQueryData<Workspace[]>(WORKSPACES_QUERY_KEY) || DEFAULT_WORKSPACES;
        const target = workspaces.find((w) => w.id === workspaceId) || workspaces[0];
        return {
          token: getAuthToken() || 'demo-token',
          workspace: target,
        };
      }
    },
    onSuccess: (res) => {
      localStorage.setItem(ACTIVE_WORKSPACE_STORAGE_KEY, res.workspace.id);
      queryClient.invalidateQueries();
    },
  });
};

export const useWorkspaceMembersQuery = (workspaceId?: string) => {
  return useQuery<WorkspaceMember[], Error>({
    queryKey: ['workspaces', workspaceId, 'members'],
    enabled: Boolean(workspaceId),
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) {
        return DEFAULT_MEMBERS;
      }
      try {
        const data = await workspaceApi.fetchWorkspaceMembers(workspaceId);
        return data.length > 0 ? data : DEFAULT_MEMBERS;
      } catch {
        return DEFAULT_MEMBERS;
      }
    },
  });
};

export const useRemoveMemberMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();
  const queryKey = ['workspaces', workspaceId, 'members'];

  return useMutation({
    mutationFn: async (userId: string) => {
      if (!workspaceId || !getAuthToken()) {
        return;
      }
      try {
        await workspaceApi.removeWorkspaceMember(workspaceId, userId);
      } catch {
        // Fallback for mock mode
      }
    },
    onMutate: async (userId: string) => {
      await queryClient.cancelQueries({ queryKey });
      const previousMembers = queryClient.getQueryData<WorkspaceMember[]>(queryKey);
      queryClient.setQueryData<WorkspaceMember[]>(queryKey, (old = []) =>
        old.filter((m) => (m.userId || m.user_id) !== userId && m.id !== userId)
      );
      return { previousMembers };
    },
    onError: (_err, _userId, context) => {
      if (context?.previousMembers) {
        queryClient.setQueryData(queryKey, context.previousMembers);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};

export const useCreateInviteMutation = (workspaceId?: string) => {
  return useMutation({
    mutationFn: async (): Promise<WorkspaceInvite> => {
      if (!workspaceId || !getAuthToken()) {
        const mockToken = `inv-${Math.random().toString(36).substring(2, 10)}`;
        return {
          id: `inv-${Date.now()}`,
          workspaceId: workspaceId || 'demo-workspace',
          token: mockToken,
          inviteUrl: `/invite/${mockToken}`,
          expiresAt: new Date(Date.now() + 7 * 86400000).toISOString(),
          createdAt: new Date().toISOString(),
        };
      }
      try {
        return await workspaceApi.createWorkspaceInvite(workspaceId);
      } catch {
        const mockToken = `inv-${Math.random().toString(36).substring(2, 10)}`;
        return {
          id: `inv-${Date.now()}`,
          workspaceId: workspaceId || 'demo-workspace',
          token: mockToken,
          inviteUrl: `/invite/${mockToken}`,
          expiresAt: new Date(Date.now() + 7 * 86400000).toISOString(),
          createdAt: new Date().toISOString(),
        };
      }
    },
  });
};

export const useInvitePreviewQuery = (token?: string | null) => {
  return useQuery<InvitePreview, Error>({
    queryKey: ['invites', token],
    enabled: Boolean(token),
    queryFn: async () => {
      if (!token) throw new Error('Token is required');
      try {
        return await workspaceApi.fetchInvitePreview(token);
      } catch {
        return {
          token,
          workspaceName: 'Команда Takt',
          workspace_name: 'Команда Takt',
          invitedByName: 'Павел',
          invited_by_name: 'Павел',
          expiresAt: new Date(Date.now() + 7 * 86400000).toISOString(),
        };
      }
    },
  });
};

export const useAcceptInviteMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (token: string): Promise<WorkspaceSwitchResponse> => {
      try {
        const res = await workspaceApi.acceptInvite(token);
        if (res.token) {
          setAuthToken(res.token);
        }
        return res;
      } catch {
        const fallback: WorkspaceSwitchResponse = {
          token: getAuthToken() || 'demo-token',
          workspace: {
            id: 'team-workspace',
            name: 'Команда Takt',
            role: 'member',
            owner_id: 'owner-user',
            created_at: new Date().toISOString(),
          },
        };
        return fallback;
      }
    },
    onSuccess: (res) => {
      localStorage.setItem(ACTIVE_WORKSPACE_STORAGE_KEY, res.workspace.id);
      queryClient.invalidateQueries();
    },
  });
};
