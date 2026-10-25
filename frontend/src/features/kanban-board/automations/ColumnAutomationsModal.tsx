import React, { useState, useEffect, useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
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
import { Modal, Button } from '@/shared/components/ui';
import {
  type Column,
  type Board,
  useProjectsQuery,
  useProjectAllColumnsQuery,
  projectApi,
  DEMO_BOARDS,
} from '@/entities/project';
import { getAuthToken } from '@/shared/api';
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

export * from './columnAutomationConfig';
import {
  type ColumnAutomationConfig,
  getStoredColumnAutomationConfig,
  saveStoredColumnAutomationConfig,
  PRIORITIES,
  PRIORITY_ACTIVE_STYLES,
} from './columnAutomationConfig';

export interface ColumnAutomationsModalProps {
  isOpen: boolean;
  onClose: () => void;
  column: Column;
  allColumns: Column[];
  workspaceId?: string;
  boardColumnGroups?: { boardName: string; projectName?: string; columns: { id: string; name: string }[] }[];
  allBoards?: { id: string; name: string; projectId?: string; projectName?: string }[];
}

export const ColumnAutomationsModal: React.FC<ColumnAutomationsModalProps> = ({
  isOpen,
  onClose,
  column,
  allColumns,
  workspaceId,
  boardColumnGroups: propGroups,
  allBoards: propBoards,
}) => {
  const otherColumns = allColumns.filter((c) => c.id !== column.id);
  const defaultTargetId =
    otherColumns.find((c) => c.statusCategory === 'done')?.id ?? otherColumns[0]?.id ?? null;

  const [config, setConfig] = useState<ColumnAutomationConfig>(() => getStoredColumnAutomationConfig(column.id, defaultTargetId));
  const [isBuilderOpen, setIsBuilderOpen] = useState(false);
  const [isAttachOpen, setIsAttachOpen] = useState(false);
  const [detailRule, setDetailRule] = useState<Automation | null>(null);
  const [editingRule, setEditingRule] = useState<Automation | null>(null);
  const [catalogFilter, setCatalogFilter] = useState<'all' | 'global' | 'this_board' | 'other_columns'>('all');

  const { data: automations = [] } = useAutomationsQuery(workspaceId);
  const toggleMutation = useToggleAutomationMutation(workspaceId);
  const deleteMutation = useDeleteAutomationMutation(workspaceId);
  const updateMutation = useUpdateAutomationMutation(workspaceId);
  const createMutation = useCreateAutomationMutation(workspaceId);

  const { data: projects = [] } = useProjectsQuery(workspaceId);
  const projectBoardsQueries = useQueries({
    queries: projects.map((p) => ({
      queryKey: ['boards', workspaceId, p.id],
      queryFn: async (): Promise<Board[]> => {
        if (!workspaceId || !getAuthToken()) return DEMO_BOARDS.filter((b) => b.projectId === p.id);
        try { return (await projectApi.listProjectBoards(workspaceId, p.id)) ?? []; } catch { return []; }
      },
      enabled: !!workspaceId && !!p.id,
    })),
  });

  const rawBoards = useMemo(() => {
    const fetched = projectBoardsQueries.flatMap((q) => q.data ?? []);
    return fetched.length > 0 ? fetched : DEMO_BOARDS;
  }, [projectBoardsQueries]);

  const computedAllBoards = useMemo(() => {
    if (propBoards && propBoards.length > 0) return propBoards;
    return rawBoards.map((b) => {
      const proj = projects.find((p) => p.id === (b as { projectId?: string }).projectId);
      return { ...b, projectName: proj?.name };
    });
  }, [propBoards, rawBoards, projects]);

  const { data: fetchedBoardGroups = [] } = useProjectAllColumnsQuery(workspaceId, rawBoards);

  const computedBoardColumnGroups = useMemo(() => {
    if (propGroups && propGroups.length > 0) return propGroups;
    if (fetchedBoardGroups.length > 0) {
      return fetchedBoardGroups
        .map((bg) => {
          const proj = projects.find((p) => p.id === (bg.board as { projectId?: string }).projectId);
          return { boardName: bg.boardName, projectName: proj?.name, columns: bg.columns };
        })
        .filter((g) => g.columns.length > 0);
    }
    return computedAllBoards
      .map((b) => ({ boardName: b.name, projectName: b.projectName, columns: allColumns }))
      .filter((g) => g.columns.length > 0);
  }, [propGroups, fetchedBoardGroups, computedAllBoards, projects, allColumns]);

  const columnAutomations = automations.filter((a) => a.columnId === column.id);
  const availableAutomations = automations.filter((a) => a.columnId !== column.id);

  const filteredAvailableRules = useMemo(() => {
    if (catalogFilter === 'global') return availableAutomations.filter((a) => !a.boardId && !a.columnId);
    if (catalogFilter === 'this_board') return availableAutomations.filter((a) => a.boardId === column.boardId);
    if (catalogFilter === 'other_columns') return availableAutomations.filter((a) => Boolean(a.columnId && a.columnId !== column.id));
    return availableAutomations;
  }, [availableAutomations, catalogFilter, column.boardId, column.id]);

  const groupedAvailableRules = useMemo(() => {
    const groups = new Map<string, Automation[]>();
    for (const rule of filteredAvailableRules) {
      let groupKey = 'Глобальные правила';
      if (rule.boardId || rule.columnId) {
        const effBoardId = rule.boardId || allColumns.find((c) => c.id === rule.columnId)?.boardId;
        const b = computedAllBoards.find((board) => board.id === effBoardId);
        groupKey = b ? (b.projectName ? `${b.projectName} → ${b.name}` : b.name) : (effBoardId === column.boardId ? 'Текущая доска' : 'Другие доски');
      }
      if (!groups.has(groupKey)) groups.set(groupKey, []);
      groups.get(groupKey)!.push(rule);
    }
    return Array.from(groups.entries()).map(([groupTitle, rules]) => ({ groupTitle, rules }));
  }, [filteredAvailableRules, computedAllBoards, allColumns, column.boardId]);

  useEffect(() => {
    if (isOpen) {
      setConfig(getStoredColumnAutomationConfig(column.id, defaultTargetId));
      setIsAttachOpen(false);
      setCatalogFilter('all');
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

  const handleBind = (a: Automation) => {
    const isColumnScoped = Boolean(a.columnId);
    const isDifferentBoard = Boolean(a.boardId && a.boardId !== column.boardId);

    if (isColumnScoped || isDifferentBoard) {
      createMutation.mutate({
        name: a.name,
        triggerType: a.triggerType,
        boardId: column.boardId || a.boardId || null,
        columnId: column.id,
        conditionsJson: a.conditionsJson,
        actionsJson: a.actionsJson,
        isEnabled: a.isEnabled,
        ruleDefinition: a.ruleDefinition,
      });
    } else {
      updateMutation.mutate({
        automationId: a.id,
        data: { columnId: column.id, boardId: column.boardId || a.boardId || null },
      });
    }
    setIsAttachOpen(false);
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
                {PRIORITIES.map((p) => {
                  const isSelected = config.enterPriority === p.value || (p.value === 'pending' && config.enterPriority === 'urgent');
                  return (
                    <button
                      key={p.value}
                      type="button"
                      onClick={() => handleUpdate({ enterPriority: p.value })}
                      className={cn(
                        'text-xs px-2.5 py-0.5 rounded-full border transition-all',
                        isSelected
                          ? `${PRIORITY_ACTIVE_STYLES[p.value]} font-medium shadow-washi-sm`
                          : 'border-subtle/60 text-muted hover:text-primary hover:bg-canvas/50'
                      )}
                    >
                      {p.label}
                    </button>
                  );
                })}
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
                <div className="flex items-center justify-between gap-1 flex-wrap">
                  <span className="text-[10px] font-mono uppercase text-muted">Каталог правил:</span>
                  <div className="flex items-center gap-1 flex-wrap">
                    {(
                      [
                        { id: 'all', label: 'Все' },
                        { id: 'global', label: 'Глобальные' },
                        { id: 'this_board', label: 'На этой доске' },
                        { id: 'other_columns', label: 'Из других колонок' },
                      ] as const
                    ).map((chip) => (
                      <button
                        key={chip.id}
                        type="button"
                        onClick={() => setCatalogFilter(chip.id)}
                        className={cn(
                          'px-2 py-0.5 rounded-full text-[10px] font-medium transition-colors border',
                          catalogFilter === chip.id
                            ? 'bg-accent-solid text-white border-accent-solid shadow-washi-sm'
                            : 'bg-surface border-subtle text-muted hover:text-primary hover:bg-canvas'
                        )}
                      >
                        {chip.label}
                      </button>
                    ))}
                  </div>
                </div>

                {groupedAvailableRules.length === 0 ? (
                  <p className="text-xs text-muted italic py-1">Нет правил в выбранной категории.</p>
                ) : (
                  <div className="space-y-2 max-h-48 overflow-y-auto pr-0.5">
                    {groupedAvailableRules.map(({ groupTitle, rules }) => (
                      <div key={groupTitle} className="space-y-1">
                        <div className="text-[10px] font-mono uppercase tracking-wider text-muted px-1">{groupTitle}</div>
                        {rules.map((a) => (
                          <div key={a.id} className="flex items-center justify-between gap-2 p-2 rounded-lg border border-subtle bg-canvas/40 text-xs hover:bg-canvas/70 transition-colors">
                            <div className="min-w-0 cursor-pointer flex-1" onClick={() => setDetailRule(a)}>
                              <div className="font-medium text-primary truncate hover:underline">{a.name}</div>
                              <div className="text-[10px] text-muted truncate">
                                {formatTriggerLabel(a.triggerType, a.columnId, allColumns, a.ruleDefinition?.trigger?.params)}
                              </div>
                            </div>
                            <div className="flex items-center gap-1.5 shrink-0">
                              <Button variant="ghost" size="sm" onClick={() => setDetailRule(a)} className="text-[10px] h-6 px-1.5 text-muted hover:text-primary">
                                Подробнее
                              </Button>
                              <Button variant="solid" size="sm" onClick={() => handleBind(a)} className="text-[10px] h-6 px-2 shrink-0">
                                Привязать
                              </Button>
                            </div>
                          </div>
                        ))}
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
                      <Button variant="ghost" size="sm" onClick={() => setDetailRule(rule)} className="text-[10px] h-6 px-1.5 text-muted hover:text-primary">Подробнее</Button>
                      <button type="button" onClick={() => handleDuplicate(rule)} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Создать копию"><Copy className="w-3.5 h-3.5" /></button>
                      <button type="button" onClick={() => handleEdit(rule)} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Редактировать правило"><Pencil className="w-3.5 h-3.5" /></button>
                      <button type="button" onClick={() => updateMutation.mutate({ automationId: rule.id, data: { columnId: null } })} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Отвязать"><Link2Off className="w-3.5 h-3.5" /></button>
                      <button type="button" onClick={() => deleteMutation.mutate(rule.id)} className="p-1 rounded text-muted hover:text-primary hover:bg-surface/80" title="Удалить"><Trash2 className="w-3.5 h-3.5" /></button>
                      <button
                        type="button" role="switch" aria-checked={rule.isEnabled}
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
        prefilledBoardId={column.boardId}
        prefilledColumnId={column.id}
        prefilledColumnName={column.name}
        allColumns={allColumns}
        allBoards={computedAllBoards}
        boardColumnGroups={computedBoardColumnGroups}
        mode={editingRule ? 'edit' : 'create'}
        initialRule={editingRule}
      />

      {/* RuleDetailModal */}
      <RuleDetailModal
        isOpen={Boolean(detailRule)}
        onClose={() => setDetailRule(null)}
        rule={detailRule}
        allColumns={allColumns}
        allBoards={computedAllBoards}
        boardColumnGroups={computedBoardColumnGroups}
        onEdit={(r) => { setDetailRule(null); handleEdit(r); }}
        onDuplicate={handleDuplicate}
        onBind={detailRule && detailRule.columnId !== column.id ? (r) => { handleBind(r); setDetailRule(null); } : undefined}
      />
    </>
  );
};
