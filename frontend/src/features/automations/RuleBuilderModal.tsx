import React from 'react';
import { Sparkles, Clock, Zap, Copy, Pencil, ArrowRight, Paperclip } from 'lucide-react';
import { Modal, Button, Input, Badge } from '@/shared/components/ui';
import type { Automation, AutomationCondition, AutomationAction } from '@/entities/automation';
import {
  useRuleBuilder,
  formatTriggerLabel,
  formatConditionLabel,
  formatActionLabel,
} from './useRuleBuilder';
import {
  TriggerSection,
  ConditionsSection,
  ActionsSection,
} from './RuleBuilderSections';

export interface RuleBuilderModalProps {
  isOpen: boolean;
  onClose: () => void;
  workspaceId?: string;
  prefilledColumnId?: string | null;
  prefilledColumnName?: string | null;
  allColumns?: { id: string; name: string }[];
  mode?: 'create' | 'edit';
  initialRule?: Automation | null;
  onSuccess?: (saved: Automation) => void;
}

export const RuleBuilderModal: React.FC<RuleBuilderModalProps> = ({
  isOpen,
  onClose,
  workspaceId,
  prefilledColumnId,
  prefilledColumnName,
  allColumns = [],
  mode = 'create',
  initialRule,
  onSuccess,
}) => {
  const isEdit = mode === 'edit';
  const {
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
    isSubmitting,
  } = useRuleBuilder({
    workspaceId,
    prefilledColumnId,
    prefilledColumnName,
    allColumns,
    mode,
    initialRule,
    onSuccess,
    onClose,
  });

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isEdit ? 'Редактирование правила' : 'Конструктор правил автоматизации'}
      description={
        isEdit
          ? 'Измените параметры триггера, условий или действий'
          : 'Настройте условия и действия для автоматического управления задачами'
      }
      className="max-w-2xl"
    >
      <div className="space-y-4 max-h-[75vh] overflow-y-auto pr-1">
        {/* Name Input */}
        <div className="space-y-1">
          <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
            Название правила
          </label>
          <Input
            value={ruleName}
            onChange={(e) => {
              setIsCustomName(true);
              setRuleName(e.target.value);
            }}
            placeholder="Например: Срочные задачи → в колонку В работе"
            className="text-xs bg-surface border-subtle"
          />
        </div>

        {/* STEP 1: WHEN */}
        <TriggerSection
          triggerType={triggerType}
          setTriggerType={setTriggerType}
          triggerColumnId={triggerColumnId}
          setTriggerColumnId={setTriggerColumnId}
          triggerSlaHours={triggerSlaHours}
          setTriggerSlaHours={setTriggerSlaHours}
          triggerField={triggerField}
          setTriggerField={setTriggerField}
          allColumns={allColumns}
        />

        {/* STEP 2: IF */}
        <ConditionsSection
          conditions={conditions}
          logicalOperator={logicalOperator}
          onLogicalOperatorChange={setLogicalOperator}
          onAdd={handleAddCondition}
          onRemove={handleRemoveCondition}
          onUpdate={handleUpdateCondition}
        />

        {/* STEP 3: THEN */}
        <ActionsSection
          actions={actions}
          onAdd={handleAddAction}
          onRemove={handleRemoveAction}
          onUpdate={handleUpdateAction}
          allColumns={allColumns}
        />

        {/* Action Buttons */}
        <div className="flex justify-end gap-2 pt-2 border-t border-subtle/60">
          <Button variant="ghost" size="sm" onClick={onClose}>
            Отмена
          </Button>
          <Button
            variant="solid"
            size="sm"
            onClick={handleSave}
            isLoading={isSubmitting}
            disabled={!ruleName.trim() || actions.length === 0}
            leftIcon={isEdit ? <Pencil className="w-3.5 h-3.5" /> : <Sparkles className="w-3.5 h-3.5" />}
          >
            {isEdit ? 'Сохранить изменения' : 'Создать правило'}
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export interface RuleDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  rule: Automation | null;
  allColumns?: { id: string; name: string }[];
  onEdit?: (rule: Automation) => void;
  onDuplicate?: (rule: Automation) => void;
  onBind?: (rule: Automation) => void;
}

