import React from 'react';
import { Zap, Clock, Sparkles, ArrowRight, Trash2, Pencil, Copy } from 'lucide-react';
import { Badge, Button } from '@/shared/components/ui';
import type { Automation } from '@/entities/automation';
import { cn } from '@/shared/lib/cn';
import {
  formatTriggerLabel,
  formatConditionLabel,
  formatActionLabel,
  formatAutomationScope,
} from './useRuleBuilder';

export interface AutomationCardProps {
  automation: Automation;
  index?: number;
  onToggle: (automation: Automation) => void;
  onDelete?: (automationId: string) => void;
  onDetail?: (automation: Automation) => void;
  onEdit?: (automation: Automation) => void;
  onDuplicate?: (automation: Automation) => void;
  columnName?: string;
  allColumns?: { id: string; name: string; boardId?: string }[];
  allBoards?: { id: string; name: string; projectId?: string; projectName?: string }[];
  boardColumnGroups?: { boardName: string; projectName?: string; columns: { id: string; name: string }[] }[];
  isDeletable?: boolean;
}

export const AutomationCard: React.FC<AutomationCardProps> = ({
  automation,
  index,
  onToggle,
  onDelete,
  onDetail,
  onEdit,
  onDuplicate,
  columnName,
  allColumns = [],
  allBoards = [],
  boardColumnGroups,
  isDeletable = true,
}) => {
  const triggerType = automation.triggerType;
  const isTimeTrigger = triggerType === 'column_sla_timer' || triggerType === 'scheduled_sweep';

  const triggerLabel = formatTriggerLabel(
    triggerType,
    automation.columnId,
    allColumns.length > 0 ? allColumns : columnName ? [{ id: automation.columnId || '', name: columnName }] : [],
    automation.ruleDefinition?.trigger?.params
  );

  const scopeBreadcrumb = formatAutomationScope(automation, {
    allColumns,
    allBoards,
    boardColumnGroups,
    columnName,
  });

  const conditions = automation.ruleDefinition?.conditions ?? [];
  const actions = automation.ruleDefinition?.actions ?? [];

  return (
    <div
      className={cn(
        'p-3.5 rounded-xl border transition-all duration-150 bg-surface/70',
        automation.isEnabled
          ? 'border-subtle shadow-washi-sm hover:border-subtle/90'
          : 'border-subtle/50 opacity-70'
      )}
    >
      <div className="flex items-start justify-between gap-3">
        {/* Left icon & details */}
        <div className="flex items-start gap-3 min-w-0 flex-1">
          <div className="w-8 h-8 rounded-lg bg-canvas border border-subtle flex items-center justify-center shrink-0 shadow-washi-sm mt-0.5">
            {isTimeTrigger ? (
              <Clock className="w-4 h-4 text-accent-solid" />
            ) : triggerType === 'task_completed' || triggerType === 'task_completed_in_column' ? (
              <Sparkles className="w-4 h-4 text-accent-solid" />
            ) : (
              <Zap className="w-4 h-4 text-accent-solid" />
            )}
          </div>

          <div className="space-y-1.5 min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap">
              {typeof index === 'number' && (
                <Badge variant="subtle" size="sm" className="font-mono text-[10px]">
                  #A-{(index + 1).toString().padStart(2, '0')}
                </Badge>
              )}
              <h4
                className={cn(
                  'font-serif text-sm font-normal text-primary',
                  onDetail && 'cursor-pointer hover:underline'
                )}
                onClick={() => onDetail && onDetail(automation)}
              >
                {automation.name}
              </h4>
              <Badge
                variant="subtle"
                size="sm"
                className="text-[10px] font-mono text-muted max-w-[240px] truncate"
                title={scopeBreadcrumb}
              >
                {scopeBreadcrumb}
              </Badge>
              <Badge
                variant={automation.isEnabled ? 'matcha' : 'subtle'}
                size="sm"
                dot
              >
                {automation.isEnabled ? 'Вкл' : 'Выкл'}
              </Badge>
            </div>

            {/* Sentence Builder Visual Chain */}
            <div className="flex items-center gap-1.5 flex-wrap text-[11px] font-mono">
              <span className="text-muted text-[10px] uppercase">КОГДА</span>
              <span className="bg-canvas px-1.5 py-0.5 rounded border border-subtle text-primary">
                {triggerLabel}
              </span>

              {conditions.length > 0 && (
                <>
                  <span className="text-muted text-[10px] uppercase">
                    ЕСЛИ {automation.ruleDefinition?.logicalOperator === 'or' ? '(ИЛИ)' : ''}
                  </span>
                  {conditions.map((c, i) => (
                    <span
                      key={i}
                      className="bg-accent-surface/30 px-1.5 py-0.5 rounded border border-subtle text-primary"
                    >
                      {formatConditionLabel(c)}
                    </span>
                  ))}
                </>
              )}

              {actions.length > 0 && (
                <>
                  <ArrowRight className="w-3 h-3 text-muted shrink-0" />
                  <span className="text-muted text-[10px] uppercase">ТО</span>
                  {actions.map((a, i) => (
                    <span
                      key={i}
                      className="bg-canvas px-1.5 py-0.5 rounded border border-subtle text-accent-solid font-medium"
                    >
                      {formatActionLabel(a, allColumns)}
                    </span>
                  ))}
                </>
              )}
            </div>
          </div>
        </div>

        {/* Right actions: Detail, Duplicate, Edit, Delete and Toggle */}
        <div className="flex items-center gap-1.5 shrink-0">
          {onDetail && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onDetail(automation)}
              className="text-[11px] h-7 px-2 text-muted hover:text-primary"
            >
              Подробнее
            </Button>
          )}

          {onDuplicate && (
            <button
              type="button"
              onClick={() => onDuplicate(automation)}
              className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
              title="Создать копию"
            >
              <Copy className="w-3.5 h-3.5" />
            </button>
          )}

          {onEdit && (
            <button
              type="button"
              onClick={() => onEdit(automation)}
              className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
              title="Редактировать правило"
            >
              <Pencil className="w-3.5 h-3.5" />
            </button>
          )}

          {isDeletable && onDelete && (
            <button
              type="button"
              onClick={() => onDelete(automation.id)}
              className="p-1 rounded text-muted hover:text-primary hover:bg-canvas transition-colors"
              title="Удалить правило"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}

          <button
            type="button"
            role="switch"
            aria-checked={automation.isEnabled}
            onClick={() => onToggle(automation)}
            className={cn(
              'w-10 h-5 flex items-center rounded-full p-0.5 transition-colors duration-200 cursor-pointer ml-1',
              automation.isEnabled ? 'bg-accent-solid' : 'bg-canvas border border-subtle'
            )}
            title={automation.isEnabled ? 'Выключить' : 'Включить'}
          >
            <span
              className={cn(
                'w-4 h-4 rounded-full bg-surface shadow-washi-sm transition-transform duration-200',
                automation.isEnabled ? 'translate-x-5' : 'translate-x-0'
              )}
            />
          </button>
        </div>
      </div>
    </div>
  );
};

