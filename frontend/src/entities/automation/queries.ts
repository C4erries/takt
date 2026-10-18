import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import type { Automation, CreateAutomationRequest } from './types';
import { automationsApi } from './api';

const CUSTOM_AUTOMATIONS_STORAGE_KEY = 'takt_custom_automations';

function getStoredCustomAutomations(): Automation[] {
  try {
    const raw = localStorage.getItem(CUSTOM_AUTOMATIONS_STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveStoredCustomAutomations(automations: Automation[]): void {
  try {
    localStorage.setItem(CUSTOM_AUTOMATIONS_STORAGE_KEY, JSON.stringify(automations));
  } catch {
    // ignore
  }
}

export const DEFAULT_PRESET_AUTOMATIONS: Automation[] = [
  {
    id: 'auto-preset-done-archive',
    workspaceId: 'demo-workspace',
    name: 'Done & Archive',
    triggerType: 'task_completed',
    conditionsJson: {},
    actionsJson: { action: 'move_to_done' },
    isEnabled: true,
    ruleDefinition: {
      trigger: { type: 'task_completed', params: {} },
      conditions: [],
      actions: [{ type: 'archive_task', params: {} }],
    },
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'auto-preset-auto-focus',
    workspaceId: 'demo-workspace',
    name: 'Auto-Focus',
    triggerType: 'task_updated',
    conditionsJson: { due_date: '<= today' },
    actionsJson: { action: 'set_focus_order' },
    isEnabled: true,
    ruleDefinition: {
      trigger: { type: 'task_field_changed', params: { field: 'due_date' } },
      conditions: [{ field: 'due_date', operator: 'equals', value: 'today' }],
      actions: [{ type: 'set_priority', params: { priority: 'high' } }],
    },
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 'auto-preset-inbox-zero',
    workspaceId: 'demo-workspace',
    name: 'Inbox Zero',
    triggerType: 'task_updated',
    conditionsJson: { has_project: true, no_column: true },
    actionsJson: { action: 'move_to_first_column' },
    isEnabled: false,
    ruleDefinition: {
      trigger: { type: 'task_entered_column', params: {} },
      conditions: [{ field: 'priority', operator: 'not_equals', value: 'none' }],
      actions: [{ type: 'move_to_column', params: {} }],
    },
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const useAutomationsQuery = (workspaceId?: string) => {
  return useQuery<Automation[], Error>({
    queryKey: ['automations', workspaceId],
    queryFn: async () => {
      const customLocal = getStoredCustomAutomations();
      if (!workspaceId || !getAuthToken()) {
        return [...DEFAULT_PRESET_AUTOMATIONS, ...customLocal];
      }
      try {
        const data = await automationsApi.fetchAutomations(workspaceId);
        const combined = data.length > 0 ? data : DEFAULT_PRESET_AUTOMATIONS;
        return [...combined, ...customLocal.filter((c) => !combined.some((item) => item.id === c.id))];
      } catch {
        return [...DEFAULT_PRESET_AUTOMATIONS, ...customLocal];
      }
    },
  });
};

export interface ToggleAutomationVariables {
  automationId: string;
  isEnabled: boolean;
}

export const useToggleAutomationMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ automationId, isEnabled }: ToggleAutomationVariables) => {
      const customLocal = getStoredCustomAutomations();
      const customIdx = customLocal.findIndex((a) => a.id === automationId);
      if (customIdx >= 0) {
        customLocal[customIdx].isEnabled = isEnabled;
        saveStoredCustomAutomations(customLocal);
      }

      if (!workspaceId || !getAuthToken()) {
        return {
          id: automationId,
          workspaceId: workspaceId || 'demo-workspace',
          name: '',
          triggerType: '',
          conditionsJson: {},
          actionsJson: {},
          isEnabled,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        } as Automation;
      }
      try {
        return await automationsApi.toggleAutomation(workspaceId, automationId, isEnabled);
      } catch {
        // Fallback gracefully for local custom automations
        return {
          id: automationId,
          workspaceId,
          name: '',
          triggerType: '',
          conditionsJson: {},
          actionsJson: {},
          isEnabled,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        } as Automation;
      }
    },
    onMutate: async ({ automationId, isEnabled }) => {
      const queryKey = ['automations', workspaceId];
      await queryClient.cancelQueries({ queryKey });
      const previousAutomations = queryClient.getQueryData<Automation[]>(queryKey);

      queryClient.setQueryData<Automation[]>(queryKey, (old = []) =>
        old.map((item) =>
          item.id === automationId ? { ...item, isEnabled } : item
        )
      );

      return { previousAutomations };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousAutomations) {
        queryClient.setQueryData(['automations', workspaceId], context.previousAutomations);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['automations', workspaceId] });
    },
  });
};

