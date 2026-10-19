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
  taskId?: string | null;
  onSuccess?: (saved: Automation) => void;
  onClose: () => void;
}

export const TRIGGER_OPTIONS: { type: AutomationTriggerType; label: string; desc: string }[] = [
  { type: 'task_entered_column', label: 'Попадание в колонку', desc: 'Задача создана или перемещена в колонку' },
  { type: 'task_completed_in_column', label: 'Завершение в колонке', desc: 'Задача выполнена в колонке' },
  { type: 'task_field_changed', label: 'Изменение поля', desc: 'Смена приоритета, дедлайна или статуса' },
  { type: 'column_sla_timer', label: 'SLA таймер в колонке', desc: 'Задача в колонке дольше N часов' },
  { type: 'task_timer', label: 'Таймер задачи', desc: 'Через N часов выполнить действие' },
  { type: 'scheduled_sweep', label: 'Ежедневное расписание', desc: 'Регулярная утренняя проверка' },
];

export const CONDITION_FIELDS: { field: AutomationConditionField; label: string }[] = [
  { field: 'priority', label: 'Приоритет' },
  { field: 'due_date', label: 'Дедлайн' },
  { field: 'tags', label: 'Теги' },
  { field: 'time_in_column_hours', label: 'Время в колонке (ч)' },
  { field: 'is_completed', label: 'Завершена' },
  { field: 'due_date_status', label: 'Статус дедлайна' },
];

export const CONDITION_OPERATORS: { operator: AutomationConditionOperator; label: string }[] = [
  { operator: 'equals', label: 'равно' }, { operator: 'not_equals', label: 'не равно' },
  { operator: 'contains', label: 'содержит' }, { operator: 'not_contains', label: 'не содержит' },
  { operator: 'is_not_empty', label: 'заполнено' }, { operator: 'is_empty', label: 'пусто' },
  { operator: 'greater_than', label: 'больше чем' }, { operator: 'less_than', label: 'меньше чем' },
  { operator: 'in', label: 'один из' }, { operator: 'not_in', label: 'ни один из' },
];

export const FIELD_OPERATORS_MAP: Record<AutomationConditionField, AutomationConditionOperator[]> = {
  time_in_column_hours: ['greater_than', 'less_than', 'equals'],
  is_completed: ['equals', 'not_equals'],
  priority: ['equals', 'not_equals', 'in', 'not_in'],
  due_date_status: ['equals', 'not_equals', 'in'],
  due_date: ['equals', 'greater_than', 'less_than', 'is_not_empty', 'is_empty'],
  tags: ['contains', 'not_contains', 'in', 'is_not_empty', 'is_empty'],
};

export const ACTION_TYPES: { type: AutomationActionType; label: string; desc: string }[] = [
  { type: 'move_to_column', label: 'Переместить в колонку', desc: 'Перенос в целевую колонку' },
  { type: 'set_priority', label: 'Установить приоритет', desc: 'Изменить приоритет задачи' },
  { type: 'add_tag', label: 'Добавить тег', desc: 'Прикрепить метку' },
  { type: 'remove_tag', label: 'Удалить тег', desc: 'Снять метку' },
  { type: 'set_due_date', label: 'Назначить срок', desc: 'Установить дедлайн' },
  { type: 'archive_task', label: 'Архивировать задачу', desc: 'Убрать в архив' },
  { type: 'mark_completed', label: 'Отметить выполненной', desc: 'Установить галочку выполнения' },
  { type: 'mark_uncompleted', label: 'Снять выполнение', desc: 'Снять галочку выполнения' },
  { type: 'clear_due_date', label: 'Очистить дедлайн', desc: 'Сбросить дату дедлайна' },
];

export const PRIORITIES = [
  { value: 'pending', label: 'В ожидании' }, { value: 'high', label: 'Высокий' },
  { value: 'medium', label: 'Средний' }, { value: 'low', label: 'Низкий' }, { value: 'none', label: 'Обычный' },
];

export const PRIORITY_ACTIVE_STYLES: Record<string, string> = {
  pending: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  urgent: 'bg-[#4A6B82]/15 text-[#4A6B82] border-[#4A6B82]/40',
  high: 'bg-[#D97746]/15 text-[#D97746] border-[#D97746]/40',
  medium: 'bg-[#B8943F]/15 text-[#B8943F] border-[#B8943F]/40',
  low: 'bg-[#566B44]/15 text-[#566B44] border-[#566B44]/40',
  none: 'bg-surface text-primary border-subtle',
};

