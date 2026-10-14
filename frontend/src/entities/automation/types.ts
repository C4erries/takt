export interface Automation {
  id: string;
  workspaceId: string;
  boardId?: string | null;
  name: string;
  triggerType: string;
  conditionsJson: Record<string, unknown>;
  actionsJson: Record<string, unknown>;
  isEnabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ToggleAutomationRequest {
  is_enabled: boolean;
}
