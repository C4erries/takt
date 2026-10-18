import React, { useState, useEffect } from 'react';
import {
  Zap,
  Sparkles,
  ArrowRight,
  ShieldCheck,
  Plus,
  Trash2,
  Paperclip,
  Link2Off,
  ChevronDown,
  Pencil,
  Copy,
} from 'lucide-react';
import { Modal, Button, Badge } from '@/shared/components/ui';
import type { Column } from '@/entities/project';
import type { Priority } from '@/shared/api';
import {
  useAutomationsQuery,
  useToggleAutomationMutation,
  useDeleteAutomationMutation,
  useUpdateAutomationMutation,
  useCreateAutomationMutation,
  type Automation,
} from '@/entities/automation';
import {
  RuleBuilderModal,
  RuleDetailModal,
  formatTriggerLabel,
} from '@/features/automations';
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

export function getStoredColumnAutomationConfig(colId: string, defId: string | null = null): ColumnAutomationConfig {
  try {
    const raw = localStorage.getItem(`${STORAGE_PREFIX}${colId}`);
    if (raw) return JSON.parse(raw);
  } catch {}
  return { enterPriorityEnabled: false, enterPriority: 'high', completeMoveEnabled: false, completeTargetColumnId: defId };
}

export function saveStoredColumnAutomationConfig(colId: string, cfg: ColumnAutomationConfig): void {
  try { localStorage.setItem(`${STORAGE_PREFIX}${colId}`, JSON.stringify(cfg)); } catch {}
}

const PRIORITIES = [
  { value: 'none' as const, label: 'Обычный', badgeVariant: 'subtle' as const },
  { value: 'low' as const, label: 'Низкий', badgeVariant: 'subtle' as const },
  { value: 'medium' as const, label: 'Средний', badgeVariant: 'straw' as const },
  { value: 'high' as const, label: 'Высокий', badgeVariant: 'denim' as const },
  { value: 'urgent' as const, label: 'Срочный', badgeVariant: 'persimmon' as const },
];