export const DUE_DATE_STATUS_OPTIONS = [
  { value: 'overdue', label: 'Просрочен' }, { value: 'due_soon', label: 'Горит (<= 48ч)' },
  { value: 'on_track', label: 'В графике' }, { value: 'no_due_date', label: 'Без срока' },
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
    case 'task_timer':
      return `Таймер задачи (> ${slaHours || 24} ч)`;
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

  if (condition.operator === 'is_empty' || condition.operator === 'is_not_empty') {
    return `${fieldName} ${opName}`;
  }

  let val = condition.value;
  if (condition.field === 'is_completed') {
    const isTrue = val === true || val === 'true';
    return condition.operator === 'equals'
      ? `Завершена: ${isTrue ? 'Да' : 'Нет'}`
      : `Завершена ${opName} ${isTrue ? 'Да' : 'Нет'}`;
  }

  if (condition.field === 'due_date_status' && typeof val === 'string') {
    const st = DUE_DATE_STATUS_OPTIONS.find((s) => s.value === val);
    val = st ? st.label : val;
  } else if (condition.field === 'priority') {
    if (Array.isArray(val)) {
      val = val.map((v) => PRIORITIES.find((p) => p.value === v)?.label || v).join(', ');
    } else if (typeof val === 'string' && val.includes(',')) {
      val = val.split(',').map((v) => PRIORITIES.find((p) => p.value === v.trim())?.label || v.trim()).join(', ');
    } else if (typeof val === 'string') {
      const p = PRIORITIES.find((item) => item.value === val);
      if (p) val = p.label;
    }
  } else if (Array.isArray(val)) {
    val = val.join(', ');
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
    case 'mark_completed':
      return 'Отметить задачу выполненной';
    case 'mark_uncompleted':
      return 'Снять отметку выполнения';
    case 'clear_due_date':
      return 'Очистить дедлайн';
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
  taskId,
  onSuccess,
  onClose,
}: UseRuleBuilderParams) {
  const createMutation = useCreateAutomationMutation(workspaceId);
  const updateMutation = useUpdateAutomationMutation(workspaceId);

  const [ruleName, setRuleName] = useState('');
  const [isCustomName, setIsCustomName] = useState(false);
  const [triggerType, setTriggerType] = useState<AutomationTriggerType>(
    taskId ? 'task_timer' : 'task_entered_column'
  );
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
    } else if (taskId) {
      setTriggerType('task_timer');
    } else if (prefilledColumnId) {
      setTriggerColumnId(prefilledColumnId);
      setTriggerType('task_entered_column');
    } else if (allColumns.length > 0 && !triggerColumnId) {
      setTriggerColumnId(allColumns[0].id);
    }
  }, [initialRule, prefilledColumnId, allColumns, taskId]);

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
  const handleRemoveCondition = (index: number) => setConditions((prev) => prev.filter((_, i) => i !== index));
  const handleUpdateCondition = (index: number, patch: Partial<AutomationCondition>) => {
    setConditions((prev) =>
      prev.map((c, i) => {
        if (i !== index) return c;
        const updated = { ...c, ...patch };
        if (patch.field && patch.field !== c.field) {
          const validOps = FIELD_OPERATORS_MAP[patch.field as AutomationConditionField] || ['equals'];
          if (!validOps.includes(updated.operator as AutomationConditionOperator)) {
            updated.operator = validOps[0];
          }
        }
        return updated;
      })
    );
  };

  const handleAddAction = () => {
    if (actions.length >= 5) return;
    const defaultTargetCol = allColumns.find((c) => c.id !== triggerColumnId)?.id ?? '';
    setActions((prev) => [...prev, { type: 'move_to_column', params: { target_column_id: defaultTargetCol } }]);
  };
  const handleRemoveAction = (index: number) => setActions((prev) => prev.filter((_, i) => i !== index));
  const handleUpdateAction = (index: number, patch: Partial<AutomationAction>) => {
    setActions((prev) => prev.map((a, i) => (i === index ? { ...a, ...patch } : a)));
  };

  const handleSave = async () => {
    if (!ruleName.trim() || actions.length === 0) return;
    const effectiveColumnId = prefilledColumnId || triggerColumnId || null;
    const triggerParams: Record<string, unknown> = {};
    if (
      triggerType === 'task_entered_column' ||
      triggerType === 'task_completed_in_column' ||
      triggerType === 'scheduled_sweep'
    ) {
      if (effectiveColumnId) triggerParams.column_id = effectiveColumnId;
    } else if (triggerType === 'column_sla_timer' || triggerType === 'task_timer') {
      if (effectiveColumnId) triggerParams.column_id = effectiveColumnId;
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
            columnId: effectiveColumnId,
            taskId: initialRule.taskId ?? taskId ?? null,
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
          columnId: effectiveColumnId,
          taskId: taskId ?? null,
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
    ruleName, setRuleName, setIsCustomName,
    triggerType, setTriggerType, triggerColumnId, setTriggerColumnId,
    triggerSlaHours, setTriggerSlaHours, triggerField, setTriggerField,
    conditions, logicalOperator, setLogicalOperator,
    handleAddCondition, handleRemoveCondition, handleUpdateCondition,
    actions, handleAddAction, handleRemoveAction, handleUpdateAction,
    handleSave, isSubmitting: createMutation.isPending || updateMutation.isPending,
  };
}
