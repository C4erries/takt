import React, { useState, useEffect } from 'react';
import { Zap, Sparkles, ArrowRight, ShieldCheck, Plus, Trash2 } from 'lucide-react';
import { Modal, Button, Badge } from '@/shared/components/ui';
import type { Column } from '@/entities/project';
import type { Priority } from '@/shared/api';
import {
  useAutomationsQuery,
  useToggleAutomationMutation,
  useDeleteAutomationMutation,
  type Automation,
} from '@/entities/automation';
import { RuleBuilderModal } from '@/features/automations';
import { cn } from '@/shared/lib/cn';

export interface ColumnAutomationsModalProps {
  isOpen: boolean;
  onClose: () => void;
  column: Column;
  allColumns: Column[];
  workspaceId?: string;
}

export interface ColumnAutomationConfig {
  enterPriorityEnabled: boolean;
  enterPriority: Priority;
  completeMoveEnabled: boolean;
  completeTargetColumnId: string | null;
}

export const STORAGE_PREFIX = 'takt_col_auto_';

export function getStoredColumnAutomationConfig(
  columnId: string,
  defaultTargetId: string | null = null
): ColumnAutomationConfig {
  try {
    const raw = localStorage.getItem(`${STORAGE_PREFIX}${columnId}`);
    if (raw) return JSON.parse(raw);
  } catch {
    // fallback
  }
  return {
    enterPriorityEnabled: false,
    enterPriority: 'high',
    completeMoveEnabled: false,
    completeTargetColumnId: defaultTargetId,
  };
}

export function saveStoredColumnAutomationConfig(
  columnId: string,
  config: ColumnAutomationConfig
): void {
  try {
    localStorage.setItem(`${STORAGE_PREFIX}${columnId}`, JSON.stringify(config));
  } catch {
    // ignore
  }
}

const PRIORITIES: { value: Priority; label: string; badgeVariant: 'subtle' | 'straw' | 'denim' | 'persimmon' }[] = [
  { value: 'none', label: 'Обычный', badgeVariant: 'subtle' },
  { value: 'low', label: 'Низкий', badgeVariant: 'subtle' },
  { value: 'medium', label: 'Средний', badgeVariant: 'straw' },
  { value: 'high', label: 'Высокий', badgeVariant: 'denim' },
  { value: 'urgent', label: 'Срочный', badgeVariant: 'persimmon' },
];

