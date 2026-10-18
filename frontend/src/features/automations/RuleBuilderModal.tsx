import React from 'react';
import { Sparkles } from 'lucide-react';
import { Modal, Button, Input } from '@/shared/components/ui';
import type { Automation } from '@/entities/automation';
import { useRuleBuilder } from './useRuleBuilder';
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
  onSuccess?: (created: Automation) => void;
}

export const RuleBuilderModal: React.FC<RuleBuilderModalProps> = ({
  isOpen,
  onClose,
  workspaceId,
  prefilledColumnId,
  prefilledColumnName,
  allColumns = [],
  onSuccess,
}) => {
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
    onSuccess,
    onClose,
  });

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Конструктор правил автоматизации"
      description="Настройте условия и действия для автоматического управления задачами"
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
            leftIcon={<Sparkles className="w-3.5 h-3.5" />}
          >
            Создать правило
          </Button>
        </div>
      </div>
    </Modal>
  );
};
