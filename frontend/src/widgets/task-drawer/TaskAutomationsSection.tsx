import React, { useState } from 'react';
import { Zap, Plus, Trash2, CheckCircle2, Play } from 'lucide-react';
import { Button, Badge } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import type { Automation } from '@/entities/automation';
import {
  useTaskAutomationsQuery,
  useToggleAutomationMutation,
  useDeleteAutomationMutation,
} from '@/entities/automation';
import { RuleBuilderModal, formatTriggerLabel } from '@/features/automations';

export interface TaskAutomationsSectionProps {
  taskId: string;
  columnId?: string | null;
  boardId?: string | null;
  workspaceId?: string;
  projectId?: string | null;
  allColumns?: { id: string; name: string }[];
  boardColumnGroups?: { boardName: string; columns: { id: string; name: string }[] }[];
  onApplyAction?: (rule: Automation) => void;
  className?: string;
}

export const TaskAutomationsSection: React.FC<TaskAutomationsSectionProps> = ({
  taskId,
  columnId,
  boardId,
  workspaceId,
  allColumns = [],
  boardColumnGroups,
  onApplyAction,
  className,
}) => {
  const { data: automations = [] } = useTaskAutomationsQuery(workspaceId, taskId);
  const toggleMutation = useToggleAutomationMutation(workspaceId);
  const deleteMutation = useDeleteAutomationMutation(workspaceId);

  const [isBuilderOpen, setIsBuilderOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<Automation | null>(null);
  const [appliedRuleId, setAppliedRuleId] = useState<string | null>(null);

  const handleApply = (rule: Automation) => {
    if (onApplyAction) {
      onApplyAction(rule);
    }
    setAppliedRuleId(rule.id);
    setTimeout(() => setAppliedRuleId(null), 2500);
  };

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between">
        <label className="text-[10px] font-mono uppercase tracking-wider text-muted flex items-center gap-1.5">
          <Zap className="w-3 h-3 text-accent-solid" />
          <span>Автоматизации задачи</span>
          {automations.length > 0 && (
            <Badge variant="subtle" size="sm" className="font-mono text-[10px] px-1.5 py-0">
              {automations.length}
            </Badge>
          )}
        </label>
        <button
          type="button"
          onClick={() => {
            setEditingRule(null);
            setIsBuilderOpen(true);
          }}
          className="text-xs text-muted hover:text-primary transition-colors flex items-center gap-1 p-0.5 rounded hover:bg-canvas/60"
          title="Добавить правило для задачи"
        >
          <Plus className="w-3 h-3 text-accent-solid" />
          <span className="text-[11px] font-medium">Добавить</span>
        </button>
      </div>

      {automations.length > 0 ? (
        <div className="space-y-1.5">
          {automations.map((rule) => {
            const isApplied = appliedRuleId === rule.id;
            return (
              <div
                key={rule.id}
                className="flex items-center justify-between gap-2 p-2 rounded-xl bg-surface/80 border border-subtle shadow-washi-sm transition-all"
              >
                <div className="min-w-0 flex-1 space-y-0.5">
                  <div className="flex items-center gap-1.5 flex-wrap">
                    <span className="text-xs font-medium text-primary truncate max-w-[170px]">
                      {rule.name}
                    </span>
                    <Badge variant="subtle" size="sm" className="text-[9px] px-1 py-0 font-mono">
                      {formatTriggerLabel(rule.triggerType)}
                    </Badge>
                  </div>
                  {rule.ruleDefinition?.actions && rule.ruleDefinition.actions.length > 0 && (
                    <p className="text-[10px] text-muted truncate">
                      Действий: {rule.ruleDefinition.actions.length}
                    </p>
                  )}
                </div>

                <div className="flex items-center gap-1.5 shrink-0">
                  {/* Manual Run / Apply button (D-072-4) */}
                  <button
                    type="button"
                    onClick={() => handleApply(rule)}
                    className={cn(
                      'text-[10px] px-2 py-0.5 rounded-md border transition-all flex items-center gap-1',
                      isApplied
                        ? 'bg-accent-solid text-white border-accent-solid font-medium'
                        : 'border-subtle bg-canvas/40 text-muted hover:text-primary hover:bg-canvas'
                    )}
                    title="Применить действия правила сейчас"
                  >
                    {isApplied ? (
                      <>
                        <CheckCircle2 className="w-2.5 h-2.5" />
                        <span>Готово</span>
                      </>
                    ) : (
                      <>
                        <Play className="w-2.5 h-2.5 text-accent-solid fill-accent-solid" />
                        <span>Применить</span>
                      </>
                    )}
                  </button>

                  {/* Toggle On/Off */}
                  <button
                    type="button"
                    role="switch"
                    aria-checked={rule.isEnabled}
                    onClick={() =>
                      toggleMutation.mutate({ automationId: rule.id, isEnabled: !rule.isEnabled })
                    }
                    className={cn(
                      'w-7 h-4 flex items-center rounded-full p-0.5 transition-colors',
                      rule.isEnabled ? 'bg-accent-solid justify-end' : 'bg-canvas border border-subtle justify-start'
                    )}
                    title={rule.isEnabled ? 'Отключить правило' : 'Включить правило'}
                  >
                    <span className="w-3 h-3 rounded-full bg-white shadow-xs" />
                  </button>

                  {/* Delete Button */}
                  <button
                    type="button"
                    onClick={() => deleteMutation.mutate(rule.id)}
                    className="p-1 text-muted hover:text-primary rounded hover:bg-canvas transition-colors"
                    title="Удалить правило"
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="p-2.5 rounded-xl border border-dashed border-subtle/80 bg-canvas/30 text-center">
          <p className="text-[11px] text-muted">Нет персональных правил для задачи</p>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => {
              setEditingRule(null);
              setIsBuilderOpen(true);
            }}
            className="mt-1 h-6 text-[10px] text-accent-solid"
          >
            + Создать правило (таймер, перенос)
          </Button>
        </div>
      )}

      {/* Rule Builder Modal */}
      {isBuilderOpen && (
        <RuleBuilderModal
          isOpen={isBuilderOpen}
          onClose={() => {
            setIsBuilderOpen(false);
            setEditingRule(null);
          }}
          workspaceId={workspaceId}
          taskId={taskId}
          prefilledBoardId={boardId ?? undefined}
          prefilledColumnId={columnId ?? undefined}
          allColumns={allColumns}
          boardColumnGroups={boardColumnGroups}
          mode={editingRule ? 'edit' : 'create'}
          initialRule={editingRule}
        />
      )}
    </div>
  );
};
