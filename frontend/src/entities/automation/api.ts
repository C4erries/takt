import { request } from '@/shared/api';
import type { Automation, ToggleAutomationRequest } from './types';

export const automationsApi = {
  async fetchAutomations(workspaceId: string): Promise<Automation[]> {
    return request<Automation[]>(`/workspaces/${workspaceId}/automations`, {
      method: 'GET',
    });
  },

  async toggleAutomation(
    workspaceId: string,
    automationId: string,
    isEnabled: boolean
  ): Promise<Automation> {
    const payload: ToggleAutomationRequest = { is_enabled: isEnabled };
    return request<Automation>(
      `/workspaces/${workspaceId}/automations/${automationId}/toggle`,
      {
        method: 'PATCH',
        body: JSON.stringify(payload),
      }
    );
  },
};
