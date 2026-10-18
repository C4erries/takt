import { useState, useEffect } from 'react';
import type {
  Automation,
  AutomationAction,
  AutomationActionType,
  AutomationCondition,
  AutomationConditionField,
  AutomationConditionOperator,
  AutomationTrigger,
  AutomationTriggerType,
  RuleDefinition,
} from '@/entities/automation';
import {
  useCreateAutomationMutation,
  useUpdateAutomationMutation,
} from '@/entities/automation';

export interface UseRuleBuilderParams {
  workspaceId?: string;
  prefilledColumnId?: string | null;
  prefilledColumnName?: string | null;
  allColumns?: { id: string; name: string }[];
  mode?: 'create' | 'edit';
  initialRule?: Automation | null;
  onSuccess?: (saved: Automation) => void;
  onClose: () => void;
}

export const TRIGGER_OPTIONS: { type: AutomationTriggerType; label: string; desc: string }[] = [
  { type: 'task_entered_column', label: 'Попадание в колонку', desc: 'Задача создана или перемещена в колонку' },
  { type: 'task_completed_in_column', label: 'Завершение в колонке', desc: 'Задача выполнена в колонке' },
  { type: 'task_field_changed', label: 'Изменение поля', desc: 'Смена приоритета, дедлайна или статуса' },
  { type: 'column_sla_timer', label: 'SLA таймер в колонке', desc: 'Задача в колонке дольше N часов' },
  { type: 'scheduled_sweep', label: 'Ежедневное расписание', desc: 'Регулярная утренняя проверка' },
];

export const CONDITION_FIELDS: { field: AutomationConditionField; label: string }[] = [
  { field: 'priority', label: 'Приоритет' },
  { field: 'due_date', label: 'Дедлайн' },
  { field: 'tags', label: 'Теги' },
  { field: 'time_in_column_hours', label: 'Время в колонке (ч)' },
  { field: 'is_completed', label: 'Завершена' },
];

export const CONDITION_OPERATORS: { operator: AutomationConditionOperator; label: string }[] = [
  { operator: 'equals', label: 'равно' },
  { operator: 'not_equals', label: 'не равно' },
  { operator: 'contains', label: 'содержит' },
  { operator: 'not_contains', label: 'не содержит' },
  { operator: 'is_not_empty', label: 'заполнено' },
  { operator: 'is_empty', label: 'пусто' },
  { operator: 'greater_than', label: 'больше чем' },
  { operator: 'less_than', label: 'меньше чем' },
];

export const ACTION_TYPES: { type: AutomationActionType; label: string; desc: string }[] = [
  { type: 'move_to_column', label: 'Переместить в колонку', desc: 'Перенос в целевую колонку' },
  { type: 'set_priority', label: 'Установить приоритет', desc: 'Изменить приоритет задачи' },
  { type: 'add_tag', label: 'Добавить тег', desc: 'Прикрепить метку' },
  { type: 'remove_tag', label: 'Удалить тег', desc: 'Снять метку' },
  { type: 'set_due_date', label: 'Назначить срок', desc: 'Установить дедлайн' },
  { type: 'archive_task', label: 'Архивировать задачу', desc: 'Убрать в архив' },
];

export const PRIORITIES = [
  { value: 'urgent', label: 'Срочный' },
  { value: 'high', label: 'Высокий' },
  { value: 'medium', label: 'Средний' },
  { value: 'low', label: 'Низкий' },
  { value: 'none', label: 'Обычный' },
];

export function formatTriggerLabel(
  triggerType: string,
  columnId?: string | null,
  allColumns: { id: string; name: string }[] = [],
  params?: Record<string, unknown>
): string {
  const colName =
    allColumns.find((c) => c.id === columnId)?.name ||
    (params?.column_id ? allColumns.find((c) => c.id === params.column_id)?.name : undefined);
  const slaHours = params?.duration_hours;

  switch (triggerType) {
    case 'task_entered_column':
      return colName ? `При попадании в [${colName}]` : 'При попадании в колонку';
    case 'task_completed_in_column':
      return colName ? `При завершении в [${colName}]` : 'При завершении в колонке';
    case 'task_field_changed':
      return 'При изменении поля задачи';
    case 'column_sla_timer':
      return colName
        ? `SLA таймер в [${colName}] (> ${slaHours || 24} ч)`
        : `SLA таймер (> ${slaHours || 24} ч)`;
    case 'scheduled_sweep':
      return 'Ежедневная проверка по расписанию';
    case 'task_completed':
      return 'При завершении задачи';
    case 'task_updated':
      return 'При обновлении задачи';
    case 'task_created':
      return 'При создании задачи';
    case 'due_date_reached':
      return 'При наступлении дедлайна';
    default:
      return triggerType;
  }
}

