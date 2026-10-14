import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAuthToken } from '@/shared/api';
import type { Automation } from './types';
import { automationsApi } from './api';

export const DEFAULT_PRESET_AUTOMATIONS: Automation[] = [
  {
    id: 'auto-preset-done-archive',
    workspaceId: 'demo-workspace',
    name: 'Done & Archive',
    triggerType: 'task_completed',
    conditionsJson: {},
    actionsJson: { action: 'move_to_done' },
    isEnabled: true,
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
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const useAutomationsQuery = (workspaceId?: string) => {
  return useQuery<Automation[], Error>({
    queryKey: ['automations', workspaceId],
    queryFn: async () => {
      if (!workspaceId || !getAuthToken()) {
        return DEFAULT_PRESET_AUTOMATIONS;
      }
      try {
        const data = await automationsApi.fetchAutomations(workspaceId);
        return data.length > 0 ? data : DEFAULT_PRESET_AUTOMATIONS;
      } catch {
        return DEFAULT_PRESET_AUTOMATIONS;
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
      if (!workspaceId || !getAuthToken()) {
        // Return demo updated item
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
      return automationsApi.toggleAutomation(workspaceId, automationId, isEnabled);
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
