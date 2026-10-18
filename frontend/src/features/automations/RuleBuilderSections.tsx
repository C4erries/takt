import React from 'react';
import { Plus, Trash2 } from 'lucide-react';
import { Button, Input, Badge } from '@/shared/components/ui';
import type {
  AutomationAction,
  AutomationActionType,
  AutomationCondition,
  AutomationConditionField,
  AutomationConditionOperator,
  AutomationTriggerType,
} from '@/entities/automation';
import {
  TRIGGER_OPTIONS,
  CONDITION_FIELDS,
  CONDITION_OPERATORS,
  ACTION_TYPES,
  PRIORITIES,
} from './useRuleBuilder';

export interface TriggerSectionProps {
  triggerType: AutomationTriggerType;
  setTriggerType: (t: AutomationTriggerType) => void;
  triggerColumnId: string;
  setTriggerColumnId: (id: string) => void;
  triggerSlaHours: number;
  setTriggerSlaHours: (h: number) => void;
  triggerField: string;
  setTriggerField: (f: string) => void;
  allColumns: { id: string; name: string }[];
}

export const TriggerSection: React.FC<TriggerSectionProps> = ({
  triggerType,
  setTriggerType,
  triggerColumnId,
  setTriggerColumnId,
  triggerSlaHours,
  setTriggerSlaHours,
  triggerField,
  setTriggerField,
  allColumns,
}) => {
  const isColTrigger =
    triggerType === 'task_entered_column' ||
    triggerType === 'task_completed_in_column' ||
    triggerType === 'column_sla_timer' ||
    triggerType === 'scheduled_sweep';

  return (
    <div className="p-3 rounded-xl border border-subtle bg-surface/60 space-y-2.5">
      <div className="flex items-center gap-2">
        <span className="w-5 h-5 rounded-full bg-accent-solid text-white text-[10px] font-bold flex items-center justify-center shrink-0">
          1
        </span>
        <span className="text-xs font-semibold uppercase tracking-wider text-primary">
          КОГДА (Триггер события)
        </span>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        <div className="space-y-1">
          <label className="text-[10px] font-mono text-muted">Событие:</label>
          <select
            value={triggerType}
            onChange={(e) => setTriggerType(e.target.value as AutomationTriggerType)}
            className="w-full text-xs bg-surface border border-subtle rounded-lg px-2.5 py-1.5 text-primary focus:outline-none focus:border-accent-solid"
          >
            {TRIGGER_OPTIONS.map((t) => (
              <option key={t.type} value={t.type}>{t.label}</option>
            ))}
          </select>
        </div>
        {isColTrigger && (
          <div className="space-y-1">
            <label className="text-[10px] font-mono text-muted">В колонке:</label>
            <select
              value={triggerColumnId}
              onChange={(e) => setTriggerColumnId(e.target.value)}
              className="w-full text-xs bg-surface border border-subtle rounded-lg px-2.5 py-1.5 text-primary focus:outline-none focus:border-accent-solid"
            >
              {allColumns.length === 0 && <option value="">Любая колонка</option>}
              {allColumns.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>
        )}
        {triggerType === 'column_sla_timer' && (
          <div className="space-y-1">
            <label className="text-[10px] font-mono text-muted">Порог нахождения:</label>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                min={1}
                max={720}
                value={triggerSlaHours}
                onChange={(e) => setTriggerSlaHours(Number(e.target.value) || 24)}
                className="text-xs bg-surface border-subtle h-8"
              />
              <span className="text-xs text-muted">часов</span>
            </div>
          </div>
        )}
        {triggerType === 'task_field_changed' && (
          <div className="space-y-1">
            <label className="text-[10px] font-mono text-muted">Поле задачи:</label>
            <select
              value={triggerField}
              onChange={(e) => setTriggerField(e.target.value)}
              className="w-full text-xs bg-surface border border-subtle rounded-lg px-2.5 py-1.5 text-primary focus:outline-none focus:border-accent-solid"
            >
              <option value="priority">Приоритет</option>
              <option value="due_date">Дедлайн</option>
              <option value="status">Статус</option>
            </select>
          </div>
        )}
      </div>
    </div>
  );
};

export interface ConditionsSectionProps {
  conditions: AutomationCondition[];
  logicalOperator?: 'and' | 'or';
  onLogicalOperatorChange?: (op: 'and' | 'or') => void;
  onAdd: () => void;
  onRemove: (idx: number) => void;
  onUpdate: (idx: number, patch: Partial<AutomationCondition>) => void;
}

export const ConditionsSection: React.FC<ConditionsSectionProps> = ({
  conditions,
  logicalOperator = 'and',
  onLogicalOperatorChange,
  onAdd,
  onRemove,
  onUpdate,
}) => (
  <div className="p-3 rounded-xl border border-subtle bg-surface/60 space-y-2.5">
    <div className="flex items-center justify-between flex-wrap gap-2">
      <div className="flex items-center gap-2">
        <span className="w-5 h-5 rounded-full bg-accent-surface text-primary text-[10px] font-bold flex items-center justify-center shrink-0 border border-subtle">
          2
        </span>
        <span className="text-xs font-semibold uppercase tracking-wider text-primary">
          ЕСЛИ (Условия {logicalOperator === 'or' ? 'ИЛИ' : 'И'})
        </span>
        <Badge variant="subtle" size="sm" className="font-mono text-[10px]">{conditions.length}/5</Badge>
      </div>

      <div className="flex items-center gap-2">
        {conditions.length > 1 && onLogicalOperatorChange && (
          <div className="flex items-center bg-canvas p-0.5 rounded-lg border border-subtle">
            <button
              type="button"
              onClick={() => onLogicalOperatorChange('and')}
              className={`px-2 py-0.5 rounded text-[10px] font-medium transition-colors ${
                logicalOperator === 'and' ? 'bg-accent-solid text-white' : 'text-muted hover:text-primary'
              }`}
            >
              Все (И)
            </button>
            <button
              type="button"
              onClick={() => onLogicalOperatorChange('or')}
              className={`px-2 py-0.5 rounded text-[10px] font-medium transition-colors ${
                logicalOperator === 'or' ? 'bg-accent-solid text-white' : 'text-muted hover:text-primary'
              }`}
            >
              Хотя бы одно (ИЛИ)
            </button>
          </div>
        )}
        {conditions.length < 5 && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onAdd}
            leftIcon={<Plus className="w-3.5 h-3.5" />}
            className="text-[11px] h-7 px-2"
          >
            Добавить
          </Button>
        )}
      </div>
    </div>

    {conditions.length === 0 ? (
      <p className="text-xs text-muted italic px-2 py-1 bg-canvas/40 rounded-lg border border-subtle/40">
        Действие выполняется всегда без дополнительных условий.
      </p>
    ) : (
      <div className="space-y-2">
        {conditions.map((cond, index) => (
          <div key={index} className="flex items-center gap-2 p-2 rounded-lg bg-surface border border-subtle text-xs">
            <span className="text-[10px] font-mono text-muted shrink-0 w-8">
              {index === 0 ? 'ЕСЛИ' : (logicalOperator === 'or' ? 'ИЛИ' : 'И')}
            </span>
            <select
              value={cond.field}
              onChange={(e) => onUpdate(index, { field: e.target.value as AutomationConditionField })}
              className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary"
            >
              {CONDITION_FIELDS.map((f) => (
                <option key={f.field} value={f.field}>{f.label}</option>
              ))}
            </select>
            <select
              value={cond.operator}
              onChange={(e) => onUpdate(index, { operator: e.target.value as AutomationConditionOperator })}
              className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary"
            >
              {CONDITION_OPERATORS.map((op) => (
                <option key={op.operator} value={op.operator}>{op.label}</option>
              ))}
            </select>
            {cond.operator !== 'is_empty' && cond.operator !== 'is_not_empty' && (
              cond.field === 'priority' ? (
                <select
                  value={String(cond.value || 'high')}
                  onChange={(e) => onUpdate(index, { value: e.target.value })}
                  className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary flex-1"
                >
                  {PRIORITIES.map((p) => (
                    <option key={p.value} value={p.value}>{p.label}</option>
                  ))}
                </select>
              ) : (
                <input
                  type="text"
                  value={String(cond.value || '')}
                  onChange={(e) => onUpdate(index, { value: e.target.value })}
                  placeholder="Значение..."
                  className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary flex-1 focus:outline-none"
                />
              )
            )}
            <button
              type="button"
              onClick={() => onRemove(index)}
              className="text-muted hover:text-primary p-1 rounded transition-colors shrink-0"
              title="Удалить условие"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          </div>
        ))}
      </div>
    )}
  </div>
);

