export type AutomationTriggerType =
  | 'task_entered_column'
  | 'task_completed_in_column'
  | 'task_field_changed'
  | 'column_sla_timer'
  | 'scheduled_sweep'
  | 'task_completed'
  | 'task_updated'
  | 'task_created'
  | 'due_date_reached'
  | 'task_timer';

export interface AutomationTrigger {
  type: AutomationTriggerType | string;
  params?: Record<string, unknown>;
}

export type AutomationConditionField =
  | 'priority'
  | 'is_completed'
  | 'due_date'
  | 'time_in_column_hours'
  | 'tags'
  | 'due_date_status';

export type AutomationConditionOperator =
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'not_contains'
  | 'is_empty'
  | 'is_not_empty'
  | 'greater_than'
  | 'less_than'
  | 'in'
  | 'not_in';

export interface AutomationCondition {
  field: AutomationConditionField | string;
  operator: AutomationConditionOperator | string;
  value?: unknown;
}

export type AutomationActionType =
  | 'move_to_column'
  | 'set_priority'
  | 'add_tag'
  | 'remove_tag'
  | 'set_due_date'
  | 'archive_task'
  | 'mark_completed'
  | 'mark_uncompleted'
  | 'clear_due_date';

export interface AutomationAction {
  type: AutomationActionType | string;
  params?: Record<string, unknown>;
}

export interface RuleDefinition {
  trigger: AutomationTrigger;
  conditions?: AutomationCondition[];
  logicalOperator?: 'and' | 'or';
  actions: AutomationAction[];
}

export interface Automation {
  id: string;
  workspaceId: string;
  boardId?: string | null;
  columnId?: string | null;
  taskId?: string | null;
  name: string;
  triggerType: AutomationTriggerType | string;
  conditionsJson: Record<string, unknown>;
  actionsJson: Record<string, unknown>;
  isEnabled: boolean;
  ruleDefinition?: RuleDefinition | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAutomationRequest {
  name: string;
  triggerType: AutomationTriggerType | string;
  boardId?: string | null;
  columnId?: string | null;
  taskId?: string | null;
  conditionsJson?: Record<string, unknown>;
  actionsJson?: Record<string, unknown>;
  isEnabled?: boolean;
  ruleDefinition?: RuleDefinition | null;
}

export interface UpdateAutomationRequest {
  name?: string;
  triggerType?: AutomationTriggerType | string;
  boardId?: string | null;
  columnId?: string | null;
  taskId?: string | null;
  conditionsJson?: Record<string, unknown>;
  actionsJson?: Record<string, unknown>;
  isEnabled?: boolean;
  ruleDefinition?: RuleDefinition | null;
}

export interface ToggleAutomationRequest {
  is_enabled: boolean;
}
