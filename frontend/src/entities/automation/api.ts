import { request } from '@/shared/api';
import type {
  Automation,
  CreateAutomationRequest,
  ToggleAutomationRequest,
  UpdateAutomationRequest,
} from './types';

export const automationsApi = {
  async fetchAutomations(workspaceId: string, taskId?: string): Promise<Automation[]> {
    const query = taskId ? `?taskId=${encodeURIComponent(taskId)}` : '';
    return request<Automation[]>(`/workspaces/${workspaceId}/automations${query}`, {
      method: 'GET',
    });
  },

  async createAutomation(
    workspaceId: string,
    payload: CreateAutomationRequest
  ): Promise<Automation> {
    return request<Automation>(`/workspaces/${workspaceId}/automations`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async updateAutomation(
    workspaceId: string,
    automationId: string,
    payload: UpdateAutomationRequest
  ): Promise<Automation> {
    return request<Automation>(`/workspaces/${workspaceId}/automations/${automationId}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
  },

  async deleteAutomation(
    workspaceId: string,
    automationId: string
  ): Promise<void> {
    return request<void>(`/workspaces/${workspaceId}/automations/${automationId}`, {
      method: 'DELETE',
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
