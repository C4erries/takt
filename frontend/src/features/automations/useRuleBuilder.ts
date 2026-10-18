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
import { useCreateAutomationMutation } from '@/entities/automation';

export interface UseRuleBuilderParams {
  workspaceId?: string;
  prefilledColumnId?: string | null;
  prefilledColumnName?: string | null;
  allColumns?: { id: string; name: string }[];
  onSuccess?: (created: Automation) => void;
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

export function useRuleBuilder({
  workspaceId,
  prefilledColumnId,
  prefilledColumnName,
  allColumns = [],
  onSuccess,
  onClose,
}: UseRuleBuilderParams) {
  const createMutation = useCreateAutomationMutation(workspaceId);

  const [ruleName, setRuleName] = useState('');
  const [isCustomName, setIsCustomName] = useState(false);
  const [triggerType, setTriggerType] = useState<AutomationTriggerType>('task_entered_column');
  const [triggerColumnId, setTriggerColumnId] = useState<string>(prefilledColumnId || '');
  const [triggerSlaHours, setTriggerSlaHours] = useState<number>(24);
  const [triggerField, setTriggerField] = useState<string>('priority');

  const [conditions, setConditions] = useState<AutomationCondition[]>([]);
  const [actions, setActions] = useState<AutomationAction[]>([
    { type: 'set_priority', params: { priority: 'high' } },
  ]);

  useEffect(() => {
    if (prefilledColumnId) {
      setTriggerColumnId(prefilledColumnId);
      setTriggerType('task_entered_column');
    } else if (allColumns.length > 0 && !triggerColumnId) {
      setTriggerColumnId(allColumns[0].id);
    }
  }, [prefilledColumnId, allColumns]);

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
    const ruleDefinition: RuleDefinition = { trigger, conditions, actions };

    try {
      const created = await createMutation.mutateAsync({
        name: ruleName.trim(),
        triggerType,
        columnId: triggerColumnId || null,
        conditionsJson: { conditions },
        actionsJson: { actions },
        isEnabled: true,
        ruleDefinition,
      });
      if (onSuccess) onSuccess(created);
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
    handleAddCondition,
    handleRemoveCondition,
    handleUpdateCondition,
    actions,
    handleAddAction,
    handleRemoveAction,
    handleUpdateAction,
    handleSave,
    isSubmitting: createMutation.isPending,
  };
}