export const ColumnAutomationsModal: React.FC<ColumnAutomationsModalProps> = ({
  isOpen,
  onClose,
  column,
  allColumns,
  workspaceId,
}) => {
  const otherColumns = allColumns.filter((c) => c.id !== column.id);
  const defaultTargetId = otherColumns.find((c) => c.statusCategory === 'done')?.id ?? otherColumns[0]?.id ?? null;

  const [config, setConfig] = useState<ColumnAutomationConfig>(() =>
    getStoredColumnAutomationConfig(column.id, defaultTargetId)
  );
  const [isBuilderOpen, setIsBuilderOpen] = useState(false);

  const { data: automations = [] } = useAutomationsQuery(workspaceId);
  const toggleMutation = useToggleAutomationMutation(workspaceId);
  const deleteMutation = useDeleteAutomationMutation(workspaceId);

  const columnAutomations = automations.filter((a) => a.columnId === column.id);

  useEffect(() => {
    if (isOpen) {
      setConfig(getStoredColumnAutomationConfig(column.id, defaultTargetId));
    }
  }, [isOpen, column.id, defaultTargetId]);

  const handleUpdate = (updated: Partial<ColumnAutomationConfig>) => {
    setConfig((prev) => {
      const next = { ...prev, ...updated };
      saveStoredColumnAutomationConfig(column.id, next);
      return next;
    });
  };

  const handleToggleCustom = (auto: Automation) => {
    toggleMutation.mutate({
      automationId: auto.id,
      isEnabled: !auto.isEnabled,
    });
  };

  return (
    <>
      <Modal
        isOpen={isOpen}
        onClose={onClose}
        title={`Автоматизации колонки: ${column.name}`}
        description="Правила и пресеты, автоматически срабатывающие для задач этой колонки"
        className="max-w-xl"
      >
        <div className="space-y-4 max-h-[75vh] overflow-y-auto pr-1">
          {/* Banner about Cycle Guard */}
          <div className="flex items-center justify-between p-3 rounded-xl bg-accent-surface/40 border border-subtle text-xs text-primary">
            <div className="flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-accent-solid shrink-0" />
              <span>EventBus & Cycle Guard (глубина ≤ 3 шагов).</span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setIsBuilderOpen(true)}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
              className="text-[11px] h-7 bg-surface/80 shadow-washi-sm"
            >
              Собрать своё правило
            </Button>
          </div>

          {/* Preset 1: Upon Entering Column */}
          <div className="p-3.5 rounded-xl border border-subtle bg-surface/50 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Zap className="w-4 h-4 text-accent-solid" />
                <div>
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-primary">
                    При попадании в колонку
                  </h4>
                  <p className="text-[11px] text-muted">
                    Автоматически менять приоритет задачи при создании или переносе в эту колонку
                  </p>
                </div>
              </div>
              <button
                type="button"
                role="switch"
                aria-checked={config.enterPriorityEnabled}
                onClick={() => handleUpdate({ enterPriorityEnabled: !config.enterPriorityEnabled })}
                className={cn(
                  'w-10 h-5 flex items-center rounded-full p-0.5 transition-colors duration-200 shrink-0',
                  config.enterPriorityEnabled ? 'bg-accent-solid justify-end' : 'bg-surface border border-subtle justify-start'
                )}
              >
                <span className="w-4 h-4 rounded-full bg-white shadow-sm transition-transform duration-200" />
              </button>
            </div>

            {config.enterPriorityEnabled && (
              <div className="pt-2 border-t border-subtle/50 space-y-1.5 animate-in fade-in duration-150">
                <span className="text-[11px] font-mono uppercase text-muted">Установить приоритет:</span>
                <div className="flex flex-wrap gap-1.5">
                  {PRIORITIES.map((p) => {
                    const isSelected = config.enterPriority === p.value;
                    return (
                      <button
                        key={p.value}
                        type="button"
                        onClick={() => handleUpdate({ enterPriority: p.value })}
                        className={cn(
                          'rounded-full transition-all',
                          isSelected ? 'ring-2 ring-accent-solid scale-105' : 'opacity-60 hover:opacity-100'
                        )}
                      >
                        <Badge variant={p.badgeVariant} size="sm">
                          {p.label}
                        </Badge>
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Preset 2: Upon Task Completion */}
          <div className="p-3.5 rounded-xl border border-subtle bg-surface/50 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Sparkles className="w-4 h-4 text-accent-solid" />
                <div>
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-primary">
                    При завершении задачи
                  </h4>
                  <p className="text-[11px] text-muted">
                    Автоматически переносить задачу в целевую колонку при отметке чекбокса выполнения
                  </p>
                </div>
              </div>
              <button
                type="button"
                role="switch"
                aria-checked={config.completeMoveEnabled}
                onClick={() => handleUpdate({ completeMoveEnabled: !config.completeMoveEnabled })}
                className={cn(
                  'w-10 h-5 flex items-center rounded-full p-0.5 transition-colors duration-200 shrink-0',
                  config.completeMoveEnabled ? 'bg-accent-solid justify-end' : 'bg-surface border border-subtle justify-start'
                )}
              >
                <span className="w-4 h-4 rounded-full bg-white shadow-sm transition-transform duration-200" />
              </button>
            </div>

            {config.completeMoveEnabled && (
              <div className="pt-2 border-t border-subtle/50 space-y-1.5 animate-in fade-in duration-150">
                <span className="text-[11px] font-mono uppercase text-muted">Целевая колонка для переноса:</span>
                {otherColumns.length === 0 ? (
                  <p className="text-xs text-muted italic">На доске нет других колонок для переноса.</p>
                ) : (
                  <div className="grid grid-cols-2 gap-2">
                    {otherColumns.map((c) => {
                      const isSelected = config.completeTargetColumnId === c.id;
                      return (
                        <button
                          key={c.id}
                          type="button"
                          onClick={() => handleUpdate({ completeTargetColumnId: c.id })}
                          className={cn(
                            'flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs border text-left transition-all',
                            isSelected
                              ? 'border-accent-solid bg-accent-surface/30 font-medium text-primary shadow-washi-sm'
                              : 'border-subtle bg-canvas/40 text-muted hover:text-primary hover:bg-surface/80'
                          )}
                        >
                          <span className="truncate">{c.name}</span>
                          {isSelected && <ArrowRight className="w-3.5 h-3.5 text-accent-solid shrink-0" />}
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Custom Rules for this column */}
          {columnAutomations.length > 0 && (
            <div className="p-3.5 rounded-xl border border-subtle bg-surface/50 space-y-2.5">
              <h4 className="text-xs font-semibold uppercase tracking-wider text-primary">
                Кастомные правила для колонки ({columnAutomations.length})
              </h4>
              <div className="space-y-2">
                {columnAutomations.map((rule) => (
                  <div
                    key={rule.id}
                    className="flex items-center justify-between gap-3 p-2.5 rounded-lg border border-subtle bg-canvas/50 text-xs"
                  >
                    <div className="min-w-0">
                      <div className="font-medium text-primary truncate">{rule.name}</div>
                      <div className="text-[10px] font-mono text-muted truncate">
                        {rule.triggerType}
                      </div>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <button
                        type="button"
                        onClick={() => deleteMutation.mutate(rule.id)}
                        className="p-1 text-muted hover:text-primary transition-colors"
                        title="Удалить"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                      <button
                        type="button"
                        role="switch"
                        aria-checked={rule.isEnabled}
                        onClick={() => handleToggleCustom(rule)}
                        className={cn(
                          'w-8 h-4 flex items-center rounded-full p-0.5 transition-colors duration-200',
                          rule.isEnabled ? 'bg-accent-solid justify-end' : 'bg-surface border border-subtle justify-start'
                        )}
                      >
                        <span className="w-3 h-3 rounded-full bg-white shadow-sm" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Modal Actions */}
          <div className="flex justify-end pt-2 border-t border-subtle">
            <Button variant="solid" size="sm" onClick={onClose} className="px-4">
              Готово
            </Button>
          </div>
        </div>
      </Modal>

      {/* RuleBuilderModal */}
      <RuleBuilderModal
        isOpen={isBuilderOpen}
        onClose={() => setIsBuilderOpen(false)}
        workspaceId={workspaceId}
        prefilledColumnId={column.id}
        prefilledColumnName={column.name}
        allColumns={allColumns}
      />
    </>
  );
};