export const useCreateAutomationMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: CreateAutomationRequest) => {
      const localId = `auto-custom-${Date.now()}`;
      const newRule: Automation = {
        id: localId,
        workspaceId: workspaceId || 'demo-workspace',
        boardId: payload.boardId ?? null,
        columnId: payload.columnId ?? null,
        name: payload.name,
        triggerType: payload.triggerType,
        conditionsJson: payload.conditionsJson ?? {},
        actionsJson: payload.actionsJson ?? {},
        isEnabled: payload.isEnabled ?? true,
        ruleDefinition: payload.ruleDefinition ?? null,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      if (!workspaceId || !getAuthToken()) {
        const customLocal = getStoredCustomAutomations();
        saveStoredCustomAutomations([...customLocal, newRule]);
        return newRule;
      }

      try {
        return await automationsApi.createAutomation(workspaceId, payload);
      } catch {
        // Fallback to storing locally so user action isn't lost
        const customLocal = getStoredCustomAutomations();
        saveStoredCustomAutomations([...customLocal, newRule]);
        return newRule;
      }
    },
    onMutate: async (newRuleReq) => {
      const queryKey = ['automations', workspaceId];
      await queryClient.cancelQueries({ queryKey });
      const previousAutomations = queryClient.getQueryData<Automation[]>(queryKey);

      const optimisticRule: Automation = {
        id: `optimistic-${Date.now()}`,
        workspaceId: workspaceId || 'demo-workspace',
        boardId: newRuleReq.boardId ?? null,
        columnId: newRuleReq.columnId ?? null,
        name: newRuleReq.name,
        triggerType: newRuleReq.triggerType,
        conditionsJson: newRuleReq.conditionsJson ?? {},
        actionsJson: newRuleReq.actionsJson ?? {},
        isEnabled: newRuleReq.isEnabled ?? true,
        ruleDefinition: newRuleReq.ruleDefinition ?? null,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      queryClient.setQueryData<Automation[]>(queryKey, (old = []) => [
        ...old,
        optimisticRule,
      ]);

      return { previousAutomations };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousAutomations) {
        queryClient.setQueryData(['automations', workspaceId], context.previousAutomations);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['automations', workspaceId] });
    },
  });
};

export const useDeleteAutomationMutation = (workspaceId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (automationId: string) => {
      const customLocal = getStoredCustomAutomations();
      saveStoredCustomAutomations(customLocal.filter((a) => a.id !== automationId));

      if (!workspaceId || !getAuthToken()) {
        return;
      }

      try {
        await automationsApi.deleteAutomation(workspaceId, automationId);
      } catch {
        // Silently handled for local/unwired rules
      }
    },
    onMutate: async (automationId) => {
      const queryKey = ['automations', workspaceId];
      await queryClient.cancelQueries({ queryKey });
      const previousAutomations = queryClient.getQueryData<Automation[]>(queryKey);

      queryClient.setQueryData<Automation[]>(queryKey, (old = []) =>
        old.filter((item) => item.id !== automationId)
      );

      return { previousAutomations };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousAutomations) {
        queryClient.setQueryData(['automations', workspaceId], context.previousAutomations);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['automations', workspaceId] });
    },
  });
};