export interface ActionsSectionProps {
  actions: AutomationAction[];
  onAdd: () => void;
  onRemove: (idx: number) => void;
  onUpdate: (idx: number, patch: Partial<AutomationAction>) => void;
  allColumns: { id: string; name: string }[];
}

export const ActionsSection: React.FC<ActionsSectionProps> = ({
  actions,
  onAdd,
  onRemove,
  onUpdate,
  allColumns,
}) => (
  <div className="p-3 rounded-xl border border-subtle bg-surface/60 space-y-2.5">
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <span className="w-5 h-5 rounded-full bg-accent-solid text-white text-[10px] font-bold flex items-center justify-center shrink-0">
          3
        </span>
        <span className="text-xs font-semibold uppercase tracking-wider text-primary">
          ТО (Последовательность действий)
        </span>
        <Badge variant="subtle" size="sm" className="font-mono text-[10px]">{actions.length}/5</Badge>
      </div>
      {actions.length < 5 && (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onAdd}
          leftIcon={<Plus className="w-3.5 h-3.5" />}
          className="text-[11px] h-7 px-2"
        >
          Добавить действие
        </Button>
      )}
    </div>

    <div className="space-y-2">
      {actions.map((act, index) => (
        <div key={index} className="flex items-center gap-2 p-2 rounded-lg bg-surface border border-subtle text-xs">
          <span className="text-[10px] font-mono text-muted shrink-0 w-6">
            {index + 1}.
          </span>
          <select
            value={act.type}
            onChange={(e) => onUpdate(index, { type: e.target.value as AutomationActionType, params: {} })}
            className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary"
          >
            {ACTION_TYPES.map((a) => (
              <option key={a.type} value={a.type}>{a.label}</option>
            ))}
          </select>

          {act.type === 'move_to_column' && (
            <select
              value={String(act.params?.target_column_id || '')}
              onChange={(e) => onUpdate(index, { params: { ...act.params, target_column_id: e.target.value } })}
              className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary flex-1"
            >
              <option value="">Выберите колонку...</option>
              {allColumns.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          )}

          {act.type === 'set_priority' && (
            <select
              value={String(act.params?.priority || 'high')}
              onChange={(e) => onUpdate(index, { params: { ...act.params, priority: e.target.value } })}
              className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary flex-1"
            >
              {PRIORITIES.map((p) => (
                <option key={p.value} value={p.value}>{p.label}</option>
              ))}
            </select>
          )}

          {(act.type === 'add_tag' || act.type === 'remove_tag') && (
            <input
              type="text"
              value={String(act.params?.name || '')}
              onChange={(e) => onUpdate(index, { params: { ...act.params, name: e.target.value } })}
              placeholder="Имя тега (напр. @бэкенд)"
              className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary flex-1 focus:outline-none"
            />
          )}

          {act.type === 'set_due_date' && (
            <select
              value={String(act.params?.preset || 'today')}
              onChange={(e) => onUpdate(index, { params: { ...act.params, preset: e.target.value } })}
              className="bg-canvas/50 border border-subtle rounded px-2 py-1 text-xs text-primary flex-1"
            >
              <option value="today">Сегодня</option>
              <option value="tomorrow">Завтра</option>
              <option value="clear">Очистить срок</option>
            </select>
          )}

          {actions.length > 1 && (
            <button
              type="button"
              onClick={() => onRemove(index)}
              className="text-muted hover:text-primary p-1 rounded transition-colors shrink-0"
              title="Удалить действие"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      ))}
    </div>
  </div>
);