export function formatConditionLabel(condition: AutomationCondition): string {
  const fieldObj = CONDITION_FIELDS.find((f) => f.field === condition.field);
  const fieldName = fieldObj?.label || String(condition.field);
  const opObj = CONDITION_OPERATORS.find((o) => o.operator === condition.operator);
  const opName = opObj?.label || String(condition.operator);

  let val = condition.value;
  if (condition.field === 'priority' && typeof val === 'string') {
    const p = PRIORITIES.find((item) => item.value === val);
    if (p) val = p.label;
  }
  if (condition.operator === 'is_empty' || condition.operator === 'is_not_empty') {
    return `${fieldName} ${opName}`;
  }
  return `${fieldName} ${opName} ${val !== undefined && val !== '' ? `«${val}»` : ''}`.trim();
}

export function formatActionLabel(
  action: AutomationAction,
  allColumns: { id: string; name: string }[] = []
): string {
  const params = action.params || {};
  switch (action.type) {
    case 'move_to_column': {
      const targetCol = allColumns.find((c) => c.id === params.target_column_id)?.name;
      return targetCol ? `Переместить в [${targetCol}]` : 'Переместить в колонку';
    }
    case 'set_priority': {
      const p = PRIORITIES.find((item) => item.value === params.priority);
      return `Установить приоритет: ${p?.label || params.priority || 'Высокий'}`;
    }
    case 'add_tag':
      return `Добавить тег: «${params.tag_name || params.tag || ''}»`;
    case 'remove_tag':
      return `Удалить тег: «${params.tag_name || params.tag || ''}»`;
    case 'set_due_date': {
      const presetMap: Record<string, string> = {
        today: 'сегодня',
        tomorrow: 'завтра',
        next_week: 'следующая неделя',
      };
      const presetStr = String(params.preset || '');
      return `Назначить срок: ${presetMap[presetStr] || presetStr || 'срок'}`;
    }
    case 'archive_task':
      return 'Архивировать задачу';
    default:
      return ACTION_TYPES.find((a) => a.type === action.type)?.label || String(action.type);
  }
}