export const RuleDetailModal: React.FC<RuleDetailModalProps> = ({
  isOpen,
  onClose,
  rule,
  allColumns = [],
  onEdit,
  onDuplicate,
  onBind,
}) => {
  if (!rule) return null;

  const isTimeTrigger =
    rule.triggerType === 'column_sla_timer' || rule.triggerType === 'scheduled_sweep';
  const triggerLabel = formatTriggerLabel(
    rule.triggerType,
    rule.columnId,
    allColumns,
    rule.ruleDefinition?.trigger?.params
  );

  const conditions: AutomationCondition[] =
    rule.ruleDefinition?.conditions ||
    ((rule.conditionsJson as { conditions?: AutomationCondition[] })?.conditions) ||
    [];
  const logicalOp =
    rule.ruleDefinition?.logicalOperator ||
    ((rule.conditionsJson as { logical_operator?: 'and' | 'or' })?.logical_operator) ||
    'and';

  const actions: AutomationAction[] =
    rule.ruleDefinition?.actions ||
    ((rule.actionsJson as { actions?: AutomationAction[] })?.actions) ||
    [];

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Детали правила"
      description="Просмотр настроенного сценария автоматизации"
      className="max-w-xl"
    >
      <div className="space-y-4 max-h-[75vh] overflow-y-auto pr-1">
        {/* Name and State Banner */}
        <div className="flex items-start justify-between gap-3 p-3.5 rounded-xl bg-surface border border-subtle shadow-washi-sm">
          <div className="space-y-1 min-w-0">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted">
              Название сценария
            </span>
            <h3 className="font-serif text-base font-normal text-primary break-words">
              {rule.name}
            </h3>
          </div>
          <Badge variant={rule.isEnabled ? 'matcha' : 'subtle'} size="sm" dot className="shrink-0 mt-0.5">
            {rule.isEnabled ? 'Активно' : 'Отключено'}
          </Badge>
        </div>

        {/* Step 1: WHEN */}
        <div className="p-3.5 rounded-xl border border-subtle bg-surface/50 space-y-2">
          <div className="flex items-center gap-1.5 text-[11px] font-mono uppercase tracking-wider text-muted">
            {isTimeTrigger ? (
              <Clock className="w-3.5 h-3.5 text-accent-solid" />
            ) : rule.triggerType.includes('completed') ? (
              <Sparkles className="w-3.5 h-3.5 text-accent-solid" />
            ) : (
              <Zap className="w-3.5 h-3.5 text-accent-solid" />
            )}
            <span>КОГДА (Триггер)</span>
          </div>
          <div className="p-2.5 rounded-lg border border-subtle bg-canvas text-xs text-primary font-medium">
            {triggerLabel}
          </div>
        </div>

        {/* Step 2: IF */}
        <div className="p-3.5 rounded-xl border border-subtle bg-surface/50 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono uppercase tracking-wider text-muted">
              ЕСЛИ (Условия)
            </span>
            {conditions.length > 1 && (
              <span className="text-[10px] font-mono text-muted uppercase">
                {logicalOp === 'or' ? 'Любое из условий (ИЛИ)' : 'Все условия (И)'}
              </span>
            )}
          </div>
          {conditions.length === 0 ? (
            <p className="text-xs text-muted italic px-2 py-1 bg-canvas/30 rounded border border-subtle/30">
              Без дополнительных условий (срабатывает всегда)
            </p>
          ) : (
            <div className="space-y-1.5">
              {conditions.map((c, i) => (
                <div
                  key={i}
                  className="flex items-center gap-2 p-2 rounded-lg border border-subtle bg-canvas text-xs text-primary"
                >
                  <span className="text-[10px] font-mono text-muted w-4">{i + 1}.</span>
                  <span>{formatConditionLabel(c)}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Step 3: THEN */}
        <div className="p-3.5 rounded-xl border border-subtle bg-surface/50 space-y-2">
          <div className="flex items-center gap-1.5 text-[11px] font-mono uppercase tracking-wider text-muted">
            <ArrowRight className="w-3.5 h-3.5 text-accent-solid" />
            <span>ТО (Действия)</span>
          </div>
          {actions.length === 0 ? (
            <p className="text-xs text-muted italic px-2 py-1 bg-canvas/30 rounded border border-subtle/30">
              Действия не настроены
            </p>
          ) : (
            <div className="space-y-1.5">
              {actions.map((a, i) => (
                <div
                  key={i}
                  className="flex items-center gap-2 p-2 rounded-lg border border-subtle bg-canvas text-xs text-primary font-medium"
                >
                  <span className="text-[10px] font-mono text-muted w-4">{i + 1}.</span>
                  <span>{formatActionLabel(a, allColumns)}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="flex items-center justify-between gap-2 pt-3 border-t border-subtle">
          <div className="flex items-center gap-2">
            {onDuplicate && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  onDuplicate(rule);
                  onClose();
                }}
                leftIcon={<Copy className="w-3.5 h-3.5" />}
                className="text-xs h-8 shadow-washi-sm"
              >
                Создать копию
              </Button>
            )}
            {onBind && (
              <Button
                variant="solid"
                size="sm"
                onClick={() => {
                  onBind(rule);
                  onClose();
                }}
                leftIcon={<Paperclip className="w-3.5 h-3.5" />}
                className="text-xs h-8"
              >
                Привязать к колонке
              </Button>
            )}
          </div>
          <div className="flex items-center gap-2">
            {onEdit && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  onEdit(rule);
                  onClose();
                }}
                leftIcon={<Pencil className="w-3.5 h-3.5" />}
                className="text-xs h-8"
              >
                Редактировать
              </Button>
            )}
            <Button variant="ghost" size="sm" onClick={onClose} className="text-xs h-8">
              Закрыть
            </Button>
          </div>
        </div>
      </div>
    </Modal>
  );
};