export const ColumnAutomationsModal: React.FC<ColumnAutomationsModalProps> = ({
  isOpen,
  onClose,
  column,
  allColumns,
  workspaceId,
}) => {
  const otherColumns = allColumns.filter((c) => c.id !== column.id);
  const defaultTargetId =
    otherColumns.find((c) => c.statusCategory === 'done')?.id ?? otherColumns[0]?.id ?? null;

  const [config, setConfig] = useState<ColumnAutomationConfig>(() =>
    getStoredColumnAutomationConfig(column.id, defaultTargetId)
  );
  const [isBuilderOpen, setIsBuilderOpen] = useState(false);
  const [isAttachOpen, setIsAttachOpen] = useState(false);
  const [detailRule, setDetailRule] = useState<Automation | null>(null);
  const [editingRule, setEditingRule] = useState<Automation | null>(null);

  const { data: automations = [] } = useAutomationsQuery(workspaceId);
  const toggleMutation = useToggleAutomationMutation(workspaceId);
  const deleteMutation = useDeleteAutomationMutation(workspaceId);
  const updateMutation = useUpdateAutomationMutation(workspaceId);
  const createMutation = useCreateAutomationMutation(workspaceId);

  const columnAutomations = automations.filter((a) => a.columnId === column.id);
  const availableAutomations = automations.filter((a) => a.columnId !== column.id);

  useEffect(() => {
    if (isOpen) {
      setConfig(getStoredColumnAutomationConfig(column.id, defaultTargetId));
      setIsAttachOpen(false);
    }
  }, [isOpen, column.id, defaultTargetId]);

  const handleUpdate = (updated: Partial<ColumnAutomationConfig>) => {
    setConfig((prev) => {
      const next = { ...prev, ...updated };
      saveStoredColumnAutomationConfig(column.id, next);
      return next;
    });
  };

  const handleDuplicate = (rule: Automation) => {
    createMutation.mutate({
      name: `${rule.name} (копия)`,
      triggerType: rule.triggerType,
      boardId: rule.boardId,
      columnId: rule.columnId ?? column.id,
      conditionsJson: rule.conditionsJson,
      actionsJson: rule.actionsJson,
      isEnabled: rule.isEnabled,
      ruleDefinition: rule.ruleDefinition,
    });
  };

  const handleEdit = (rule: Automation) => {
    setEditingRule(rule);
    setIsBuilderOpen(true);
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
          {/* Banner */}
          <div className="flex items-center justify-between p-3 rounded-xl bg-accent-surface/40 border border-subtle text-xs text-primary gap-2">
            <div className="flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-accent-solid shrink-0" />
              <span>Правила срабатывают мгновенно и защищены от зацикливания.</span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => { setEditingRule(null); setIsBuilderOpen(true); }}
              leftIcon={<Plus className="w-3.5 h-3.5" />}
              className="whitespace-nowrap min-w-fit px-3.5 py-1.5 h-8 text-[11px] bg-surface/80 shadow-washi-sm"
            >
              Собрать правило
            </Button>
          </div>

          {/* Preset 1: Upon Entering Column */}
          <div className="p-3 rounded-xl border border-subtle bg-surface/50 space-y-2.5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Zap className="w-4 h-4 text-accent-solid" />
                <div>
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-primary">При попадании в колонку</h4>
                  <p className="text-[11px] text-muted">Менять приоритет задачи при создании или переносе сюда</p>
                </div>
              </div>
              <button
                type="button"
                role="switch"
                aria-checked={config.enterPriorityEnabled}
                onClick={() => handleUpdate({ enterPriorityEnabled: !config.enterPriorityEnabled })}
                className={cn('w-10 h-5 flex items-center rounded-full p-0.5 transition-colors shrink-0', config.enterPriorityEnabled ? 'bg-accent-solid justify-end' : 'bg-surface border border-subtle justify-start')}
              >
                <span className="w-4 h-4 rounded-full bg-white shadow-sm" />
              </button>
            </div>
            {config.enterPriorityEnabled && (
              <div className="pt-2 border-t border-subtle/50 flex flex-wrap gap-1.5 items-center">
                <span className="text-[11px] font-mono uppercase text-muted mr-1">Приоритет:</span>
                {PRIORITIES.map((p) => (
                  <button
                    key={p.value}
                    type="button"
                    onClick={() => handleUpdate({ enterPriority: p.value })}
                    className={cn('rounded-full transition-all', config.enterPriority === p.value ? 'ring-2 ring-accent-solid scale-105' : 'opacity-60 hover:opacity-100')}
                  >
                    <Badge variant={p.badgeVariant} size="sm">{p.label}</Badge>
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Preset 2: Upon Task Completion */}
          <div className="p-3 rounded-xl border border-subtle bg-surface/50 space-y-2.5">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Sparkles className="w-4 h-4 text-accent-solid" />
                <div>
                  <h4 className="text-xs font-semibold uppercase tracking-wider text-primary">При завершении задачи</h4>
                  <p className="text-[11px] text-muted">Переносить задачу в целевую колонку по клику на чекбокс</p>
                </div>
              </div>
              <button
                type="button"
                role="switch"
                aria-checked={config.completeMoveEnabled}
                onClick={() => handleUpdate({ completeMoveEnabled: !config.completeMoveEnabled })}
                className={cn('w-10 h-5 flex items-center rounded-full p-0.5 transition-colors shrink-0', config.completeMoveEnabled ? 'bg-accent-solid justify-end' : 'bg-surface border border-subtle justify-start')}
              >
                <span className="w-4 h-4 rounded-full bg-white shadow-sm" />
              </button>
            </div>
            {config.completeMoveEnabled && (
              <div className="pt-2 border-t border-subtle/50 space-y-1">
                <span className="text-[11px] font-mono uppercase text-muted">Куда переместить:</span>
                <div className="grid grid-cols-2 gap-1.5">
                  {otherColumns.map((c) => (
                    <button
                      key={c.id}
                      type="button"
                      onClick={() => handleUpdate({ completeTargetColumnId: c.id })}
                      className={cn(
                        'flex items-center justify-between px-2 py-1.5 rounded-lg text-xs border text-left transition-all truncate',
                        config.completeTargetColumnId === c.id ? 'border-accent-solid bg-accent-surface/30 font-medium text-primary shadow-washi-sm' : 'border-subtle bg-canvas/40 text-muted hover:text-primary hover:bg-surface/80'
                      )}
                    >
                      <span className="truncate">{c.name}</span>
                      {config.completeTargetColumnId === c.id && <ArrowRight className="w-3.5 h-3.5 text-accent-solid shrink-0 ml-1" />}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Custom Rules Section & Catalog Attachment */}
          <div className="p-3 rounded-xl border border-subtle bg-surface/50 space-y-2.5">
            <div className="flex items-center justify-between gap-2 flex-wrap">
              <h4 className="text-xs font-semibold uppercase tracking-wider text-primary">
                Кастомные правила для колонки ({columnAutomations.length})
              </h4>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsAttachOpen((prev) => !prev)}
                leftIcon={<Paperclip className="w-3.5 h-3.5" />}
                rightIcon={<ChevronDown className={cn('w-3 h-3 transition-transform', isAttachOpen && 'rotate-180')} />}
                className="text-[11px] h-7 px-2"
              >
                Прикрепить из каталога
              </Button>
            </div>

            {/* Catalog Picker dropdown */}
            {isAttachOpen && (
              <div className="p-2.5 rounded-lg border border-subtle bg-surface space-y-2 animate-in fade-in duration-150">
                <span className="text-[10px] font-mono uppercase text-muted">Выберите правило для привязки:</span>
                {availableAutomations.length === 0 ? (
                  <p className="text-xs text-muted italic py-1">Нет других правил для прикрепления.</p>
                ) : (
                  <div className="space-y-1.5 max-h-36 overflow-y-auto pr-0.5">
                    {availableAutomations.map((a) => (
                      <div key={a.id} className="flex items-center justify-between gap-2 p-2 rounded-lg border border-subtle bg-canvas/40 text-xs hover:bg-canvas/70 transition-colors">
                        <div className="min-w-0 cursor-pointer flex-1" onClick={() => setDetailRule(a)}>
                          <div className="font-medium text-primary truncate hover:underline">{a.name}</div>
                          <div className="text-[10px] text-muted truncate">
                            {formatTriggerLabel(a.triggerType, a.columnId, allColumns, a.ruleDefinition?.trigger?.params)}
                          </div>
                        </div>
                        <div className="flex items-center gap-1.5 shrink-0">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setDetailRule(a)}
                            className="text-[10px] h-6 px-1.5 text-muted hover:text-primary"
                          >
                            Подробнее
                          </Button>
                          <Button
                            variant="solid"
                            size="sm"
                            onClick={() => { updateMutation.mutate({ automationId: a.id, data: { columnId: column.id } }); setIsAttachOpen(false); }}
                            className="text-[10px] h-6 px-2 shrink-0"
                          >
                            Привязать
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* List of Attached Rules */}
            {columnAutomations.length === 0 ? (
              <p className="text-xs text-muted italic px-2 py-1 bg-canvas/30 rounded border border-subtle/40">
                К этой колонке пока не привязано кастомных правил.
              </p>
            ) : (
              <div className="space-y-1.5">
                {columnAutomations.map((rule) => (
                  <div key={rule.id} className="flex items-center justify-between gap-2 p-2 rounded-lg border border-subtle bg-canvas/50 text-xs">
                    <div className="min-w-0 flex-1 cursor-pointer hover:opacity-80 transition-opacity" onClick={() => setDetailRule(rule)}>
                      <div className="font-medium text-primary truncate hover:underline">{rule.name}</div>
                      <div className="text-[10px] text-muted truncate">
                        {formatTriggerLabel(rule.triggerType, rule.columnId, allColumns, rule.ruleDefinition?.trigger?.params)}
                      </div>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setDetailRule(rule)}
                        className="text-[10px] h-6 px-1.5 text-muted hover:text-primary"
                      >
                        Подробнее
                      </Button>
                      <button type="button" onClick={() => handleDuplicate(rule)} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Создать копию"><Copy className="w-3.5 h-3.5" /></button>
                      <button type="button" onClick={() => handleEdit(rule)} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Редактировать правило"><Pencil className="w-3.5 h-3.5" /></button>
                      <button type="button" onClick={() => updateMutation.mutate({ automationId: rule.id, data: { columnId: null } })} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Отвязать"><Link2Off className="w-3.5 h-3.5" /></button>
                      <button type="button" onClick={() => deleteMutation.mutate(rule.id)} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Удалить"><Trash2 className="w-3.5 h-3.5" /></button>
                      <button
                        type="button"
                        role="switch"
                        aria-checked={rule.isEnabled}
                        onClick={() => toggleMutation.mutate({ automationId: rule.id, isEnabled: !rule.isEnabled })}
                        className={cn('w-8 h-4 flex items-center rounded-full p-0.5 ml-0.5 transition-colors cursor-pointer', rule.isEnabled ? 'bg-accent-solid justify-end' : 'bg-surface border border-subtle justify-start')}
                        title={rule.isEnabled ? 'Выключить' : 'Включить'}
                      >
                        <span className="w-3 h-3 rounded-full bg-white shadow-sm" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

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
        onClose={() => { setIsBuilderOpen(false); setEditingRule(null); }}
        workspaceId={workspaceId}
        prefilledColumnId={column.id}
        prefilledColumnName={column.name}
        allColumns={allColumns}
        mode={editingRule ? 'edit' : 'create'}
        initialRule={editingRule}
      />

      {/* RuleDetailModal */}
      <RuleDetailModal
        isOpen={Boolean(detailRule)}
        onClose={() => setDetailRule(null)}
        rule={detailRule}
        allColumns={allColumns}
        onEdit={(r) => { setDetailRule(null); handleEdit(r); }}
        onDuplicate={handleDuplicate}
        onBind={
          detailRule && detailRule.columnId !== column.id
            ? (r) => {
                updateMutation.mutate({ automationId: r.id, data: { columnId: column.id } });
                setDetailRule(null);
              }
            : undefined
        }
      />
    </>
  );
};