export function useRuleBuilder({
  workspaceId,
  prefilledColumnId,
  prefilledColumnName,
  allColumns = [],
  mode = 'create',
  initialRule,
  onSuccess,
  onClose,
}: UseRuleBuilderParams) {
  const createMutation = useCreateAutomationMutation(workspaceId);
  const updateMutation = useUpdateAutomationMutation(workspaceId);

  const [ruleName, setRuleName] = useState('');
  const [isCustomName, setIsCustomName] = useState(false);
  const [triggerType, setTriggerType] = useState<AutomationTriggerType>('task_entered_column');
  const [triggerColumnId, setTriggerColumnId] = useState<string>(prefilledColumnId || '');
  const [triggerSlaHours, setTriggerSlaHours] = useState<number>(24);
  const [triggerField, setTriggerField] = useState<string>('priority');

  const [conditions, setConditions] = useState<AutomationCondition[]>([]);
  const [logicalOperator, setLogicalOperator] = useState<'and' | 'or'>('and');
  const [actions, setActions] = useState<AutomationAction[]>([
    { type: 'set_priority', params: { priority: 'high' } },
  ]);

  useEffect(() => {
    if (initialRule) {
      setRuleName(initialRule.name);
      setIsCustomName(true);
      const trig = initialRule.ruleDefinition?.trigger;
      const trigType = (trig?.type || initialRule.triggerType || 'task_entered_column') as AutomationTriggerType;
      setTriggerType(trigType);
      const colId = (trig?.params?.column_id as string) || initialRule.columnId || prefilledColumnId || '';
      setTriggerColumnId(colId);
      const sla = Number(trig?.params?.duration_hours);
      if (sla) setTriggerSlaHours(sla);
      const fld = trig?.params?.field as string;
      if (fld) setTriggerField(fld);
      const conds =
        initialRule.ruleDefinition?.conditions ||
        ((initialRule.conditionsJson as { conditions?: AutomationCondition[] })?.conditions) ||
        [];
      setConditions(conds);
      const op =
        initialRule.ruleDefinition?.logicalOperator ||
        ((initialRule.conditionsJson as { logical_operator?: 'and' | 'or' })?.logical_operator) ||
        'and';
      setLogicalOperator(op);
      const acts =
        initialRule.ruleDefinition?.actions ||
        ((initialRule.actionsJson as { actions?: AutomationAction[] })?.actions) ||
        [];
      setActions(acts.length > 0 ? acts : [{ type: 'set_priority', params: { priority: 'high' } }]);
    } else if (prefilledColumnId) {
      setTriggerColumnId(prefilledColumnId);
      setTriggerType('task_entered_column');
    } else if (allColumns.length > 0 && !triggerColumnId) {
      setTriggerColumnId(allColumns[0].id);
    }
  }, [initialRule, prefilledColumnId, allColumns]);

  useEffect(() => {
    if (!isCustomName) {
      const triggerLabel = TRIGGER_OPTIONS.find((t) => t.type === triggerType)?.label ?? triggerType;
      const col = allColumns.find((c) => c.id === triggerColumnId)?.name ?? prefilledColumnName ?? '';
      const actionLabel = ACTION_TYPES.find((a) => a.type === actions[0]?.type)?.label ?? 'Действие';
      const name = col ? `${triggerLabel} [${col}] → ${actionLabel}` : `${triggerLabel} → ${actionLabel}`;
      setRuleName(name);
    }
  }, [triggerType, triggerColumnId, actions, isCustomName, allColumns, prefilledColumnName]);

  const handleAddCondition = () => {
    if (conditions.length >= 5) return;
    setConditions((prev) => [...prev, { field: 'priority', operator: 'equals', value: 'high' }]);
  };

  const handleRemoveCondition = (index: number) => {
    setConditions((prev) => prev.filter((_, i) => i !== index));
  };

  const handleUpdateCondition = (index: number, patch: Partial<AutomationCondition>) => {
    setConditions((prev) => prev.map((c, i) => (i === index ? { ...c, ...patch } : c)));
  };

  const handleAddAction = () => {
    if (actions.length >= 5) return;
    const defaultTargetCol = allColumns.find((c) => c.id !== triggerColumnId)?.id ?? '';
    setActions((prev) => [
      ...prev,
      { type: 'move_to_column', params: { target_column_id: defaultTargetCol } },
    ]);
  };

  const handleRemoveAction = (index: number) => {
    setActions((prev) => prev.filter((_, i) => i !== index));
  };

  const handleUpdateAction = (index: number, patch: Partial<AutomationAction>) => {
    setActions((prev) => prev.map((a, i) => (i === index ? { ...a, ...patch } : a)));
  };

  const handleSave = async () => {
    if (!ruleName.trim() || actions.length === 0) return;

    const triggerParams: Record<string, unknown> = {};
    if (
      triggerType === 'task_entered_column' ||
      triggerType === 'task_completed_in_column' ||
      triggerType === 'scheduled_sweep'
    ) {
      if (triggerColumnId) triggerParams.column_id = triggerColumnId;
    } else if (triggerType === 'column_sla_timer') {
      if (triggerColumnId) triggerParams.column_id = triggerColumnId;
      triggerParams.duration_hours = Number(triggerSlaHours) || 24;
    } else if (triggerType === 'task_field_changed') {
      triggerParams.field = triggerField;
    }

    const trigger: AutomationTrigger = { type: triggerType, params: triggerParams };
    const ruleDefinition: RuleDefinition = { trigger, conditions, logicalOperator, actions };

    try {
      if (mode === 'edit' && initialRule?.id) {
        const updated = await updateMutation.mutateAsync({
          automationId: initialRule.id,
          data: {
            name: ruleName.trim(),
            triggerType,
            columnId: triggerColumnId || null,
            conditionsJson: { conditions, logical_operator: logicalOperator },
            actionsJson: { actions },
            ruleDefinition,
          },
        });
        if (onSuccess) onSuccess(updated);
      } else {
        const created = await createMutation.mutateAsync({
          name: ruleName.trim(),
          triggerType,
          columnId: triggerColumnId || null,
          conditionsJson: { conditions, logical_operator: logicalOperator },
          actionsJson: { actions },
          isEnabled: true,
          ruleDefinition,
        });
        if (onSuccess) onSuccess(created);
      }
      onClose();
    } catch {
      // Handled in mutation
    }
  };

  return {
    ruleName,
    setRuleName,
    setIsCustomName,
    triggerType,
    setTriggerType,
    triggerColumnId,
    setTriggerColumnId,
    triggerSlaHours,
    setTriggerSlaHours,
    triggerField,
    setTriggerField,
    conditions,
    logicalOperator,
    setLogicalOperator,
    handleAddCondition,
    handleRemoveCondition,
    handleUpdateCondition,
    actions,
    handleAddAction,
    handleRemoveAction,
    handleUpdateAction,
    handleSave,
    isSubmitting: createMutation.isPending || updateMutation.isPending,
  };
}
